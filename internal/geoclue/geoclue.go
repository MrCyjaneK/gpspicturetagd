package geoclue

import (
	"context"
	"fmt"
	"log"
	"math"
	"time"

	"github.com/godbus/dbus/v5"
)

const (
	fieldLatitude  int32 = 1
	fieldLongitude int32 = 2
	fieldAltitude  int32 = 4
)

type Accuracy struct {
	Level      int32
	Horizontal float64
	Vertical   float64
}

// Fix is any usable lock: lat+lon present. Better GPS may never arrive.
type Fix struct {
	Lat, Lon, Alt float64
	HasAlt        bool
	Time          time.Time
	Level         int32
	Horizontal    float64
	Provider      string
}

type provider struct {
	dest, path, name string
}

var providers = []provider{
	{"org.freedesktop.Geoclue.Providers.Mlsdb", "/org/freedesktop/Geoclue/Providers/Mlsdb", "MLSDB"},
	{"org.freedesktop.Geoclue.Providers.Hybris", "/org/freedesktop/Geoclue/Providers/Hybris", "Hybris"},
}

type ref struct {
	obj  dbus.BusObject
	name string
}

type Client struct {
	conn *dbus.Conn
	refs []ref
}

func Open() (*Client, error) {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return nil, fmt.Errorf("session bus: %w", err)
	}
	c := &Client{conn: conn}
	for _, p := range providers {
		obj := conn.Object(p.dest, dbus.ObjectPath(p.path))
		if err := obj.Call("org.freedesktop.Geoclue.AddReference", 0).Err; err != nil {
			log.Printf("geoclue %s AddReference: %v", p.name, err)
			continue
		}
		if err := conn.AddMatchSignal(
			dbus.WithMatchSender(p.dest),
			dbus.WithMatchObjectPath(dbus.ObjectPath(p.path)),
			dbus.WithMatchInterface("org.freedesktop.Geoclue.Position"),
			dbus.WithMatchMember("PositionChanged"),
		); err != nil {
			c.Close()
			return nil, err
		}
		c.refs = append(c.refs, ref{obj: obj, name: p.name})
		log.Printf("geoclue listening %s", p.name)
	}
	if len(c.refs) == 0 {
		conn.Close()
		return nil, fmt.Errorf("no geoclue providers")
	}
	return c, nil
}

func gpsEnough(f Fix) bool {
	return f.Provider == "Hybris"
}

func acquireErr(ctx context.Context, timeout time.Duration, err error) error {
	if ctx.Err() == context.DeadlineExceeded {
		return fmt.Errorf("no lock within %s", timeout)
	}
	if err == nil {
		err = ctx.Err()
	}
	if err == nil {
		err = fmt.Errorf("no lock")
	}
	return err
}

// Acquire turns GPS on, waits for a Hybris lock (or the best lock before
// timeout), then drops GeoClue references so the radio can sleep.
func Acquire(ctx context.Context, timeout time.Duration) (Fix, error) {
	if timeout <= 0 {
		timeout = 15 * time.Minute
	}
	client, err := Open()
	if err != nil {
		return Fix{}, err
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	out := make(chan Fix, 8)
	errc := make(chan error, 1)
	go func() {
		errc <- client.Run(ctx, out)
	}()

	var (
		best Fix
		got  bool
	)
	for {
		select {
		case f := <-out:
			if !got || Better(f, best) {
				best, got = f, true
				log.Printf("gps %s", f)
			}
			if gpsEnough(best) {
				return best, nil
			}
		case err := <-errc:
			if got {
				return best, nil
			}
			return Fix{}, acquireErr(ctx, timeout, err)
		case <-ctx.Done():
			if got {
				return best, nil
			}
			return Fix{}, acquireErr(ctx, timeout, ctx.Err())
		}
	}
}

func (c *Client) Close() {
	if c == nil {
		return
	}
	for _, r := range c.refs {
		_ = r.obj.Call("org.freedesktop.Geoclue.RemoveReference", 0).Err
	}
	if c.conn != nil {
		c.conn.Close()
	}
}

func parseFix(fields int32, ts int32, lat, lon, alt float64, acc Accuracy, provider string) (Fix, bool) {
	if fields&fieldLatitude == 0 || fields&fieldLongitude == 0 {
		return Fix{}, false
	}
	if math.Abs(lat) > 90 || math.Abs(lon) > 180 {
		return Fix{}, false
	}
	t := time.Unix(int64(ts), 0).UTC()
	if ts == 0 {
		t = time.Now().UTC()
	}
	f := Fix{
		Lat: lat, Lon: lon, Time: t,
		Level: acc.Level, Horizontal: acc.Horizontal, Provider: provider,
	}
	if fields&fieldAltitude != 0 && !math.IsNaN(alt) {
		f.Alt = alt
		f.HasAlt = true
	}
	return f, true
}

func getPosition(obj dbus.BusObject, name string) (Fix, bool, error) {
	var fields, ts int32
	var lat, lon, alt float64
	var acc Accuracy
	err := obj.Call("org.freedesktop.Geoclue.Position.GetPosition", 0).
		Store(&fields, &ts, &lat, &lon, &alt, &acc)
	if err != nil {
		return Fix{}, false, err
	}
	f, ok := parseFix(fields, ts, lat, lon, alt, acc, name)
	return f, ok, nil
}

func Better(a, b Fix) bool {
	if a.Level != b.Level {
		return a.Level > b.Level
	}
	ah, bh := a.Horizontal, b.Horizontal
	if ah <= 0 {
		ah = 1e9
	}
	if bh <= 0 {
		bh = 1e9
	}
	return ah < bh
}

func (f Fix) String() string {
	s := fmt.Sprintf("lat=%.6f lon=%.6f acc=%.0fm level=%d src=%s", f.Lat, f.Lon, f.Horizontal, f.Level, f.Provider)
	if f.HasAlt {
		s += fmt.Sprintf(" alt=%.1f", f.Alt)
	}
	return s
}

func nameOf(path dbus.ObjectPath) string {
	for _, p := range providers {
		if dbus.ObjectPath(p.path) == path {
			return p.name
		}
	}
	return string(path)
}

// Run holds provider references and sends the first usable lock, then updates.
func (c *Client) Run(ctx context.Context, out chan<- Fix) error {
	sigCh := make(chan *dbus.Signal, 16)
	c.conn.Signal(sigCh)

	emit := func(f Fix) bool {
		select {
		case out <- f:
			return true
		case <-ctx.Done():
			return false
		}
	}

	for _, r := range c.refs {
		if f, ok, err := getPosition(r.obj, r.name); err != nil {
			log.Printf("geoclue %s GetPosition: %v", r.name, err)
		} else if ok {
			if !emit(f) {
				return ctx.Err()
			}
		}
	}

	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	var lastWait time.Time

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case s, ok := <-sigCh:
			if !ok {
				return fmt.Errorf("dbus closed")
			}
			if s.Name != "org.freedesktop.Geoclue.Position.PositionChanged" {
				continue
			}
			var fields, ts int32
			var lat, lon, alt float64
			var acc Accuracy
			if err := dbus.Store(s.Body, &fields, &ts, &lat, &lon, &alt, &acc); err != nil {
				continue
			}
			f, ok := parseFix(fields, ts, lat, lon, alt, acc, nameOf(s.Path))
			if ok && !emit(f) {
				return ctx.Err()
			}
		case <-ticker.C:
			var (
				best Fix
				got  bool
			)
			for _, r := range c.refs {
				f, ok, err := getPosition(r.obj, r.name)
				if err != nil {
					log.Printf("geoclue %s GetPosition: %v", r.name, err)
					continue
				}
				if !ok {
					continue
				}
				if !got || Better(f, best) {
					best, got = f, true
				}
			}
			if got {
				if !emit(best) {
					return ctx.Err()
				}
			} else if time.Since(lastWait) > 30*time.Second {
				lastWait = time.Now()
				log.Printf("geoclue: waiting for lat/lon")
			}
		}
	}
}

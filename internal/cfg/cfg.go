package cfg

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"gopkg.in/yaml.v3"

	"gpspicturetagd/config"
)

type NotifyMode string

const (
	NotifyNever  NotifyMode = "never"
	NotifyAlways NotifyMode = "always"
	NotifyError  NotifyMode = "error"
	NotifyChange NotifyMode = "change"
)

type Config struct {
	Dir        string
	SkipDirs   []string
	MaxAge     time.Duration
	ScanEvery  time.Duration
	GPSTimeout time.Duration
	Notify     NotifyMode
	LoadedFrom string
}

func Default() Config {
	dir := xdg.UserDirs.Pictures
	if dir == "" {
		dir = "~/Pictures"
	}
	return Config{
		Dir:        dir,
		SkipDirs:   []string{"Default", "Screenshots"},
		MaxAge:     2 * time.Hour,
		ScanEvery:  5 * time.Minute,
		GPSTimeout: 15 * time.Minute,
		Notify:     NotifyChange,
	}
}

func ParseNotify(s string) NotifyMode {
	s = strings.ToLower(strings.TrimSpace(s))
	s = strings.ReplaceAll(s, "-", "_")
	s = strings.ReplaceAll(s, " ", "_")
	switch s {
	case "never", "off", "none", "no":
		return NotifyNever
	case "always", "all", "on":
		return NotifyAlways
	case "error", "errors", "on_error":
		return NotifyError
	case "change", "changed", "on_change":
		return NotifyChange
	default:
		return ""
	}
}

// Want reports whether a finished pass should notify.
// failed is true for scan/GPS/write errors; tagged is how many files were written.
func (m NotifyMode) Want(tagged int, failed bool) bool {
	switch m {
	case NotifyAlways:
		return true
	case NotifyError:
		return failed
	case NotifyChange:
		return tagged > 0 || failed
	default:
		return false
	}
}

func ExpandPath(p string) string {
	p = strings.TrimSpace(p)
	if p != "~" && !strings.HasPrefix(p, "~/") {
		return p
	}
	if xdg.Home == "" {
		return p
	}
	if p == "~" {
		return xdg.Home
	}
	return filepath.Join(xdg.Home, p[2:])
}

func Load(explicit string) (Config, error) {
	c := Default()
	if explicit != "" {
		if err := apply(&c, explicit); err != nil {
			return c, err
		}
		c.LoadedFrom = explicit
		c.Dir = ExpandPath(c.Dir)
		return c, nil
	}
	wrote, writeErr := ensureUserConfig()
	var files []string
	for _, p := range searchPaths() {
		err := apply(&c, p)
		if err == nil {
			files = append(files, p)
			continue
		}
		if os.IsNotExist(err) {
			continue
		}
		return c, fmt.Errorf("%s: %w", p, err)
	}
	from := "defaults"
	if len(files) > 0 {
		from = "defaults + " + strings.Join(files, ", ")
	}
	if wrote != "" {
		from = "wrote " + wrote + "; " + from
	} else if writeErr != nil {
		from = fmt.Sprintf("%s (user config not created: %v)", from, writeErr)
	}
	c.LoadedFrom = from
	c.Dir = ExpandPath(c.Dir)
	return c, nil
}

type duration time.Duration

func (d *duration) UnmarshalYAML(n *yaml.Node) error {
	var s string
	if err := n.Decode(&s); err != nil {
		return err
	}
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	parsed, err := time.ParseDuration(s)
	if err != nil {
		return err
	}
	if parsed <= 0 {
		return fmt.Errorf("duration must be positive")
	}
	*d = duration(parsed)
	return nil
}

type fileConfig struct {
	Dir        *string   `yaml:"dir"`
	SkipDirs   []string  `yaml:"skip_dirs"`
	MaxAge     *duration `yaml:"max_age"`
	ScanEvery  *duration `yaml:"scan_every"`
	GPSTimeout *duration `yaml:"gps_timeout"`
	Notify     *string   `yaml:"notify"`
}

func apply(dst *Config, path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	var f fileConfig
	if err := yaml.Unmarshal(b, &f); err != nil {
		return err
	}
	if f.Dir != nil && strings.TrimSpace(*f.Dir) != "" {
		dst.Dir = strings.TrimSpace(*f.Dir)
	}
	if f.SkipDirs != nil {
		dst.SkipDirs = append([]string(nil), f.SkipDirs...)
	}
	if f.MaxAge != nil && *f.MaxAge > 0 {
		dst.MaxAge = time.Duration(*f.MaxAge)
	}
	if f.ScanEvery != nil && *f.ScanEvery > 0 {
		dst.ScanEvery = time.Duration(*f.ScanEvery)
	}
	if f.GPSTimeout != nil && *f.GPSTimeout > 0 {
		dst.GPSTimeout = time.Duration(*f.GPSTimeout)
	}
	if f.Notify != nil {
		if m := ParseNotify(*f.Notify); m != "" {
			dst.Notify = m
		}
	}
	return nil
}

func UserPath() string {
	if xdg.ConfigHome == "" {
		return ""
	}
	return filepath.Join(xdg.ConfigHome, "gpspicturetagd", "config.yaml")
}

func searchPaths() []string {
	out := make([]string, 0, len(xdg.DataDirs)+2)
	for _, dir := range xdg.DataDirs {
		out = append(out, filepath.Join(dir, "gpspicturetagd", "config.yaml"))
	}
	out = append(out, "/etc/gpspicturetagd/config.yaml")
	if p := UserPath(); p != "" {
		out = append(out, p)
	}
	return out
}

func ensureUserConfig() (string, error) {
	p := UserPath()
	if p == "" {
		return "", fmt.Errorf("no home/config dir")
	}
	if _, err := os.Stat(p); err == nil {
		return "", nil
	} else if !os.IsNotExist(err) {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return "", err
	}
	if err := os.WriteFile(p, config.DefaultYAML, 0o644); err != nil {
		return "", err
	}
	return p, nil
}

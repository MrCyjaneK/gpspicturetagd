package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"

	"github.com/adrg/xdg"

	"gpspicturetagd/internal/cfg"
	"gpspicturetagd/internal/geoclue"
	"gpspicturetagd/internal/notify"
	"gpspicturetagd/internal/seen"
	"gpspicturetagd/internal/tag"
	"gpspicturetagd/internal/userunit"
	"gpspicturetagd/internal/watch"
)

var version = "dev"

func main() {
	log.SetFlags(log.LstdFlags | log.Lmsgprefix)
	log.SetPrefix("gpspicturetagd: ")

	configPath := flag.String("config", "", "YAML config (default: search path after compiled defaults)")
	showVersion := flag.Bool("version", false, "print version and exit")
	install := flag.Bool("install", false, "install user systemd timer (scan_every) and enable it")
	uninstall := flag.Bool("uninstall", false, "disable and remove the user systemd timer")
	logPath := flag.String("log", "", "append logs to this file (also stderr)")
	flag.Parse()

	if *showVersion {
		log.Printf("%s", version)
		return
	}

	logFile := *logPath
	if logFile == "" {
		logFile = filepath.Join(xdg.CacheHome, "gpspicturetagd.log")
	}

	if *logPath != "" || *install || *uninstall {
		if err := os.MkdirAll(filepath.Dir(logFile), 0o755); err != nil {
			log.Fatalf("log dir: %v", err)
		}
		f, err := os.OpenFile(logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			log.Fatalf("log: %v", err)
		}
		_ = f.Chmod(0o600)
		defer f.Close()
		log.SetOutput(io.MultiWriter(os.Stderr, f))
	}

	conf, err := cfg.Load(*configPath)
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if *uninstall {
		if err := userunit.Uninstall(); err != nil {
			log.Fatalf("uninstall: %v", err)
		}
		log.Printf("removed user timer")
		return
	}
	if *install {
		bin, err := os.Executable()
		if err != nil {
			log.Fatalf("executable: %v", err)
		}
		if resolved, err := filepath.EvalSymlinks(bin); err == nil {
			bin = resolved
		}
		if err := userunit.Install(bin, logFile, conf.ScanEvery, conf.GPSTimeout); err != nil {
			log.Fatalf("install: %v", err)
		}
		log.Printf("installed user timer every %s exe=%s", conf.ScanEvery, bin)
		return
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	logConfig(conf)
	if err := os.MkdirAll(conf.Dir, 0o755); err != nil {
		log.Fatalf("dir: %v", err)
	}
	w := watch.New(conf.Dir, conf.SkipDirs)
	cache := seen.Open(filepath.Join(xdg.CacheHome, "gpspicturetagd.seen"))
	runPass(ctx, w, conf, cache)
}

func logConfig(conf cfg.Config) {
	log.Printf("%s config=%s dir=%s skip_dirs=%s max_age=%s scan_every=%s gps_timeout=%s notify=%s",
		version, conf.LoadedFrom, conf.Dir, strings.Join(conf.SkipDirs, ","),
		conf.MaxAge, conf.ScanEvery, conf.GPSTimeout, conf.Notify)
}

func runPass(ctx context.Context, w *watch.Watcher, conf cfg.Config, cache *seen.Cache) {
	existing, err := w.Scan()
	if err != nil {
		log.Printf("scan: %v", err)
		finishPass(ctx, conf, passResult{
			errMsg:  err.Error(),
			summary: "Scan failed",
			body:    err.Error(),
		})
		return
	}
	defer func() {
		if err := cache.Save(); err != nil {
			log.Printf("cache: %v", err)
		}
	}()

	var (
		pending                   []string
		skippedOld, skippedTagged int
	)
	for _, p := range existing {
		switch tag.Check(p, conf.MaxAge, cache) {
		case "too old":
			skippedOld++
		case "already tagged":
			skippedTagged++
		case "":
			pending = append(pending, p)
		}
	}
	log.Printf("scan %d JPEG(s): %d eligible, %d already tagged, %d too old (>%s)",
		len(existing), len(pending), skippedTagged, skippedOld, conf.MaxAge)
	if len(pending) == 0 {
		log.Printf("gps off")
		finishPass(ctx, conf, passResult{
			summary: "No photos to tag",
			body:    fmt.Sprintf("%d JPEG(s) · %d already tagged · %d too old", len(existing), skippedTagged, skippedOld),
		})
		return
	}

	log.Printf("gps on: %d file(s) need coordinates (timeout %s)", len(pending), conf.GPSTimeout)
	f, err := geoclue.Acquire(ctx, conf.GPSTimeout)
	if err != nil {
		log.Printf("gps: %v", err)
		finishPass(ctx, conf, passResult{
			errMsg:  err.Error(),
			summary: "No GPS lock",
			body:    fmt.Sprintf("%d photo(s) waiting · %s", len(pending), err),
		})
		return
	}
	log.Printf("gps lock %s", f)

	pos := tag.FromFix(f.Lat, f.Lon, f.Alt, f.HasAlt, f.Time)
	var n, tagErrs int
	why := map[string]int{}
	for _, p := range pending {
		ok, reason, err := tag.File(p, pos, conf.MaxAge, cache)
		if err != nil {
			log.Printf("tag %s: %v", p, err)
			tagErrs++
			continue
		}
		if ok {
			n++
			log.Printf("tagged %s %s", p, f)
			continue
		}
		if reason != "" {
			why[reason]++
			log.Printf("skip %s: %s", p, reason)
		}
	}
	if n > 0 {
		log.Printf("wrote GPS to %d file(s)", n)
	}
	for reason, c := range why {
		log.Printf("skip %s: %d file(s)", reason, c)
	}
	log.Printf("gps off")

	body := f.String()
	if tagErrs > 0 {
		body = fmt.Sprintf("%s · %d write error(s)", body, tagErrs)
	}
	summary := fmt.Sprintf("Tagged %d photo(s)", n)
	if n == 0 && tagErrs > 0 {
		summary = "Tagging failed"
	}
	res := passResult{tagged: n, summary: summary, body: body}
	if tagErrs > 0 {
		res.errMsg = "write error"
	}
	finishPass(ctx, conf, res)
}

type passResult struct {
	tagged        int
	errMsg        string
	summary, body string
}

func finishPass(ctx context.Context, conf cfg.Config, r passResult) {
	if ctx.Err() != nil {
		return
	}
	if !conf.Notify.Want(r.tagged, r.errMsg != "") {
		return
	}
	if err := notify.Send(r.summary, r.body); err != nil {
		log.Printf("notify: %v", err)
		return
	}
	log.Printf("notify %s: %s", conf.Notify, r.summary)
}

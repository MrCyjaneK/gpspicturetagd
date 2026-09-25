package cfg

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/adrg/xdg"
)

func TestParseNotify(t *testing.T) {
	tests := map[string]NotifyMode{
		"never":     NotifyNever,
		"off":       NotifyNever,
		"always":    NotifyAlways,
		"on":        NotifyAlways,
		"error":     NotifyError,
		"on-error":  NotifyError,
		"change":    NotifyChange,
		"on_change": NotifyChange,
		"nope":      "",
	}
	for in, want := range tests {
		if got := ParseNotify(in); got != want {
			t.Errorf("ParseNotify(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestNotifyWant(t *testing.T) {
	if !NotifyChange.Want(1, false) || !NotifyChange.Want(0, true) || NotifyChange.Want(0, false) {
		t.Fatal("NotifyChange")
	}
	if NotifyError.Want(3, false) || !NotifyError.Want(0, true) {
		t.Fatal("NotifyError")
	}
	if !NotifyAlways.Want(0, false) || NotifyNever.Want(1, true) {
		t.Fatal("always/never")
	}
}

func TestExpandPath(t *testing.T) {
	if got := ExpandPath("/abs"); got != "/abs" {
		t.Fatalf("abs: %q", got)
	}
	if xdg.Home == "" {
		t.Skip("no home")
	}
	if got := ExpandPath("~/Pictures"); got != filepath.Join(xdg.Home, "Pictures") {
		t.Fatalf("~: %q", got)
	}
}

func TestLoadExplicitYAML(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "c.yaml")
	body := []byte("dir: /tmp/pics\nmax_age: 30m\nnotify: never\nskip_dirs:\n  - X\n")
	if err := os.WriteFile(p, body, 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if c.Dir != "/tmp/pics" || c.MaxAge != 30*time.Minute || c.Notify != NotifyNever {
		t.Fatalf("%+v", c)
	}
	if len(c.SkipDirs) != 1 || c.SkipDirs[0] != "X" {
		t.Fatalf("skip: %v", c.SkipDirs)
	}
}

func TestLoadIgnoresLegacySkipTagged(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "c.yaml")
	if err := os.WriteFile(p, []byte("skip_tagged: false\ndir: /tmp/pics\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	c, err := Load(p)
	if err != nil {
		t.Fatal(err)
	}
	if c.Dir != "/tmp/pics" {
		t.Fatalf("%+v", c)
	}
}

func TestDefaultGPSTimeout(t *testing.T) {
	if Default().GPSTimeout != 15*time.Minute {
		t.Fatalf("GPSTimeout = %s", Default().GPSTimeout)
	}
}

func TestLoadRejectsBadDuration(t *testing.T) {
	dir := t.TempDir()
	p := filepath.Join(dir, "c.yaml")
	if err := os.WriteFile(p, []byte("max_age: bananas\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if _, err := Load(p); err == nil {
		t.Fatal("expected duration parse error")
	}
}

package userunit

import (
	"strings"
	"testing"
	"time"
)

func TestServiceBody(t *testing.T) {
	body := ServiceBody("/opt/gpspicturetagd", "/tmp/log", 2*time.Minute)
	for _, want := range []string{
		"[Unit]",
		"Description=GPS Picture Tag scan",
		"[Service]",
		"Type=oneshot",
		"ExecStart=/opt/gpspicturetagd -log /tmp/log",
		"TimeoutStartSec=240s",
		"Nice=10",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in:\n%s", want, body)
		}
	}
}

func TestServiceBodyQuotesSpaces(t *testing.T) {
	body := ServiceBody("/opt/my bin", "/tmp/my log", time.Minute)
	if !strings.Contains(body, `ExecStart="/opt/my bin" -log "/tmp/my log"`) {
		t.Fatalf("quoting:\n%s", body)
	}
}

func TestTimerBody(t *testing.T) {
	body := TimerBody(5 * time.Minute)
	for _, want := range []string{
		"[Timer]",
		"OnActiveSec=5s",
		"OnUnitActiveSec=300s",
		"Persistent=true",
		"Unit=gpspicturetagd.service",
		"[Install]",
		"WantedBy=timers.target",
	} {
		if !strings.Contains(body, want) {
			t.Errorf("missing %q in:\n%s", want, body)
		}
	}
}

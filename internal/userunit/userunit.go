package userunit

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/coreos/go-systemd/v22/unit"
	"github.com/google/renameio/v2"
)

const (
	ServiceName = "gpspicturetagd.service"
	TimerName   = "gpspicturetagd.timer"
)

func Dir() string {
	return filepath.Join(xdg.ConfigHome, "systemd", "user")
}

func systemdSec(d time.Duration) string {
	sec := int(d.Round(time.Second).Seconds())
	if sec < 1 {
		sec = 1
	}
	return fmt.Sprintf("%ds", sec)
}

func quote(s string) string {
	if s == "" || strings.ContainsAny(s, " \t\"'") {
		return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
	}
	return s
}

func encodeUnit(opts []*unit.UnitOption) string {
	b, _ := io.ReadAll(unit.Serialize(opts))
	return string(b)
}

func ServiceBody(bin, logPath string, gpsTimeout time.Duration) string {
	timeout := gpsTimeout + 2*time.Minute
	if timeout < 3*time.Minute {
		timeout = 3 * time.Minute
	}
	return encodeUnit([]*unit.UnitOption{
		unit.NewUnitOption("Unit", "Description", "GPS Picture Tag scan"),
		unit.NewUnitOption("Service", "Type", "oneshot"),
		unit.NewUnitOption("Service", "ExecStart", quote(bin)+" -log "+quote(logPath)),
		unit.NewUnitOption("Service", "TimeoutStartSec", systemdSec(timeout)),
		unit.NewUnitOption("Service", "Nice", "10"),
	})
}

func TimerBody(every time.Duration) string {
	return encodeUnit([]*unit.UnitOption{
		unit.NewUnitOption("Unit", "Description", "GPS Picture Tag periodic scan"),
		unit.NewUnitOption("Timer", "OnActiveSec", "5s"),
		unit.NewUnitOption("Timer", "OnUnitActiveSec", systemdSec(every)),
		unit.NewUnitOption("Timer", "Persistent", "true"),
		unit.NewUnitOption("Timer", "AccuracySec", "30s"),
		unit.NewUnitOption("Timer", "Unit", ServiceName),
		unit.NewUnitOption("Install", "WantedBy", "timers.target"),
	})
}

func Write(bin, logPath string, every, gpsTimeout time.Duration) (string, error) {
	dir := Dir()
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return "", err
	}
	if err := os.MkdirAll(filepath.Dir(logPath), 0o755); err != nil {
		return "", err
	}
	svc := filepath.Join(dir, ServiceName)
	tmr := filepath.Join(dir, TimerName)
	if err := renameio.WriteFile(svc, []byte(ServiceBody(bin, logPath, gpsTimeout)), 0o644); err != nil {
		return "", err
	}
	if err := renameio.WriteFile(tmr, []byte(TimerBody(every)), 0o644); err != nil {
		return "", err
	}
	return dir, nil
}

func systemctl(args ...string) error {
	cmd := exec.Command("systemctl", append([]string{"--user"}, args...)...)
	out, err := cmd.CombinedOutput()
	if err != nil {
		msg := strings.TrimSpace(string(out))
		if msg == "" {
			msg = err.Error()
		}
		return fmt.Errorf("systemctl --user %s: %s", strings.Join(args, " "), msg)
	}
	return nil
}

func Enable() error {
	_ = systemctl("disable", "--now", ServiceName)
	if err := systemctl("daemon-reload"); err != nil {
		return err
	}
	if err := systemctl("enable", TimerName); err != nil {
		return err
	}
	return systemctl("start", TimerName)
}

func Disable() error {
	_ = systemctl("disable", "--now", TimerName)
	_ = systemctl("stop", ServiceName)
	_ = systemctl("disable", "--now", ServiceName)
	_ = systemctl("daemon-reload")
	return nil
}

func Remove() error {
	dir := Dir()
	_ = os.Remove(filepath.Join(dir, ServiceName))
	_ = os.Remove(filepath.Join(dir, TimerName))
	_ = systemctl("daemon-reload")
	return nil
}

func Install(bin, logPath string, every, gpsTimeout time.Duration) error {
	if _, err := Write(bin, logPath, every, gpsTimeout); err != nil {
		return err
	}
	return Enable()
}

func Uninstall() error {
	_ = Disable()
	return Remove()
}

func TimerActive() bool {
	cmd := exec.Command("systemctl", "--user", "is-active", TimerName)
	return cmd.Run() == nil
}

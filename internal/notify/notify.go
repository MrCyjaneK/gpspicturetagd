package notify

import (
	"fmt"
	"strings"
	"time"

	fdnotify "github.com/esiqveland/notify"
	"github.com/godbus/dbus/v5"
)

const (
	appName = "GPS Picture Tag"
	appIcon = "gpspicturetagd"
)

// Send a Sailfish/Lipstick session notification. Best-effort: errors are returned
// so the caller can log them; a missing notification service must not stop tagging.
func Send(summary, body string) error {
	conn, err := dbus.ConnectSessionBus()
	if err != nil {
		return fmt.Errorf("session bus: %w", err)
	}
	defer conn.Close()

	summary = strings.TrimSpace(summary)
	body = strings.TrimSpace(body)
	if summary == "" {
		summary = appName
	}

	n := fdnotify.Notification{
		AppName:       appName,
		AppIcon:       appIcon,
		Summary:       summary,
		Body:          body,
		ExpireTimeout: 5 * time.Second,
	}
	n.SetUrgency(fdnotify.UrgencyNormal)
	n.AddHint(fdnotify.Hint{ID: "x-nemo-preview-summary", Variant: dbus.MakeVariant(summary)})
	n.AddHint(fdnotify.Hint{ID: "x-nemo-icon", Variant: dbus.MakeVariant(appIcon)})
	if body != "" {
		n.AddHint(fdnotify.Hint{ID: "x-nemo-preview-body", Variant: dbus.MakeVariant(body)})
	}

	_, err = fdnotify.SendNotification(conn, n)
	return err
}

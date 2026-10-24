//go:build linux

package notify

import (
	"context"
	"fmt"
	"os"
)

type linuxNotifier struct{}

func newPlatformNotifier() notifier {
	return linuxNotifier{}
}

func (linuxNotifier) send(ctx context.Context, message Message) error {
	if !commandExists("notify-send") {
		return fmt.Errorf("%w: notify-send was not found; install libnotify", ErrBackendUnavailable)
	}
	if os.Getenv("DBUS_SESSION_BUS_ADDRESS") == "" {
		return fmt.Errorf("%w: DBUS_SESSION_BUS_ADDRESS is not set", ErrBackendUnavailable)
	}
	return runCommand(ctx, "notify-send", "--app-name=desktop-notify", "--",
		message.Title, message.Text)
}

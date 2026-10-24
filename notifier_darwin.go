//go:build darwin

package notify

import (
	"context"
	"fmt"
)

type darwinNotifier struct{}

func newPlatformNotifier() notifier {
	return darwinNotifier{}
}

func (darwinNotifier) send(ctx context.Context, message Message) error {
	if !commandExists("osascript") {
		return fmt.Errorf("%w: osascript was not found", ErrBackendUnavailable)
	}
	return runCommand(ctx, "osascript", "-e",
		"on run argv\n display notification (item 1 of argv) with title (item 2 of argv)\nend run",
		"--", message.Text, message.Title)
}

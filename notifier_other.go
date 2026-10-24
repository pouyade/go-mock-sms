//go:build !darwin && !linux && !windows

package notify

import (
	"context"
)

type unsupportedNotifier struct{}

func newPlatformNotifier() notifier {
	return unsupportedNotifier{}
}

func (unsupportedNotifier) send(context.Context, Message) error {
	return ErrUnsupportedPlatform
}

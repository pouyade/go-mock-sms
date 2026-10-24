package notify

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"time"
)

type notifier interface {
	send(context.Context, Message) error
}

// Client sends desktop notifications. A Client is safe for concurrent use.
type Client struct {
	backend notifier
}

// New creates a Client using the current operating system's notification
// backend.
func New() *Client {
	return &Client{backend: newPlatformNotifier()}
}

// Send sends a notification using the package-level default client.
func Send(title, text string) error {
	return defaultClient.Send(Message{Title: title, Text: text})
}

// Send validates and sends a notification.
func (c *Client) Send(message Message) error {
	if c == nil || c.backend == nil {
		return ErrBackendUnavailable
	}
	if err := message.validate(); err != nil {
		return err
	}
	message = message.normalized()
	return c.backend.send(context.Background(), message)
}

var defaultClient = New()

const backendTimeout = 10 * time.Second

func runCommand(ctx context.Context, name string, args ...string) error {
	commandCtx, cancel := context.WithTimeout(ctx, backendTimeout)
	defer cancel()

	command := exec.CommandContext(commandCtx, name, args...)
	output, err := command.CombinedOutput()
	if commandCtx.Err() != nil {
		return fmt.Errorf("%w: %s timed out", ErrBackendUnavailable, name)
	}
	if err != nil {
		detail := strings.TrimSpace(string(output))
		if detail != "" {
			return fmt.Errorf("%w: %s", ErrBackendUnavailable, detail)
		}
		return fmt.Errorf("%w: %s failed", ErrBackendUnavailable, name)
	}
	return nil
}

func commandExists(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

package httpapi

import "log/slog"

// NewMux returns the notification API handler with its fixed routes.
func NewMux(sender Sender, defaultTitle string, logger *slog.Logger) *Handler {
	return New(sender, defaultTitle, logger)
}

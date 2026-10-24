package notify

import "errors"

var (
	// ErrEmptyText means that a message body was empty or whitespace-only.
	ErrEmptyText = errors.New("notification text is required")

	// ErrTextTooLong means that the message body exceeds MaxTextRunes.
	ErrTextTooLong = errors.New("notification text is too long")

	// ErrTitleTooLong means that the title exceeds MaxTitleRunes.
	ErrTitleTooLong = errors.New("notification title is too long")

	// ErrInvalidText means that the message contains unsupported control
	// characters.
	ErrInvalidText = errors.New("notification text contains unsupported control characters")

	// ErrInvalidTitle means that the title contains unsupported control
	// characters.
	ErrInvalidTitle = errors.New("notification title contains unsupported control characters")

	// ErrBackendUnavailable means that the operating-system backend could not
	// be started or is not available in the current session.
	ErrBackendUnavailable = errors.New("notification backend unavailable")

	// ErrUnsupportedPlatform means that the current operating system has no
	// implementation.
	ErrUnsupportedPlatform = errors.New("notification platform is unsupported")
)

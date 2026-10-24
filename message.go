package notify

import (
	"strings"
	"unicode"
)

const (
	// DefaultTitle is used when a caller supplies an empty title.
	DefaultTitle = "Desktop Notify"

	// MaxTitleRunes is the maximum supported title length.
	MaxTitleRunes = 128

	// MaxTextRunes is the maximum supported notification body length.
	MaxTextRunes = 1024
)

// Message describes one desktop notification.
type Message struct {
	Title string `json:"title,omitempty"`
	Text  string `json:"text"`
}

// Validate checks whether the message can be sent. It returns a sentinel
// error that can be inspected with errors.Is.
func (m Message) Validate() error {
	return m.validate()
}

func (m Message) validate() error {
	if strings.TrimSpace(m.Text) == "" {
		return ErrEmptyText
	}
	if len([]rune(m.Text)) > MaxTextRunes {
		return ErrTextTooLong
	}
	if len([]rune(m.Title)) > MaxTitleRunes {
		return ErrTitleTooLong
	}
	if containsUnsupportedControl(m.Text) {
		return ErrInvalidText
	}
	if containsUnsupportedControl(m.Title) {
		return ErrInvalidTitle
	}
	return nil
}

func containsUnsupportedControl(value string) bool {
	for _, r := range value {
		if unicode.IsControl(r) && r != '\n' && r != '\t' && r != '\r' {
			return true
		}
	}
	return false
}

func (m Message) normalized() Message {
	if strings.TrimSpace(m.Title) == "" {
		m.Title = DefaultTitle
	}
	return m
}

package notify

import (
	"context"
	"errors"
	"sync"
	"testing"
)

type mockNotifier struct {
	mu       sync.Mutex
	messages []Message
	err      error
}

func (m *mockNotifier) send(_ context.Context, message Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, message)
	return m.err
}

func (m *mockNotifier) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.messages)
}

func TestClientSend(t *testing.T) {
	backend := &mockNotifier{}
	client := &Client{backend: backend}

	if err := client.Send(Message{Text: "Your code is 123456"}); err != nil {
		t.Fatalf("Send() error = %v", err)
	}
	if backend.count() != 1 {
		t.Fatalf("backend received %d messages, want 1", backend.count())
	}
	if got := backend.messages[0].Title; got != DefaultTitle {
		t.Errorf("default title = %q, want %q", got, DefaultTitle)
	}
}

func TestClientValidation(t *testing.T) {
	tests := []struct {
		name string
		msg  Message
		want error
	}{
		{name: "empty", msg: Message{Text: ""}, want: ErrEmptyText},
		{name: "whitespace", msg: Message{Text: " \n\t "}, want: ErrEmptyText},
		{name: "long text", msg: Message{Text: string(make([]rune, MaxTextRunes+1))}, want: ErrTextTooLong},
		{name: "long title", msg: Message{Title: string(make([]rune, MaxTitleRunes+1)), Text: "hello"}, want: ErrTitleTooLong},
		{name: "control text", msg: Message{Text: "hello\x00"}, want: ErrInvalidText},
		{name: "control title", msg: Message{Title: "hello\x00", Text: "world"}, want: ErrInvalidTitle},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			backend := &mockNotifier{}
			err := (&Client{backend: backend}).Send(test.msg)
			if !errors.Is(err, test.want) {
				t.Errorf("Send() error = %v, want %v", err, test.want)
			}
			if got := backend.count(); got != 0 {
				t.Errorf("backend received %d messages for invalid input", got)
			}
		})
	}
}

func TestClientBackendFailure(t *testing.T) {
	want := errors.New("backend failed")
	err := (&Client{backend: &mockNotifier{err: want}}).Send(Message{Text: "hello"})
	if !errors.Is(err, want) {
		t.Fatalf("Send() error = %v, want %v", err, want)
	}
}

func TestClientNil(t *testing.T) {
	var client *Client
	if err := client.Send(Message{Text: "hello"}); !errors.Is(err, ErrBackendUnavailable) {
		t.Fatalf("nil Client.Send() error = %v, want %v", err, ErrBackendUnavailable)
	}
}

func TestClientConcurrentUse(t *testing.T) {
	backend := &mockNotifier{}
	client := &Client{backend: backend}
	const calls = 100

	var group sync.WaitGroup
	group.Add(calls)
	for index := 0; index < calls; index++ {
		go func(index int) {
			defer group.Done()
			if err := client.Send(Message{Text: "message"}); err != nil {
				t.Errorf("call %d: Send() error = %v", index, err)
			}
		}(index)
	}
	group.Wait()

	if got := backend.count(); got != calls {
		t.Fatalf("backend received %d messages, want %d", got, calls)
	}
}

func TestNew(t *testing.T) {
	if client := New(); client == nil || client.backend == nil {
		t.Fatal("New() returned a client without a backend")
	}
}

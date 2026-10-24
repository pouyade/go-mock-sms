package httpapi

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/pouyade/go-mock-sms"
)

type mockSender struct {
	mu       sync.Mutex
	messages []notify.Message
	err      error
}

func (m *mockSender) Send(message notify.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.messages = append(m.messages, message)
	return m.err
}

func (m *mockSender) count() int {
	m.mu.Lock()
	defer m.mu.Unlock()
	return len(m.messages)
}

func newTestHandler(sender Sender) *Handler {
	return New(sender, "Configured title", slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func perform(handler http.Handler, method, path, contentType, body string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, bytes.NewBufferString(body))
	if contentType != "" {
		request.Header.Set("Content-Type", contentType)
	}
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, request)
	return recorder
}

func decodeJSON(t *testing.T, recorder *httptest.ResponseRecorder, target any) {
	t.Helper()
	if err := json.Unmarshal(recorder.Body.Bytes(), target); err != nil {
		t.Fatalf("response is not JSON: %v; body=%q", err, recorder.Body.String())
	}
}

func TestSend(t *testing.T) {
	sender := &mockSender{}
	recorder := perform(newTestHandler(sender), http.MethodPost, "/send", "application/json; charset=utf-8",
		`{"to":"number","text":"hello","title":"Custom"}`)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	var response successResponse
	decodeJSON(t, recorder, &response)
	if !response.Success {
		t.Error("success response has success=false")
	}
	if sender.count() != 1 || sender.messages[0].Title != "Custom" || sender.messages[0].Text != "hello" {
		t.Fatalf("sender received %#v", sender.messages)
	}
}

func TestSendUsesDefaultTitle(t *testing.T) {
	sender := &mockSender{}
	perform(newTestHandler(sender), http.MethodPost, "/send", "application/json", `{"text":"hello"}`)
	if got := sender.messages[0].Title; got != "Configured title" {
		t.Fatalf("title = %q, want Configured title", got)
	}
}

func TestSendErrors(t *testing.T) {
	tests := []struct {
		name        string
		method      string
		contentType string
		body        string
		status      int
		code        string
	}{
		{name: "missing text", method: http.MethodPost, contentType: "application/json", body: `{}`, status: http.StatusBadRequest, code: "invalid_request"},
		{name: "empty text", method: http.MethodPost, contentType: "application/json", body: `{"text":"  "}`, status: http.StatusBadRequest, code: "invalid_request"},
		{name: "invalid json", method: http.MethodPost, contentType: "application/json", body: `{"text":`, status: http.StatusBadRequest, code: "invalid_json"},
		{name: "wrong content type", method: http.MethodPost, contentType: "text/plain", body: `{"text":"hello"}`, status: http.StatusUnsupportedMediaType, code: "unsupported_media_type"},
		{name: "unsupported method", method: http.MethodGet, contentType: "application/json", body: `{}`, status: http.StatusMethodNotAllowed, code: "method_not_allowed"},
		{name: "trailing json", method: http.MethodPost, contentType: "application/json", body: `{"text":"hello"} {}`, status: http.StatusBadRequest, code: "invalid_json"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recorder := perform(newTestHandler(&mockSender{}), test.method, "/send", test.contentType, test.body)
			if recorder.Code != test.status {
				t.Fatalf("status = %d, want %d", recorder.Code, test.status)
			}
			var response errorResponse
			decodeJSON(t, recorder, &response)
			if response.Success || response.Error != test.code {
				t.Fatalf("response = %#v, want error code %q", response, test.code)
			}
		})
	}
}

func TestSendTooLarge(t *testing.T) {
	handler := newTestHandler(&mockSender{})
	body := `{"text":"` + string(bytes.Repeat([]byte("a"), int(DefaultMaxBodyBytes))) + `"}`
	recorder := perform(handler, http.MethodPost, "/send", "application/json", body)
	if recorder.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("status = %d, want 413", recorder.Code)
	}
	var response errorResponse
	decodeJSON(t, recorder, &response)
	if response.Error != "payload_too_large" {
		t.Fatalf("error code = %q, want payload_too_large", response.Error)
	}
}

func TestBackendFailure(t *testing.T) {
	sender := &mockSender{err: errors.New("backend failed")}
	recorder := perform(newTestHandler(sender), http.MethodPost, "/send", "application/json", `{"text":"hello"}`)
	if recorder.Code != http.StatusInternalServerError {
		t.Fatalf("status = %d, want 500", recorder.Code)
	}
	var response errorResponse
	decodeJSON(t, recorder, &response)
	if response.Error != "notification_failed" {
		t.Fatalf("error code = %q, want notification_failed", response.Error)
	}
}

func TestHealth(t *testing.T) {
	recorder := perform(newTestHandler(&mockSender{}), http.MethodGet, "/health", "", "")
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", recorder.Code)
	}
	var response healthResponse
	decodeJSON(t, recorder, &response)
	if response.Status != "ok" {
		t.Fatalf("status response = %#v", response)
	}

	methodRecorder := perform(newTestHandler(&mockSender{}), http.MethodPost, "/health", "", "")
	if methodRecorder.Code != http.StatusMethodNotAllowed {
		t.Fatalf("POST /health status = %d, want 405", methodRecorder.Code)
	}
	if got := methodRecorder.Header().Get("Allow"); got != http.MethodGet {
		t.Fatalf("Allow = %q, want GET", got)
	}
}

func TestNotFound(t *testing.T) {
	recorder := perform(newTestHandler(&mockSender{}), http.MethodGet, "/unknown", "", "")
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", recorder.Code)
	}
}

func TestConcurrentRequests(t *testing.T) {
	sender := &mockSender{}
	handler := newTestHandler(sender)
	const requests = 100
	var group sync.WaitGroup
	group.Add(requests)
	for index := 0; index < requests; index++ {
		go func() {
			defer group.Done()
			recorder := perform(handler, http.MethodPost, "/send", "application/json", `{"text":"hello"}`)
			if recorder.Code != http.StatusOK {
				t.Errorf("status = %d, want 200", recorder.Code)
			}
		}()
	}
	group.Wait()
	if got := sender.count(); got != requests {
		t.Fatalf("sender received %d messages, want %d", got, requests)
	}
}

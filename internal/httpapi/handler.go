package httpapi

import (
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"mime"
	"net/http"
	"strings"
	"time"

	"github.com/pouyade/go-mock-sms"
)

const DefaultMaxBodyBytes int64 = 64 << 10

// Sender is the portion of the notify client needed by the HTTP API.
type Sender interface {
	Send(notify.Message) error
}

// Handler serves the local notification API.
type Handler struct {
	sender       Sender
	defaultTitle string
	maxBodyBytes int64
	logger       *slog.Logger
}

// New creates an HTTP API handler. A nil logger uses slog.Default().
func New(sender Sender, defaultTitle string, logger *slog.Logger) *Handler {
	if logger == nil {
		logger = slog.Default()
	}
	if strings.TrimSpace(defaultTitle) == "" {
		defaultTitle = notify.DefaultTitle
	}
	return &Handler{
		sender:       sender,
		defaultTitle: defaultTitle,
		maxBodyBytes: DefaultMaxBodyBytes,
		logger:       logger,
	}
}

type sendRequest struct {
	To    string `json:"to"`
	Text  string `json:"text"`
	Title string `json:"title"`
}

type successResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
}

type errorResponse struct {
	Success bool   `json:"success"`
	Error   string `json:"error"`
	Message string `json:"message"`
}

type healthResponse struct {
	Status string `json:"status"`
}

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	recorder := &statusRecorder{ResponseWriter: w, status: http.StatusOK}
	defer func() {
		if recovered := recover(); recovered != nil {
			h.writeError(recorder, http.StatusInternalServerError, "internal_error", "Internal server error")
		}
		h.logger.Info("http request",
			"method", r.Method,
			"path", r.URL.Path,
			"status", recorder.status,
			"duration_ms", time.Since(start).Milliseconds(),
		)
	}()

	switch r.URL.Path {
	case "/send":
		h.handleSend(recorder, r)
	case "/health":
		h.handleHealth(recorder, r)
	default:
		h.writeError(recorder, http.StatusNotFound, "not_found", "Endpoint not found")
	}
}

func (h *Handler) handleHealth(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		h.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}
	h.writeJSON(w, http.StatusOK, healthResponse{Status: "ok"})
}

func (h *Handler) handleSend(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", http.MethodPost)
		h.writeError(w, http.StatusMethodNotAllowed, "method_not_allowed", "Method not allowed")
		return
	}
	if !hasJSONContentType(r.Header.Get("Content-Type")) {
		h.writeError(w, http.StatusUnsupportedMediaType, "unsupported_media_type", "Content-Type must be application/json")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, h.maxBodyBytes)
	var request sendRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&request); err != nil {
		if isBodyTooLarge(err) {
			h.writeError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "Request body is too large")
			return
		}
		h.writeError(w, http.StatusBadRequest, "invalid_json", "Request body must contain valid JSON")
		return
	}
	var extra any
	if err := decoder.Decode(&extra); err != io.EOF {
		if isBodyTooLarge(err) {
			h.writeError(w, http.StatusRequestEntityTooLarge, "payload_too_large", "Request body is too large")
			return
		}
		h.writeError(w, http.StatusBadRequest, "invalid_json", "Request body must contain one JSON object")
		return
	}

	title := request.Title
	if strings.TrimSpace(title) == "" {
		title = h.defaultTitle
	}
	message := notify.Message{Title: title, Text: request.Text}
	if err := message.Validate(); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid_request", "Notification fields are invalid")
		return
	}
	if err := h.sender.Send(message); err != nil {
		h.logger.Error("notification backend failed")
		h.writeError(w, http.StatusInternalServerError, "notification_failed", "The operating system notification backend failed")
		return
	}
	h.writeJSON(w, http.StatusOK, successResponse{
		Success: true,
		Message: "Notification request accepted",
	})
}

func (h *Handler) writeError(w http.ResponseWriter, status int, code, message string) {
	h.writeJSON(w, status, errorResponse{Success: false, Error: code, Message: message})
}

func (h *Handler) writeJSON(w http.ResponseWriter, status int, value any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(value)
}

func hasJSONContentType(value string) bool {
	mediaType, _, err := mime.ParseMediaType(value)
	return err == nil && strings.EqualFold(mediaType, "application/json")
}

func isBodyTooLarge(err error) bool {
	var maxBytesError *http.MaxBytesError
	return errors.As(err, &maxBytesError)
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (w *statusRecorder) WriteHeader(status int) {
	w.status = status
	w.ResponseWriter.WriteHeader(status)
}

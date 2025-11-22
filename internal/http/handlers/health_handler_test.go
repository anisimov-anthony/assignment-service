package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHealthHandlerHealth(t *testing.T) {
	handler := NewHealthHandler()

	t.Run("GET request returns ok", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/health", nil)
		w := httptest.NewRecorder()

		handler.Health(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("expected status %d, got %d", http.StatusOK, w.Code)
		}
		if w.Header().Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type application/json, got %s", w.Header().Get("Content-Type"))
		}
		expectedBody := `{"status":"ok"}`
		if body := w.Body.String(); body != expectedBody+"\n" {
			t.Errorf("expected body %q, got %q", expectedBody+"\n", body)
		}
	})

	t.Run("POST request returns method not allowed", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/health", nil)
		w := httptest.NewRecorder()
		handler.Health(w, req)
		if w.Code != http.StatusMethodNotAllowed {
			t.Errorf("expected status %d, got %d", http.StatusMethodNotAllowed, w.Code)
		}
	})
}

// errorWriter like ResponseWriter, which always returns an error when writing
type errorWriter struct {
	header http.Header
	code   int
}

func (w *errorWriter) Header() http.Header {
	if w.header == nil {
		w.header = make(http.Header)
	}
	return w.header
}

func (w *errorWriter) Write([]byte) (int, error) {
	return 0, &json.UnsupportedTypeError{}
}

func (w *errorWriter) WriteHeader(statusCode int) {
	w.code = statusCode
}

func TestHealthHandlerEncodingError(t *testing.T) {
	handler := NewHealthHandler()

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	w := &errorWriter{}

	handler.Health(w, req)

	if w.code != http.StatusInternalServerError {
		t.Errorf("expected status %d, got %d", http.StatusInternalServerError, w.code)
	}
}

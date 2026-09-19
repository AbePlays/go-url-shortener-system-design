package api_test

import (
	"bytes"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/AbePlays/go-url-shortener-system-design/internal/api"
)

func TestLoggingMiddleware_PassesThroughAndCapturesStatus(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})

	wrapped := api.LoggingMiddleware(next)

	req := httptest.NewRequest(http.MethodGet, "/some-path", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusTeapot {
		t.Errorf("expected status %d, got %d", http.StatusTeapot, rec.Code)
	}
}

func TestLoggingMiddleware_DefaultsToOKWhenWriteHeaderNotCalled(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Write([]byte("hello"))
	})

	wrapped := api.LoggingMiddleware(next)

	req := httptest.NewRequest(http.MethodGet, "/some-path", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("expected default status %d, got %d", http.StatusOK, rec.Code)
	}
}

func TestLoggingMiddleware_LogsRequestDetails(t *testing.T) {
	var logOutput bytes.Buffer
	handler := slog.NewTextHandler(&logOutput, nil)
	originalLogger := slog.Default()
	slog.SetDefault(slog.New(handler))
	defer slog.SetDefault(originalLogger)

	next := http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	wrapped := api.LoggingMiddleware(next)

	req := httptest.NewRequest(http.MethodPost, "/api/shorten", nil)
	rec := httptest.NewRecorder()

	wrapped.ServeHTTP(rec, req)

	got := logOutput.String()
	if !strings.Contains(got, "POST") || !strings.Contains(got, "/api/shorten") || !strings.Contains(got, "201") {
		t.Errorf("expected log to contain method, path, and status, got: %s", got)
	}
}

package api_test

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/AbePlays/go-url-shortener-system-design/internal/api"
	"github.com/AbePlays/go-url-shortener-system-design/internal/store"
)

func TestShortenFormHandler_Success(t *testing.T) {
	db := testDB(t)
	c := testCache(t)
	s := store.New(db, c)
	h := api.New(s, "http://localhost:8080")

	targetUrl := "https://example.com"
	form := "url=" + url.QueryEscape(targetUrl)

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.ShortenFormHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ShortenFormHandler status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "http://localhost:8080/") {
		t.Errorf("expected rendered page to contain the short URL, got: %s", body)
	}

	// No structured response to pull the code from - query the DB directly
	// for the row we just created, so we can clean it up afterwards.
	var code string
	err := db.QueryRow("SELECT code FROM urls WHERE original_url = $1 ORDER BY created_at DESC LIMIT 1", targetUrl).Scan(&code)
	if err != nil {
		t.Fatalf("failed to find created row for cleanup: %v", err)
	}
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM urls WHERE code = $1", code)
	})
}

func TestShortenFormHandler_InvalidUrl(t *testing.T) {
	db := testDB(t)
	c := testCache(t)
	s := store.New(db, c)
	h := api.New(s, "http://localhost:8080")

	form := "url=" + url.QueryEscape("not-a-real-url")

	req := httptest.NewRequest(http.MethodPost, "/shorten", strings.NewReader(form))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()

	h.ShortenFormHandler(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("ShortenFormHandler status = %d, want %d", rec.Code, http.StatusOK)
	}

	body := rec.Body.String()
	if !strings.Contains(body, "Invalid url") {
		t.Errorf("expected rendered page to contain the error message, got: %s", body)
	}
}

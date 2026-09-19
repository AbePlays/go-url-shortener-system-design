package api_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/AbePlays/go-url-shortener-system-design/internal/api"
	"github.com/AbePlays/go-url-shortener-system-design/internal/store"
)

func TestShortenAndRedirect(t *testing.T) {
	s := store.New()
	h := api.New(s)

	body, err := json.Marshal(api.ShortenRequest{Url: "https://example.com"})
	if err != nil {
		t.Fatalf("failed to marshal request body: %v", err)
	}

	shortenReq := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewReader(body))
	shortenRec := httptest.NewRecorder()

	h.ShortenHandler(shortenRec, shortenReq)

	if shortenRec.Code != http.StatusOK {
		t.Fatalf("ShortenHandler status = %d, want %d", shortenRec.Code, http.StatusOK)
	}

	var shortenResp api.ShortenResponse
	if err := json.NewDecoder(shortenRec.Body).Decode(&shortenResp); err != nil {
		t.Fatalf("failed to decode shorten response: %v", err)
	}

	if shortenResp.ShortCode == "" {
		t.Fatal("expected non-empty ShortCode, got empty string")
	}

	redirectReq := httptest.NewRequest(http.MethodGet, "/"+shortenResp.ShortCode, nil)
	redirectReq.SetPathValue("code", shortenResp.ShortCode)
	redirectRec := httptest.NewRecorder()

	h.RedirectHandler(redirectRec, redirectReq)

	if redirectRec.Code != http.StatusFound {
		t.Fatalf("RedirectHandler status = %d, want %d", redirectRec.Code, http.StatusFound)
	}

	gotLocation := redirectRec.Header().Get("Location")
	if gotLocation != "https://example.com" {
		t.Errorf("RedirectHandler Location = %q, want %q", gotLocation, "https://example.com")
	}
}

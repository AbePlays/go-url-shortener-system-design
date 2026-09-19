package store_test

import (
	"testing"

	"github.com/AbePlays/go-url-shortener-system-design/internal/store"
)

func TestAddUrlAndGetUrl(t *testing.T) {
	s := store.New()

	url := "https://example.com"
	code := s.AddUrl(url)

	if len(code) != 8 {
		t.Errorf("AddUrl() returned code of length %d, want 8", len(code))
	}

	got, err := s.GetUrl(code)
	if err != nil {
		t.Fatalf("GetUrl(%q) returned unexpected error: %v", code, err)
	}

	if got != url {
		t.Errorf("GetUrl(%q) = %q, want %q", code, got, url)
	}
}

func TestGetUrl_NotFound(t *testing.T) {
	s := store.New()

	_, err := s.GetUrl("doesnotexist")
	if err == nil {
		t.Error("GetUrl() with unknown code expected an error, got nil")
	}
}

func TestAddUrl_UniqueCodesForSameInput(t *testing.T) {
	s := store.New()

	code1 := s.AddUrl("https://example.com")
	code2 := s.AddUrl("https://example.com")

	if code1 == code2 {
		t.Errorf("AddUrl() called twice with the same URL returned the same code: %q", code1)
	}
}

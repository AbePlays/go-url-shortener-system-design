package store_test

import (
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"

	"github.com/AbePlays/go-url-shortener-system-design/internal/store"
)

func testDB(t *testing.T) *sql.DB {
	t.Helper()

	dbUrl := os.Getenv("DATABASE_URL")
	if dbUrl == "" {
		t.Skip("DATABASE_URL not set, skipping test that requires a real database")
	}

	db, err := sql.Open("pgx", dbUrl)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	t.Cleanup(func() { db.Close() })

	if err := db.Ping(); err != nil {
		t.Fatalf("failed to ping database: %v", err)
	}

	return db
}

func cleanupCode(t *testing.T, db *sql.DB, code string) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM urls WHERE code = $1", code)
	})
}

func TestAddUrlAndGetUrl(t *testing.T) {
	db := testDB(t)
	s := store.New(db)

	url := "https://example.com"
	code, err := s.AddUrl(url)
	if err != nil {
		t.Fatalf("AddUrl() returned unexpected error: %v", err)
	}
	cleanupCode(t, db, code)

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
	db := testDB(t)
	s := store.New(db)

	_, err := s.GetUrl("doesnotexist")
	if err == nil {
		t.Error("GetUrl() with unknown code expected an error, got nil")
	}
}

func TestAddUrl_UniqueCodesForSameInput(t *testing.T) {
	db := testDB(t)
	s := store.New(db)

	code1, err := s.AddUrl("https://example.com")
	if err != nil {
		t.Fatalf("AddUrl() returned unexpected error: %v", err)
	}
	cleanupCode(t, db, code1)

	code2, err := s.AddUrl("https://example.com")
	if err != nil {
		t.Fatalf("AddUrl() returned unexpected error: %v", err)
	}
	cleanupCode(t, db, code2)

	if code1 == code2 {
		t.Errorf("AddUrl() called twice with the same URL returned the same code: %q", code1)
	}
}

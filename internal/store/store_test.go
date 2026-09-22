package store_test

import (
	"context"
	"database/sql"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"

	"github.com/AbePlays/go-url-shortener-system-design/internal/cache"
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

func testCache(t *testing.T) *cache.UrlCache {
	t.Helper()

	redisUrl := os.Getenv("REDIS_URL")
	if redisUrl == "" {
		t.Skip("REDIS_URL not set, skipping test that requires a real cache")
	}

	opts, err := redis.ParseURL(redisUrl)
	if err != nil {
		t.Fatalf("failed to parse redis url: %v", err)
	}

	client := redis.NewClient(opts)
	t.Cleanup(func() { client.Close() })

	if err := client.Ping(context.Background()).Err(); err != nil {
		t.Fatalf("failed to ping redis: %v", err)
	}

	return cache.New(client)
}

func cleanupCode(t *testing.T, db *sql.DB, code string) {
	t.Helper()
	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM urls WHERE code = $1", code)
	})
}

func TestAddUrlAndGetUrl(t *testing.T) {
	db := testDB(t)
	c := testCache(t)
	s := store.New(db, c)

	url := "https://example.com"
	code, err := s.AddUrl(url)
	if err != nil {
		t.Fatalf("AddUrl() returned unexpected error: %v", err)
	}
	cleanupCode(t, db, code)

	if len(code) != 8 {
		t.Errorf("AddUrl() returned code of length %d, want 8", len(code))
	}

	// AddUrl never touches the cache, so this call is guaranteed to be a
	// cache miss - it must fall through to Postgres to succeed at all.
	got, err := s.GetUrl(context.Background(), code)
	if err != nil {
		t.Fatalf("GetUrl(%q) returned unexpected error: %v", code, err)
	}

	if got != url {
		t.Errorf("GetUrl(%q) = %q, want %q", code, got, url)
	}
}

func TestGetUrl_CacheHit(t *testing.T) {
	db := testDB(t)
	c := testCache(t)
	s := store.New(db, c)
	ctx := context.Background()

	url := "https://example.com"
	code, err := s.AddUrl(url)
	if err != nil {
		t.Fatalf("AddUrl() returned unexpected error: %v", err)
	}
	cleanupCode(t, db, code)

	// first call - cache miss, falls through to Postgres, populates cache
	if _, err := s.GetUrl(ctx, code); err != nil {
		t.Fatalf("first GetUrl(%q) returned unexpected error: %v", code, err)
	}

	// remove it from Postgres directly, so only the cache can possibly
	// have it now - proves the next call genuinely comes from the cache
	if _, err := db.Exec("DELETE FROM urls WHERE code = $1", code); err != nil {
		t.Fatalf("failed to delete row directly: %v", err)
	}

	got, err := s.GetUrl(ctx, code)
	if err != nil {
		t.Fatalf("second GetUrl(%q) returned unexpected error (should have hit cache): %v", code, err)
	}

	if got != url {
		t.Errorf("GetUrl(%q) = %q, want %q", code, got, url)
	}
}

func TestGetUrl_NotFound(t *testing.T) {
	db := testDB(t)
	c := testCache(t)
	s := store.New(db, c)

	_, err := s.GetUrl(context.Background(), "doesnotexist")
	if err == nil {
		t.Error("GetUrl() with unknown code expected an error, got nil")
	}
}

func TestAddUrl_UniqueCodesForSameInput(t *testing.T) {
	db := testDB(t)
	c := testCache(t)
	s := store.New(db, c)

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

func TestListUrls(t *testing.T) {
	db := testDB(t)
	c := testCache(t)
	s := store.New(db, c)

	before, err := s.ListUrls(context.Background())
	if err != nil {
		t.Fatalf("ListUrls() returned unexpected error: %v", err)
	}

	code, err := s.AddUrl("https://example.com")
	if err != nil {
		t.Fatalf("AddUrl() returned unexpected error: %v", err)
	}
	cleanupCode(t, db, code)

	after, err := s.ListUrls(context.Background())
	if err != nil {
		t.Fatalf("ListUrls() returned unexpected error: %v", err)
	}

	if len(after) != len(before)+1 {
		t.Errorf("ListUrls() length = %d, want %d", len(after), len(before)+1)
	}
}

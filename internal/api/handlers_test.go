package api_test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/redis/go-redis/v9"

	"github.com/AbePlays/go-url-shortener-system-design/internal/api"
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

func TestShortenAndRedirect(t *testing.T) {
	db := testDB(t)
	c := testCache(t)
	s := store.New(db, c)
	h := api.New(s, "http://localhost:8080")

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

	t.Cleanup(func() {
		_, _ = db.Exec("DELETE FROM urls WHERE code = $1", shortenResp.ShortCode)
	})

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

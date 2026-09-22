package store

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	"github.com/AbePlays/go-url-shortener-system-design/internal/cache"
	"github.com/AbePlays/go-url-shortener-system-design/utils"
	"github.com/jackc/pgx/v5/pgconn"
)

const cacheTtl = 24 * time.Hour

type UrlStore struct {
	db    *sql.DB
	cache *cache.UrlCache
}

type Url struct {
	Code        string
	OriginalUrl string
	Clicks      int
	CreatedAt   time.Time
}

func New(db *sql.DB, cache *cache.UrlCache) *UrlStore {
	return &UrlStore{db: db, cache: cache}
}

func (s *UrlStore) AddUrl(url string) (string, error) {
	for {
		shortCode := utils.GenerateShortCode(8)

		_, err := s.db.Exec("INSERT INTO urls (code, original_url) VALUES ($1, $2)", shortCode, url)
		if err == nil {
			return shortCode, nil
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			continue
		}

		return "", err
	}
}

func (s *UrlStore) GetUrl(ctx context.Context, code string) (string, error) {
	cachedUrl, err := s.cache.Get(ctx, code)
	if err == nil {
		slog.Info("cache hit", "code", code)
		go s.incrementClicks(code)
		return cachedUrl, nil
	}

	if !errors.Is(err, cache.ErrCacheMiss) {
		return "", err
	}

	slog.Info("cache miss", "code", code)

	var url string
	err = s.db.QueryRow("SELECT original_url from urls where code = $1", code).Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("code not found")
		}

		return "", err
	}

	if err := s.cache.Set(ctx, code, url, cacheTtl); err != nil {
		slog.Error("failed to populate cache", "code", code, "error", err)
	}

	go s.incrementClicks(code)

	return url, nil
}

func (s *UrlStore) ListUrls(ctx context.Context) ([]Url, error) {
	var urls []Url
	rows, err := s.db.Query("SELECT code, original_url, clicks, created_at FROM urls ORDER BY created_at DESC")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		if rows.Err() != nil {
			return nil, rows.Err()
		}

		var url Url
		if err := rows.Scan(&url.Code, &url.OriginalUrl, &url.Clicks, &url.CreatedAt); err != nil {
			return nil, err
		}
		urls = append(urls, url)
	}

	return urls, nil
}

func (s *UrlStore) incrementClicks(code string) {
	_, err := s.db.Exec("UPDATE urls SET clicks = clicks + 1 WHERE code = $1", code)
	if err != nil {
		slog.Error("failed to increment clicks", "code", code, "error", err)
	}
}

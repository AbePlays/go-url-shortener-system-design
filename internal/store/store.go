package store

import (
	"database/sql"
	"errors"

	"github.com/AbePlays/go-url-shortener-system-design/utils"
	"github.com/jackc/pgx/v5/pgconn"
)

type UrlStore struct {
	db *sql.DB
}

func New(db *sql.DB) *UrlStore {
	return &UrlStore{db: db}
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

func (s *UrlStore) GetUrl(code string) (string, error) {
	var url string
	err := s.db.QueryRow("SELECT original_url from urls where code = $1", code).Scan(&url)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", errors.New("code not found")
		}

		return "", err
	}

	return url, nil
}

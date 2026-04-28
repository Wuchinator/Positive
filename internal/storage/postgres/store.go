package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"positive/internal/shortener"
)

type Store struct {
	pool *pgxpool.Pool
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

func (s *Store) Save(ctx context.Context, code string, originalURL string) (string, error) {
	var storedCode string
	err := s.pool.QueryRow(ctx, `
INSERT INTO short_urls (code, original_url)
VALUES ($1, $2)
ON CONFLICT (original_url) DO UPDATE SET original_url = EXCLUDED.original_url
RETURNING code`, code, originalURL).Scan(&storedCode)
	if err == nil {
		return storedCode, nil
	}

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" && pgErr.ConstraintName == "short_urls_pkey" {
		return "", shortener.ErrCodeExists
	}

	return "", err
}

func (s *Store) Resolve(ctx context.Context, code string) (string, error) {
	var originalURL string
	err := s.pool.QueryRow(ctx, `
SELECT original_url
FROM short_urls
WHERE code = $1`, code).Scan(&originalURL)
	if errors.Is(err, pgx.ErrNoRows) {
		return "", shortener.ErrNotFound
	}
	if err != nil {
		return "", err
	}

	return originalURL, nil
}

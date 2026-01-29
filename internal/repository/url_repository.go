package repository

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/edson-mazvila/url-shortener/internal/domain"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// URLRepository handles database operations for URLs.
type URLRepository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewURLRepository creates a new URL repository.
func NewURLRepository(pool *pgxpool.Pool, logger *slog.Logger) *URLRepository {
	return &URLRepository{
		pool:   pool,
		logger: logger,
	}
}

// Create creates a new shortened URL in the database.
func (r *URLRepository) Create(ctx context.Context, url *domain.URL) error {
	query := `
		INSERT INTO urls (short_code, original_url, created_at, expires_at, access_count)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id
	`

	err := r.pool.QueryRow(
		ctx,
		query,
		url.ShortCode,
		url.OriginalURL,
		url.CreatedAt,
		url.ExpiresAt,
		url.AccessCount,
	).Scan(&url.ID)

	if err != nil {
		if err.Error() == "ERROR: duplicate key value violates unique constraint \"urls_short_code_key\" (SQLSTATE 23505)" {
			return domain.ErrShortCodeAlreadyExists
		}
		return fmt.Errorf("failed to create url: %w", err)
	}

	r.logger.Debug("url created",
		slog.Int64("id", url.ID),
		slog.String("short_code", url.ShortCode),
	)

	return nil
}

// GetByShortCode retrieves a URL by its short code.
func (r *URLRepository) GetByShortCode(ctx context.Context, shortCode string) (*domain.URL, error) {
	query := `
		SELECT id, short_code, original_url, created_at, expires_at, access_count, last_accessed
		FROM urls
		WHERE short_code = $1
	`

	var url domain.URL
	err := r.pool.QueryRow(ctx, query, shortCode).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.CreatedAt,
		&url.ExpiresAt,
		&url.AccessCount,
		&url.LastAccessed,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrURLNotFound
		}
		return nil, fmt.Errorf("failed to get url by short code: %w", err)
	}

	return &url, nil
}

// GetByID retrieves a URL by its ID.
func (r *URLRepository) GetByID(ctx context.Context, id int64) (*domain.URL, error) {
	query := `
		SELECT id, short_code, original_url, created_at, expires_at, access_count, last_accessed
		FROM urls
		WHERE id = $1
	`

	var url domain.URL
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&url.ID,
		&url.ShortCode,
		&url.OriginalURL,
		&url.CreatedAt,
		&url.ExpiresAt,
		&url.AccessCount,
		&url.LastAccessed,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrURLNotFound
		}
		return nil, fmt.Errorf("failed to get url by id: %w", err)
	}

	return &url, nil
}

// Update updates an existing URL.
func (r *URLRepository) Update(ctx context.Context, url *domain.URL) error {
	query := `
		UPDATE urls
		SET access_count = $1, last_accessed = $2
		WHERE id = $3
	`

	result, err := r.pool.Exec(ctx, query, url.AccessCount, url.LastAccessed, url.ID)
	if err != nil {
		return fmt.Errorf("failed to update url: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrURLNotFound
	}

	r.logger.Debug("url updated",
		slog.Int64("id", url.ID),
		slog.Int64("access_count", url.AccessCount),
	)

	return nil
}

// Delete deletes a URL by its short code.
func (r *URLRepository) Delete(ctx context.Context, shortCode string) error {
	query := `DELETE FROM urls WHERE short_code = $1`

	result, err := r.pool.Exec(ctx, query, shortCode)
	if err != nil {
		return fmt.Errorf("failed to delete url: %w", err)
	}

	if result.RowsAffected() == 0 {
		return domain.ErrURLNotFound
	}

	r.logger.Debug("url deleted", slog.String("short_code", shortCode))

	return nil
}

// DeleteExpired deletes all expired URLs.
func (r *URLRepository) DeleteExpired(ctx context.Context) (int64, error) {
	query := `
		DELETE FROM urls
		WHERE expires_at IS NOT NULL AND expires_at < $1
	`

	result, err := r.pool.Exec(ctx, query, time.Now())
	if err != nil {
		return 0, fmt.Errorf("failed to delete expired urls: %w", err)
	}

	count := result.RowsAffected()
	if count > 0 {
		r.logger.Info("expired urls deleted", slog.Int64("count", count))
	}

	return count, nil
}

// List retrieves a paginated list of URLs.
func (r *URLRepository) List(ctx context.Context, limit, offset int) ([]*domain.URL, error) {
	query := `
		SELECT id, short_code, original_url, created_at, expires_at, access_count, last_accessed
		FROM urls
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("failed to list urls: %w", err)
	}
	defer rows.Close()

	var urls []*domain.URL
	for rows.Next() {
		var url domain.URL
		err := rows.Scan(
			&url.ID,
			&url.ShortCode,
			&url.OriginalURL,
			&url.CreatedAt,
			&url.ExpiresAt,
			&url.AccessCount,
			&url.LastAccessed,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan url row: %w", err)
		}
		urls = append(urls, &url)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating url rows: %w", err)
	}

	return urls, nil
}

// Count returns the total number of URLs in the database.
func (r *URLRepository) Count(ctx context.Context) (int64, error) {
	query := `SELECT COUNT(*) FROM urls`

	var count int64
	err := r.pool.QueryRow(ctx, query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to count urls: %w", err)
	}

	return count, nil
}

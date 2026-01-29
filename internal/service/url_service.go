package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log/slog"
	"net/url"
	"strings"
	"time"

	"github.com/edson-mazvila/url-shortener/internal/config"
	"github.com/edson-mazvila/url-shortener/internal/domain"
	"github.com/edson-mazvila/url-shortener/internal/repository"
)

// URLService provides business logic for URL operations.
type URLService struct {
	repo   *repository.URLRepository
	config *config.URLConfig
	logger *slog.Logger
}

// NewURLService creates a new URL service.
func NewURLService(repo *repository.URLRepository, cfg *config.URLConfig, logger *slog.Logger) *URLService {
	return &URLService{
		repo:   repo,
		config: cfg,
		logger: logger,
	}
}

// CreateShortURL creates a new shortened URL.
func (s *URLService) CreateShortURL(ctx context.Context, originalURL string, customCode string, ttl time.Duration) (*domain.URL, error) {
	if err := s.validateURL(originalURL); err != nil {
		return nil, fmt.Errorf("invalid url: %w", err)
	}

	var shortCode string
	var err error

	if customCode != "" {
		if err := s.validateShortCode(customCode); err != nil {
			return nil, err
		}
		shortCode = customCode
	} else {
		shortCode, err = s.generateShortCode(ctx)
		if err != nil {
			return nil, fmt.Errorf("failed to generate short code: %w", err)
		}
	}

	now := time.Now()
	var expiresAt *time.Time

	if ttl > 0 {
		expiry := now.Add(ttl)
		expiresAt = &expiry
	} else if s.config.DefaultTTL > 0 {
		expiry := now.Add(s.config.DefaultTTL)
		expiresAt = &expiry
	}

	urlEntity := &domain.URL{
		ShortCode:   shortCode,
		OriginalURL: originalURL,
		CreatedAt:   now,
		ExpiresAt:   expiresAt,
		AccessCount: 0,
	}

	if err := s.repo.Create(ctx, urlEntity); err != nil {
		return nil, fmt.Errorf("failed to create url: %w", err)
	}

	s.logger.Info("short url created",
		slog.String("short_code", shortCode),
		slog.String("original_url", originalURL),
		slog.Any("expires_at", expiresAt),
	)

	return urlEntity, nil
}

// GetOriginalURL retrieves the original URL by short code and increments access count.
func (s *URLService) GetOriginalURL(ctx context.Context, shortCode string) (*domain.URL, error) {
	urlEntity, err := s.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	if urlEntity.IsExpired() {
		s.logger.Warn("attempted to access expired url",
			slog.String("short_code", shortCode),
		)
		return nil, domain.ErrURLExpired
	}

	urlEntity.IncrementAccessCount()

	if err := s.repo.Update(ctx, urlEntity); err != nil {
		s.logger.Error("failed to update access count",
			slog.String("error", err.Error()),
			slog.String("short_code", shortCode),
		)
	}

	return urlEntity, nil
}

// GetURLMetadata retrieves URL metadata without incrementing access count.
func (s *URLService) GetURLMetadata(ctx context.Context, shortCode string) (*domain.URL, error) {
	urlEntity, err := s.repo.GetByShortCode(ctx, shortCode)
	if err != nil {
		return nil, err
	}

	return urlEntity, nil
}

// DeleteURL deletes a URL by its short code.
func (s *URLService) DeleteURL(ctx context.Context, shortCode string) error {
	if err := s.repo.Delete(ctx, shortCode); err != nil {
		return fmt.Errorf("failed to delete url: %w", err)
	}

	s.logger.Info("url deleted", slog.String("short_code", shortCode))

	return nil
}

// ListURLs retrieves a paginated list of URLs.
func (s *URLService) ListURLs(ctx context.Context, limit, offset int) ([]*domain.URL, int64, error) {
	if limit <= 0 || limit > 100 {
		limit = 20
	}

	if offset < 0 {
		offset = 0
	}

	urls, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list urls: %w", err)
	}

	total, err := s.repo.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count urls: %w", err)
	}

	return urls, total, nil
}

// CleanupExpired removes all expired URLs from the database.
func (s *URLService) CleanupExpired(ctx context.Context) (int64, error) {
	count, err := s.repo.DeleteExpired(ctx)
	if err != nil {
		return 0, fmt.Errorf("failed to cleanup expired urls: %w", err)
	}

	return count, nil
}

// GetFullURL returns the full shortened URL.
func (s *URLService) GetFullURL(shortCode string) string {
	baseURL := strings.TrimSuffix(s.config.BaseURL, "/")
	return fmt.Sprintf("%s/%s", baseURL, shortCode)
}

func (s *URLService) validateURL(rawURL string) error {
	if rawURL == "" {
		return domain.ErrInvalidURL
	}

	parsedURL, err := url.ParseRequestURI(rawURL)
	if err != nil {
		return domain.ErrInvalidURL
	}

	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return domain.ErrInvalidURL
	}

	if parsedURL.Host == "" {
		return domain.ErrInvalidURL
	}

	return nil
}

func (s *URLService) validateShortCode(code string) error {
	if len(code) < 3 || len(code) > 20 {
		return domain.ErrInvalidShortCode
	}

	for _, char := range code {
		if !((char >= 'a' && char <= 'z') ||
			(char >= 'A' && char <= 'Z') ||
			(char >= '0' && char <= '9') ||
			char == '-' || char == '_') {
			return domain.ErrInvalidShortCode
		}
	}

	return nil
}

func (s *URLService) generateShortCode(ctx context.Context) (string, error) {
	const maxAttempts = 10

	for attempt := 0; attempt < maxAttempts; attempt++ {
		code, err := generateRandomCode(s.config.ShortCodeLength)
		if err != nil {
			return "", fmt.Errorf("failed to generate random code: %w", err)
		}

		_, err = s.repo.GetByShortCode(ctx, code)
		if err == domain.ErrURLNotFound {
			return code, nil
		}
		if err != nil {
			return "", fmt.Errorf("failed to check code uniqueness: %w", err)
		}

		s.logger.Debug("short code collision, retrying",
			slog.String("code", code),
			slog.Int("attempt", attempt+1),
		)
	}

	return "", fmt.Errorf("failed to generate unique short code after %d attempts", maxAttempts)
}

func generateRandomCode(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	numBytes := (length * 6) / 8
	if (length*6)%8 != 0 {
		numBytes++
	}

	randomBytes := make([]byte, numBytes)
	if _, err := rand.Read(randomBytes); err != nil {
		return "", fmt.Errorf("failed to generate random bytes: %w", err)
	}

	encoded := base64.RawURLEncoding.EncodeToString(randomBytes)

	if len(encoded) < length {
		return "", fmt.Errorf("generated code too short")
	}

	code := encoded[:length]

	result := make([]byte, length)
	for i := 0; i < length; i++ {
		if code[i] == '-' || code[i] == '_' {
			result[i] = charset[int(randomBytes[i])%len(charset)]
		} else {
			result[i] = code[i]
		}
	}

	return string(result), nil
}

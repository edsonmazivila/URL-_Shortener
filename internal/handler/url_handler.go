package handler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/edson-mazvila/url-shortener/internal/domain"
	"github.com/edson-mazvila/url-shortener/internal/service"
	"github.com/go-chi/chi/v5"
)

// URLHandler handles HTTP requests for URL operations.
type URLHandler struct {
	service *service.URLService
	logger  *slog.Logger
}

// NewURLHandler creates a new URL handler.
func NewURLHandler(service *service.URLService, logger *slog.Logger) *URLHandler {
	return &URLHandler{
		service: service,
		logger:  logger,
	}
}

// CreateShortURLRequest represents the request body for creating a short URL.
type CreateShortURLRequest struct {
	URL        string `json:"url"`
	CustomCode string `json:"custom_code,omitempty"`
	TTL        int64  `json:"ttl,omitempty"`
}

// CreateShortURLResponse represents the response for creating a short URL.
type CreateShortURLResponse struct {
	ID          int64      `json:"id"`
	ShortCode   string     `json:"short_code"`
	ShortURL    string     `json:"short_url"`
	OriginalURL string     `json:"original_url"`
	CreatedAt   time.Time  `json:"created_at"`
	ExpiresAt   *time.Time `json:"expires_at,omitempty"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// ListURLsResponse represents the response for listing URLs.
type ListURLsResponse struct {
	URLs   []*domain.URL `json:"urls"`
	Total  int64         `json:"total"`
	Limit  int           `json:"limit"`
	Offset int           `json:"offset"`
}

// CreateShortURL handles POST /api/urls
func (h *URLHandler) CreateShortURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var req CreateShortURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.logger.Warn("invalid request body", slog.String("error", err.Error()))
		h.respondError(w, http.StatusBadRequest, "invalid request body", err.Error())
		return
	}

	if req.URL == "" {
		h.respondError(w, http.StatusBadRequest, "url is required", "")
		return
	}

	var ttl time.Duration
	if req.TTL > 0 {
		ttl = time.Duration(req.TTL) * time.Second
	}

	urlEntity, err := h.service.CreateShortURL(ctx, req.URL, req.CustomCode, ttl)
	if err != nil {
		h.handleServiceError(w, err, "failed to create short url")
		return
	}

	response := CreateShortURLResponse{
		ID:          urlEntity.ID,
		ShortCode:   urlEntity.ShortCode,
		ShortURL:    h.service.GetFullURL(urlEntity.ShortCode),
		OriginalURL: urlEntity.OriginalURL,
		CreatedAt:   urlEntity.CreatedAt,
		ExpiresAt:   urlEntity.ExpiresAt,
	}

	h.respondJSON(w, http.StatusCreated, response)
}

// RedirectToOriginal handles GET /{shortCode}
func (h *URLHandler) RedirectToOriginal(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shortCode := chi.URLParam(r, "shortCode")

	if shortCode == "" {
		h.respondError(w, http.StatusBadRequest, "short code is required", "")
		return
	}

	urlEntity, err := h.service.GetOriginalURL(ctx, shortCode)
	if err != nil {
		h.handleServiceError(w, err, "failed to get original url")
		return
	}

	h.logger.Debug("redirecting",
		slog.String("short_code", shortCode),
		slog.String("original_url", urlEntity.OriginalURL),
	)

	http.Redirect(w, r, urlEntity.OriginalURL, http.StatusMovedPermanently)
}

// GetURLMetadata handles GET /api/urls/{shortCode}
func (h *URLHandler) GetURLMetadata(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shortCode := chi.URLParam(r, "shortCode")

	if shortCode == "" {
		h.respondError(w, http.StatusBadRequest, "short code is required", "")
		return
	}

	urlEntity, err := h.service.GetURLMetadata(ctx, shortCode)
	if err != nil {
		h.handleServiceError(w, err, "failed to get url metadata")
		return
	}

	h.respondJSON(w, http.StatusOK, urlEntity)
}

// DeleteURL handles DELETE /api/urls/{shortCode}
func (h *URLHandler) DeleteURL(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	shortCode := chi.URLParam(r, "shortCode")

	if shortCode == "" {
		h.respondError(w, http.StatusBadRequest, "short code is required", "")
		return
	}

	if err := h.service.DeleteURL(ctx, shortCode); err != nil {
		h.handleServiceError(w, err, "failed to delete url")
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// ListURLs handles GET /api/urls
func (h *URLHandler) ListURLs(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	limit := 20
	offset := 0

	if limitStr := r.URL.Query().Get("limit"); limitStr != "" {
		if l, err := strconv.Atoi(limitStr); err == nil {
			limit = l
		}
	}

	if offsetStr := r.URL.Query().Get("offset"); offsetStr != "" {
		if o, err := strconv.Atoi(offsetStr); err == nil {
			offset = o
		}
	}

	urls, total, err := h.service.ListURLs(ctx, limit, offset)
	if err != nil {
		h.logger.Error("failed to list urls", slog.String("error", err.Error()))
		h.respondError(w, http.StatusInternalServerError, "failed to list urls", "")
		return
	}

	response := ListURLsResponse{
		URLs:   urls,
		Total:  total,
		Limit:  limit,
		Offset: offset,
	}

	h.respondJSON(w, http.StatusOK, response)
}

func (h *URLHandler) handleServiceError(w http.ResponseWriter, err error, logMsg string) {
	if errors.Is(err, domain.ErrURLNotFound) {
		h.respondError(w, http.StatusNotFound, "url not found", "")
		return
	}

	if errors.Is(err, domain.ErrURLExpired) {
		h.respondError(w, http.StatusGone, "url has expired", "")
		return
	}

	if errors.Is(err, domain.ErrInvalidURL) {
		h.respondError(w, http.StatusBadRequest, "invalid url", "")
		return
	}

	if errors.Is(err, domain.ErrShortCodeAlreadyExists) {
		h.respondError(w, http.StatusConflict, "short code already exists", "")
		return
	}

	if errors.Is(err, domain.ErrInvalidShortCode) {
		h.respondError(w, http.StatusBadRequest, "invalid short code", "")
		return
	}

	h.logger.Error(logMsg, slog.String("error", err.Error()))
	h.respondError(w, http.StatusInternalServerError, "internal server error", "")
}

func (h *URLHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode response", slog.String("error", err.Error()))
	}
}

func (h *URLHandler) respondError(w http.ResponseWriter, status int, error, message string) {
	response := ErrorResponse{
		Error:   error,
		Message: message,
	}
	h.respondJSON(w, status, response)
}

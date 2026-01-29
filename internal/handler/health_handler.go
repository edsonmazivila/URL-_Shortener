package handler

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
)

// HealthChecker defines the interface for health checking.
type HealthChecker interface {
	HealthCheck(ctx context.Context) error
}

// HealthHandler handles health check requests.
type HealthHandler struct {
	db     HealthChecker
	logger *slog.Logger
}

// NewHealthHandler creates a new health handler.
func NewHealthHandler(db HealthChecker, logger *slog.Logger) *HealthHandler {
	return &HealthHandler{
		db:     db,
		logger: logger,
	}
}

// HealthResponse represents the health check response.
type HealthResponse struct {
	Status   string            `json:"status"`
	Database string            `json:"database"`
	Details  map[string]string `json:"details,omitempty"`
}

// Health handles GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	response := HealthResponse{
		Status:   "healthy",
		Database: "connected",
		Details:  make(map[string]string),
	}

	if err := h.db.HealthCheck(ctx); err != nil {
		h.logger.Error("database health check failed", slog.String("error", err.Error()))
		response.Status = "unhealthy"
		response.Database = "disconnected"
		response.Details["database_error"] = err.Error()
		h.respondJSON(w, http.StatusServiceUnavailable, response)
		return
	}

	h.respondJSON(w, http.StatusOK, response)
}

func (h *HealthHandler) respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if err := json.NewEncoder(w).Encode(data); err != nil {
		h.logger.Error("failed to encode health response", slog.String("error", err.Error()))
	}
}

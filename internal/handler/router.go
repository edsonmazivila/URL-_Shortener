package handler

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// Router creates and configures the HTTP router.
func NewRouter(urlHandler *URLHandler, healthHandler *HealthHandler, logger *slog.Logger) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(LoggingMiddleware(logger))
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/health", healthHandler.Health)

	r.Route("/api", func(r chi.Router) {
		r.Route("/urls", func(r chi.Router) {
			r.Post("/", urlHandler.CreateShortURL)
			r.Get("/", urlHandler.ListURLs)
			r.Get("/{shortCode}", urlHandler.GetURLMetadata)
			r.Delete("/{shortCode}", urlHandler.DeleteURL)
		})
	})

	r.Get("/{shortCode}", urlHandler.RedirectToOriginal)

	return r
}

// LoggingMiddleware logs HTTP requests.
func LoggingMiddleware(logger *slog.Logger) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ww := middleware.NewWrapResponseWriter(w, r.ProtoMajor)

			defer func() {
				logger.Info("http request",
					slog.String("method", r.Method),
					slog.String("path", r.URL.Path),
					slog.String("remote_addr", r.RemoteAddr),
					slog.Int("status", ww.Status()),
					slog.Int("bytes", ww.BytesWritten()),
					slog.Duration("duration", time.Since(start)),
					slog.String("request_id", middleware.GetReqID(r.Context())),
				)
			}()

			next.ServeHTTP(ww, r)
		})
	}
}

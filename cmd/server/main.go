package main

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/edson-mazvila/url-shortener/internal/config"
	"github.com/edson-mazvila/url-shortener/internal/handler"
	"github.com/edson-mazvila/url-shortener/internal/repository"
	"github.com/edson-mazvila/url-shortener/internal/service"
	"github.com/edson-mazvila/url-shortener/internal/storage"
	"github.com/joho/godotenv"
)

func main() {
	if err := godotenv.Load(); err != nil {
		fmt.Fprintf(os.Stderr, "Warning: .env file not found, using environment variables\n")
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	logger := setupLogger(cfg.Logging)

	logger.Info("starting url shortener service",
		slog.String("version", "1.0.0"),
		slog.String("go_version", "1.25"),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	db, err := storage.NewPostgresDB(ctx, &cfg.Database, logger)
	if err != nil {
		logger.Error("failed to connect to database", slog.String("error", err.Error()))
		os.Exit(1)
	}
	defer db.Close()

	urlRepo := repository.NewURLRepository(db.Pool(), logger)
	urlService := service.NewURLService(urlRepo, &cfg.URL, logger)
	urlHandler := handler.NewURLHandler(urlService, logger)
	healthHandler := handler.NewHealthHandler(db, logger)

	router := handler.NewRouter(urlHandler, healthHandler, logger)

	server := &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port),
		Handler:      router,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	go func() {
		logger.Info("server starting",
			slog.String("address", server.Addr),
			slog.String("base_url", cfg.URL.BaseURL),
		)

		if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("server failed", slog.String("error", err.Error()))
			os.Exit(1)
		}
	}()

	go startCleanupWorker(context.Background(), urlService, logger)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	sig := <-quit
	logger.Info("shutdown signal received",
		slog.String("signal", sig.String()),
	)

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), cfg.Server.ShutdownTimeout)
	defer shutdownCancel()

	if err := server.Shutdown(shutdownCtx); err != nil {
		logger.Error("server shutdown failed", slog.String("error", err.Error()))
		os.Exit(1)
	}

	logger.Info("server shutdown completed")
}

func setupLogger(cfg config.LoggingConfig) *slog.Logger {
	var level slog.Level
	switch cfg.Level {
	case "debug":
		level = slog.LevelDebug
	case "info":
		level = slog.LevelInfo
	case "warn":
		level = slog.LevelWarn
	case "error":
		level = slog.LevelError
	default:
		level = slog.LevelInfo
	}

	opts := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	if cfg.Format == "json" {
		handler = slog.NewJSONHandler(os.Stdout, opts)
	} else {
		handler = slog.NewTextHandler(os.Stdout, opts)
	}

	return slog.New(handler)
}

func startCleanupWorker(ctx context.Context, urlService *service.URLService, logger *slog.Logger) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	logger.Info("cleanup worker started")

	for {
		select {
		case <-ctx.Done():
			logger.Info("cleanup worker stopped")
			return
		case <-ticker.C:
			cleanupCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			count, err := urlService.CleanupExpired(cleanupCtx)
			cancel()

			if err != nil {
				logger.Error("cleanup failed", slog.String("error", err.Error()))
			} else if count > 0 {
				logger.Info("cleanup completed", slog.Int64("deleted", count))
			}
		}
	}
}

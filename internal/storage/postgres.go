package storage

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/edson-mazvila/url-shortener/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresDB wraps the PostgreSQL connection pool.
type PostgresDB struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

// NewPostgresDB creates a new PostgreSQL database connection pool.
func NewPostgresDB(ctx context.Context, cfg *config.DatabaseConfig, logger *slog.Logger) (*PostgresDB, error) {
	poolConfig, err := pgxpool.ParseConfig(cfg.GetDSN())
	if err != nil {
		return nil, fmt.Errorf("failed to parse database config: %w", err)
	}

	poolConfig.MaxConns = int32(cfg.MaxOpenConns)
	poolConfig.MinConns = int32(cfg.MaxIdleConns)
	poolConfig.MaxConnLifetime = cfg.ConnMaxLifetime
	poolConfig.MaxConnIdleTime = cfg.ConnMaxIdleTime

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test the connection
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("database connection established",
		slog.String("host", cfg.Host),
		slog.Int("port", cfg.Port),
		slog.String("database", cfg.Database),
	)

	return &PostgresDB{
		pool:   pool,
		logger: logger,
	}, nil
}

// Pool returns the underlying connection pool.
func (db *PostgresDB) Pool() *pgxpool.Pool {
	return db.pool
}

// Close closes the database connection pool.
func (db *PostgresDB) Close() {
	db.logger.Info("closing database connection")
	db.pool.Close()
}

// HealthCheck performs a health check on the database.
func (db *PostgresDB) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()

	if err := db.pool.Ping(ctx); err != nil {
		return fmt.Errorf("database health check failed: %w", err)
	}

	return nil
}

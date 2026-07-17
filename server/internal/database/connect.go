package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect waits for PostgreSQL to become reachable. This lets the same server
// binary work both under Docker Compose and as a native systemd service.
func Connect(ctx context.Context, databaseURL string, logger *slog.Logger) (*pgxpool.Pool, error) {
	poolConfig, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("parse database URL: %w", err)
	}
	poolConfig.MaxConns = 10
	poolConfig.MinConns = 1
	poolConfig.MaxConnLifetime = time.Hour
	poolConfig.MaxConnIdleTime = 15 * time.Minute
	poolConfig.HealthCheckPeriod = 30 * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, fmt.Errorf("create PostgreSQL pool: %w", err)
	}

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for attempt := 1; ; attempt++ {
		if err := pool.Ping(ctx); err == nil {
			logger.Info("PostgreSQL connection ready", "attempt", attempt)
			return pool, nil
		} else if ctx.Err() == nil {
			logger.Warn("waiting for PostgreSQL", "attempt", attempt, "error", err)
		}

		select {
		case <-ctx.Done():
			pool.Close()
			return nil, fmt.Errorf("wait for PostgreSQL: %w", ctx.Err())
		case <-ticker.C:
		}
	}
}

package storage

import (
	"context"
	"fmt"
	"time"

	"github.com/abduss/godrive/internal/config"
	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultDBTimeout = 5 * time.Second

// NewPostgresPool connects to PostgreSQL using pgx.
func NewPostgresPool(ctx context.Context, cfg config.PostgresConfig) (*pgxpool.Pool, error) {
	poolCfg, err := pgxpool.ParseConfig(cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("parse postgres config: %w", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("create postgres pool: %w", err)
	}

	ctx, cancel := context.WithTimeout(ctx, defaultDBTimeout)
	defer cancel()

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}

	return pool, nil
}

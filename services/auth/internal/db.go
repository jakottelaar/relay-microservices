package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jakottelaar/relay-microservices/services/auth/config"
)

func NewPool(ctx context.Context, cfg *config.Config) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(cfg.DB.DatabaseUrl)
	if err != nil {
		return nil, fmt.Errorf("unable to parse database URL: %w", err)
	}

	// Connection pool settings
	config.MaxConns = cfg.DB.MaxConns
	config.MinConns = cfg.DB.MinConns
	config.MaxConnLifetime = time.Duration(cfg.DB.MaxConnLifetime) * time.Second
	config.MaxConnIdleTime = time.Duration(cfg.DB.MaxConnIdleTime) * time.Second

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("unable to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("unable to ping database: %w", err)
	}

	return pool, nil
}
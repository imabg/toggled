// Package postgres provides the pgx-backed flag repository.
package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	// maxConns bounds the pool; plenty for the API tier of a flag service.
	maxConns = 10
	// connectTimeout bounds establishing a single connection.
	connectTimeout = 5 * time.Second
	// pingTimeout bounds the startup connectivity check.
	pingTimeout = 5 * time.Second
)

// NewPool creates a connection pool from url (postgres:// DSN or keyword
// pairs), applies sane defaults, and verifies connectivity with a startup
// ping. The pool is closed if the ping fails.
func NewPool(ctx context.Context, url string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(url)
	if err != nil {
		return nil, fmt.Errorf("postgres: parse url: %w", err)
	}
	cfg.MaxConns = maxConns
	cfg.ConnConfig.ConnectTimeout = connectTimeout

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("postgres: connect: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, pingTimeout)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("postgres: ping: %w", err)
	}
	return pool, nil
}

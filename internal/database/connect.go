package db

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const defaultTimeout = 5 * time.Second

// Store is your app-level DB handle.
// It holds the pool plus the sqlc-generated Queries.
type Store struct {
	Pool    *pgxpool.Pool
	Queries *Queries
}

func NewStore(dsn string) (*Store, error) {
	ctx, cancel := context.WithTimeout(context.Background(), defaultTimeout)
	defer cancel()

	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse db config: %w", err)
	}

	// Pool tuning (closest equivalent to sqlx/*sql.DB settings)
	cfg.MaxConns = 25
	cfg.MinConns = 25
	cfg.MaxConnIdleTime = 5 * time.Minute
	cfg.MaxConnLifetime = 2 * time.Hour

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pgx pool: %w", err)
	}

	// Ensure it's reachable now
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping db: %w", err)
	}

	return &Store{
		Pool:    pool,
		Queries: New(pool), // pool implements DBTX (Exec/Query/QueryRow)
	}, nil
}

func (s *Store) Close() {
	if s != nil && s.Pool != nil {
		s.Pool.Close()
	}
}

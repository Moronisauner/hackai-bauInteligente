// Package db gerencia o pool de conexões com o Postgres (pgxpool).
package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewPool cria um pgxpool.Pool a partir do DSN e valida a conectividade com um
// Ping antes de retornar. Um ping malsucedido é devolvido como erro.
func NewPool(ctx context.Context, dsn string) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("db: parsing dsn: %w", err)
	}
	// Pool modesto para a POC.
	cfg.MaxConns = 10

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("db: creating pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("db: ping failed: %w", err)
	}

	return pool, nil
}

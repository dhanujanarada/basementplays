// Package store contains every SQL statement for the ticketing schema.
// Business rules belong in internal/domain and internal/service; handlers never
// touch SQL directly.
package store

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ErrNotFound is returned when a single-row lookup finds nothing.
var ErrNotFound = errors.New("store: not found")

// Store is the database access layer. It owns no state beyond the pool.
type Store struct {
	pool *pgxpool.Pool
}

// New wraps an existing pool. The pool's lifecycle is owned by the caller.
func New(pool *pgxpool.Pool) *Store {
	return &Store{pool: pool}
}

// InTx runs fn inside a transaction. It commits when fn returns nil and rolls
// back on any error (including a panic, because the deferred rollback runs).
func (s *Store) InTx(ctx context.Context, fn func(pgx.Tx) error) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	if err := fn(tx); err != nil {
		return err
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

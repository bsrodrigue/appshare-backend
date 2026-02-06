package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// TxManager handles database transactions.
// It provides a clean API for executing operations within a transaction.
type TxManager struct {
	pool *pgxpool.Pool
}

// NewTxManager creates a new transaction manager.
func NewTxManager(pool *pgxpool.Pool) *TxManager {
	return &TxManager{pool: pool}
}

// TxFunc is a function that executes within a transaction.
// It receives transaction-aware Queries that will be committed or rolled back together.
type TxFunc func(q *Queries) error

// WithTx executes fn within a database transaction.
//
// If fn returns nil, the transaction is committed.
// If fn returns an error, the transaction is rolled back.
// If fn panics, the transaction is rolled back and the panic is re-raised.
//
// Example:
//
//	err := txManager.WithTx(ctx, func(q *db.Queries) error {
//	    user, err := q.CreateUser(ctx, params)
//	    if err != nil {
//	        return err  // Transaction will be rolled back
//	    }
//	    _, err = q.CreateProject(ctx, projectParams)
//	    return err  // Commit if nil, rollback if error
//	})
func (m *TxManager) WithTx(ctx context.Context, fn TxFunc) error {
	return m.WithTxOptions(ctx, pgx.TxOptions{}, fn)
}

// WithTxOptions executes fn within a transaction with custom options.
// This allows you to specify isolation level, access mode, etc.
func (m *TxManager) WithTxOptions(ctx context.Context, opts pgx.TxOptions, fn TxFunc) (err error) {
	tx, err := m.pool.BeginTx(ctx, opts)
	if err != nil {
		return fmt.Errorf("begin transaction: %w", err)
	}

	// Handle panics - rollback and re-panic
	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback(ctx)
			panic(p)
		}
	}()

	// Create transaction-aware queries
	q := New(tx)

	// Execute the function
	if err = fn(q); err != nil {
		if rbErr := tx.Rollback(ctx); rbErr != nil {
			return fmt.Errorf("rollback failed: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	// Commit on success
	if err = tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit transaction: %w", err)
	}

	return nil
}

// WithSerializableTx executes fn in a SERIALIZABLE isolation level transaction.
// Use this for operations that require the strongest consistency guarantees.
func (m *TxManager) WithSerializableTx(ctx context.Context, fn TxFunc) error {
	return m.WithTxOptions(ctx, pgx.TxOptions{
		IsoLevel: pgx.Serializable,
	}, fn)
}

// WithReadOnlyTx executes fn in a read-only transaction.
// This is useful for complex read operations that need consistency across multiple queries.
func (m *TxManager) WithReadOnlyTx(ctx context.Context, fn TxFunc) error {
	return m.WithTxOptions(ctx, pgx.TxOptions{
		AccessMode: pgx.ReadOnly,
	}, fn)
}

// Pool returns the underlying connection pool.
// Use this only when you need direct pool access.
func (m *TxManager) Pool() *pgxpool.Pool {
	return m.pool
}

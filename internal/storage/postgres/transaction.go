package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// TxFunc is a function that runs within a transaction
type TxFunc func(tx *sql.Tx) error

// WithTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// If the function succeeds, the transaction is committed
func WithTransaction(ctx context.Context, db *sql.DB, fn TxFunc) error {
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			// Rollback on panic (ignore rollback error)
			_ = tx.Rollback()
			panic(p) // Re-throw panic after rollback
		}
	}()

	if err := fn(tx); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx failed: %w", errors.Join(err, rbErr))
		}
		return err
	}

	// Commit the transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithTransactionResult executes a function within a transaction and returns a result
func WithTransactionResult[T any](ctx context.Context, db *sql.DB, fn func(tx *sql.Tx) (T, error)) (T, error) {
	var result T

	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return result, fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	result, err = fn(tx)
	if err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return result, fmt.Errorf("tx failed: %w", errors.Join(err, rbErr))
		}
		return result, err
	}

	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// TxOptions represents transaction options
type TxOptions struct {
	IsolationLevel sql.IsolationLevel
	ReadOnly       bool
}

// WithTransactionOptions executes a function with custom transaction options
func WithTransactionOptions(ctx context.Context, db *sql.DB, opts TxOptions, fn TxFunc) error {
	txOpts := &sql.TxOptions{
		Isolation: opts.IsolationLevel,
		ReadOnly:  opts.ReadOnly,
	}

	tx, err := db.BeginTx(ctx, txOpts)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if p := recover(); p != nil {
			_ = tx.Rollback()
			panic(p)
		}
	}()

	if err := fn(tx); err != nil {
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("tx failed: %w", errors.Join(err, rbErr))
		}
		return err
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// Querier interface for both *sql.DB and *sql.Tx
type Querier interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// Ensure *sql.DB and *sql.Tx implement Querier
var _ Querier = (*sql.DB)(nil)
var _ Querier = (*sql.Tx)(nil)

package database

import (
	"context"
	"database/sql"
	"fmt"
)

// TxKey is the context key for database transactions
type txKeyType string

const TxKey txKeyType = "db_tx"

// TxManager provides transaction management capabilities
type TxManager struct {
	db *sql.DB
}

// NewTxManager creates a new transaction manager
func NewTxManager(db *sql.DB) *TxManager {
	return &TxManager{db: db}
}

// WithTransaction executes a function within a database transaction
// If the function returns an error, the transaction is rolled back
// If the function succeeds, the transaction is committed
func (tm *TxManager) WithTransaction(ctx context.Context, fn func(ctx context.Context) error) error {
	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Store transaction in context
	ctxWithTx := context.WithValue(ctx, TxKey, tx)

	// Execute the function
	if err := fn(ctxWithTx); err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(); rbErr != nil {
			return fmt.Errorf("failed to rollback transaction: %v (original error: %w)", rbErr, err)
		}
		return err
	}

	// Commit on success
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// WithTransactionResult executes a function within a transaction and returns a result
func WithTransactionResult[T any](tm *TxManager, ctx context.Context, fn func(ctx context.Context) (T, error)) (T, error) {
	var result T

	tx, err := tm.db.BeginTx(ctx, nil)
	if err != nil {
		return result, fmt.Errorf("failed to begin transaction: %w", err)
	}

	// Store transaction in context
	ctxWithTx := context.WithValue(ctx, TxKey, tx)

	// Execute the function
	result, err = fn(ctxWithTx)
	if err != nil {
		// Rollback on error
		if rbErr := tx.Rollback(); rbErr != nil {
			return result, fmt.Errorf("failed to rollback transaction: %v (original error: %w)", rbErr, err)
		}
		return result, err
	}

	// Commit on success
	if err := tx.Commit(); err != nil {
		return result, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return result, nil
}

// GetTx retrieves the transaction from context, or nil if not in a transaction
func GetTx(ctx context.Context) *sql.Tx {
	tx, ok := ctx.Value(TxKey).(*sql.Tx)
	if !ok {
		return nil
	}
	return tx
}

// Executor interface for running queries (works with both *sql.DB and *sql.Tx)
type Executor interface {
	ExecContext(ctx context.Context, query string, args ...interface{}) (sql.Result, error)
	QueryContext(ctx context.Context, query string, args ...interface{}) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...interface{}) *sql.Row
}

// GetExecutor returns the transaction if present in context, otherwise returns the db
func GetExecutor(ctx context.Context, db *sql.DB) Executor {
	if tx := GetTx(ctx); tx != nil {
		return tx
	}
	return db
}

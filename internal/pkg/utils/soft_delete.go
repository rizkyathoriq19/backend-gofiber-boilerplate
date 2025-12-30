package utils

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"boilerplate-be/internal/database"
)

// SoftDeleteHelper provides soft delete operations for any table
type SoftDeleteHelper struct {
	db        *sql.DB
	tableName string
}

// NewSoftDeleteHelper creates a new soft delete helper
func NewSoftDeleteHelper(db *sql.DB, tableName string) *SoftDeleteHelper {
	return &SoftDeleteHelper{
		db:        db,
		tableName: tableName,
	}
}

// SoftDelete marks a record as deleted by setting deleted_at
func (h *SoftDeleteHelper) SoftDelete(ctx context.Context, id string) error {
	exec := database.GetExecutor(ctx, h.db)
	query := fmt.Sprintf("UPDATE %s SET deleted_at = $1, updated_at = $1 WHERE id = $2 AND deleted_at IS NULL", h.tableName)
	now := time.Now()
	result, err := exec.ExecContext(ctx, query, now, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// Restore restores a soft-deleted record
func (h *SoftDeleteHelper) Restore(ctx context.Context, id string) error {
	exec := database.GetExecutor(ctx, h.db)
	query := fmt.Sprintf("UPDATE %s SET deleted_at = NULL, updated_at = $1 WHERE id = $2 AND deleted_at IS NOT NULL", h.tableName)
	result, err := exec.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// HardDelete permanently deletes a record
func (h *SoftDeleteHelper) HardDelete(ctx context.Context, id string) error {
	exec := database.GetExecutor(ctx, h.db)
	query := fmt.Sprintf("DELETE FROM %s WHERE id = $1", h.tableName)
	result, err := exec.ExecContext(ctx, query, id)
	if err != nil {
		return err
	}
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return sql.ErrNoRows
	}
	return nil
}

// IsDeleted checks if a record is soft-deleted
func (h *SoftDeleteHelper) IsDeleted(ctx context.Context, id string) (bool, error) {
	query := fmt.Sprintf("SELECT deleted_at FROM %s WHERE id = $1", h.tableName)
	var deletedAt sql.NullTime
	err := h.db.QueryRowContext(ctx, query, id).Scan(&deletedAt)
	if err != nil {
		if err == sql.ErrNoRows {
			return false, nil
		}
		return false, err
	}
	return deletedAt.Valid, nil
}

// PurgeDeleted permanently removes all soft-deleted records older than given days
func (h *SoftDeleteHelper) PurgeDeleted(ctx context.Context, olderThanDays int) (int64, error) {
	exec := database.GetExecutor(ctx, h.db)
	query := fmt.Sprintf("DELETE FROM %s WHERE deleted_at IS NOT NULL AND deleted_at < NOW() - INTERVAL '%d days'", h.tableName, olderThanDays)
	result, err := exec.ExecContext(ctx, query)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

// WithTrashed returns a query modifier to include soft-deleted records
func WithTrashed() string {
	return "" // No filter, include all
}

// OnlyTrashed returns a query modifier to only get soft-deleted records
func OnlyTrashed() string {
	return "AND deleted_at IS NOT NULL"
}

// NotTrashed returns a query modifier to exclude soft-deleted records
func NotTrashed() string {
	return "AND deleted_at IS NULL"
}

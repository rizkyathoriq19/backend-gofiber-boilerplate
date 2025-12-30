package audit

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"boilerplate-be/internal/database"
	"boilerplate-be/internal/pkg/errors"

	"github.com/google/uuid"
)

type auditRepository struct {
	db *sql.DB
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *sql.DB) AuditRepository {
	return &auditRepository{db: db}
}

func (r *auditRepository) Create(ctx context.Context, entry *AuditLogEntry) error {
	exec := database.GetExecutor(ctx, r.db)

	id, _ := uuid.NewV7()

	var oldValuesJSON, newValuesJSON, metadataJSON []byte
	var err error

	if entry.OldValues != nil {
		oldValuesJSON, err = json.Marshal(entry.OldValues)
		if err != nil {
			return errors.Wrap(err, errors.InternalServerError)
		}
	}

	if entry.NewValues != nil {
		newValuesJSON, err = json.Marshal(entry.NewValues)
		if err != nil {
			return errors.Wrap(err, errors.InternalServerError)
		}
	}

	if entry.Metadata != nil {
		metadataJSON, err = json.Marshal(entry.Metadata)
		if err != nil {
			return errors.Wrap(err, errors.InternalServerError)
		}
	}

	query := `
		INSERT INTO audit_logs (id, user_id, action, resource_type, resource_id, 
			old_values, new_values, ip_address, user_agent, request_id, metadata, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
	`

	_, err = exec.ExecContext(ctx, query,
		id.String(),
		entry.UserID,
		entry.Action,
		entry.ResourceType,
		entry.ResourceID,
		oldValuesJSON,
		newValuesJSON,
		entry.IPAddress,
		entry.UserAgent,
		entry.RequestID,
		metadataJSON,
		time.Now(),
	)

	if err != nil {
		return errors.Wrap(err, errors.DatabaseInsertFailed)
	}

	return nil
}

func (r *auditRepository) GetByID(ctx context.Context, id string) (*AuditLog, error) {
	query := `
		SELECT id, user_id, action, resource_type, resource_id, 
			old_values, new_values, ip_address, user_agent, request_id, metadata, created_at
		FROM audit_logs
		WHERE id = $1
	`

	var log AuditLog
	var userID sql.NullString
	var resourceID sql.NullString
	var oldValues, newValues, metadata []byte
	var ipAddress, userAgent, requestID sql.NullString

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&log.ID,
		&userID,
		&log.Action,
		&log.ResourceType,
		&resourceID,
		&oldValues,
		&newValues,
		&ipAddress,
		&userAgent,
		&requestID,
		&metadata,
		&log.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.ResourceNotFound)
		}
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}

	if userID.Valid {
		log.UserID = &userID.String
	}
	log.ResourceID = resourceID.String
	log.IPAddress = ipAddress.String
	log.UserAgent = userAgent.String
	log.RequestID = requestID.String

	if oldValues != nil {
		json.Unmarshal(oldValues, &log.OldValues)
	}
	if newValues != nil {
		json.Unmarshal(newValues, &log.NewValues)
	}
	if metadata != nil {
		json.Unmarshal(metadata, &log.Metadata)
	}

	return &log, nil
}

func (r *auditRepository) List(ctx context.Context, filter AuditLogFilter) ([]AuditLog, int64, error) {
	var conditions []string
	var args []interface{}
	argIndex := 1

	if filter.UserID != "" {
		conditions = append(conditions, fmt.Sprintf("user_id = $%d", argIndex))
		args = append(args, filter.UserID)
		argIndex++
	}

	if filter.Action != "" {
		conditions = append(conditions, fmt.Sprintf("action = $%d", argIndex))
		args = append(args, filter.Action)
		argIndex++
	}

	if filter.ResourceType != "" {
		conditions = append(conditions, fmt.Sprintf("resource_type = $%d", argIndex))
		args = append(args, filter.ResourceType)
		argIndex++
	}

	if filter.ResourceID != "" {
		conditions = append(conditions, fmt.Sprintf("resource_id = $%d", argIndex))
		args = append(args, filter.ResourceID)
		argIndex++
	}

	if filter.StartDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at >= $%d", argIndex))
		args = append(args, filter.StartDate)
		argIndex++
	}

	if filter.EndDate != nil {
		conditions = append(conditions, fmt.Sprintf("created_at <= $%d", argIndex))
		args = append(args, filter.EndDate)
		argIndex++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM audit_logs %s", whereClause)
	var total int64
	err := r.db.QueryRowContext(ctx, countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}

	// Set default pagination
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PageSize < 1 {
		filter.PageSize = 20
	}
	offset := (filter.Page - 1) * filter.PageSize

	// Get logs
	query := fmt.Sprintf(`
		SELECT id, user_id, action, resource_type, resource_id, 
			old_values, new_values, ip_address, user_agent, request_id, metadata, created_at
		FROM audit_logs
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d
	`, whereClause, argIndex, argIndex+1)

	args = append(args, filter.PageSize, offset)

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	var logs []AuditLog
	for rows.Next() {
		var log AuditLog
		var userID sql.NullString
		var resourceID sql.NullString
		var oldValues, newValues, metadata []byte
		var ipAddress, userAgent, requestID sql.NullString

		err := rows.Scan(
			&log.ID,
			&userID,
			&log.Action,
			&log.ResourceType,
			&resourceID,
			&oldValues,
			&newValues,
			&ipAddress,
			&userAgent,
			&requestID,
			&metadata,
			&log.CreatedAt,
		)
		if err != nil {
			return nil, 0, errors.Wrap(err, errors.DatabaseScanFailed)
		}

		if userID.Valid {
			log.UserID = &userID.String
		}
		log.ResourceID = resourceID.String
		log.IPAddress = ipAddress.String
		log.UserAgent = userAgent.String
		log.RequestID = requestID.String

		if oldValues != nil {
			json.Unmarshal(oldValues, &log.OldValues)
		}
		if newValues != nil {
			json.Unmarshal(newValues, &log.NewValues)
		}
		if metadata != nil {
			json.Unmarshal(metadata, &log.Metadata)
		}

		logs = append(logs, log)
	}

	return logs, total, nil
}

func (r *auditRepository) DeleteOlderThan(ctx context.Context, days int) (int64, error) {
	query := `DELETE FROM audit_logs WHERE created_at < NOW() - INTERVAL '1 day' * $1`

	result, err := r.db.ExecContext(ctx, query, days)
	if err != nil {
		return 0, errors.Wrap(err, errors.DatabaseDeleteFailed)
	}

	return result.RowsAffected()
}

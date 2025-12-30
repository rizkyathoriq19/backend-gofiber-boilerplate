package upload

import (
	"context"
	"database/sql"
	"encoding/json"
	"time"

	"boilerplate-be/internal/database"
	"boilerplate-be/internal/pkg/errors"

	"github.com/google/uuid"
)

type fileRepository struct {
	db *sql.DB
}

// NewFileRepository creates a new file repository
func NewFileRepository(db *sql.DB) FileRepository {
	return &fileRepository{db: db}
}

func (r *fileRepository) Create(ctx context.Context, file *FileUpload) error {
	exec := database.GetExecutor(ctx, r.db)

	if file.ID == "" {
		id, _ := uuid.NewV7()
		file.ID = id.String()
	}
	file.CreatedAt = time.Now()
	file.UpdatedAt = time.Now()

	var metadataJSON []byte
	var err error
	if file.Metadata != nil {
		metadataJSON, err = json.Marshal(file.Metadata)
		if err != nil {
			return errors.Wrap(err, errors.InternalServerError)
		}
	}

	query := `
		INSERT INTO file_uploads (id, user_id, original_name, stored_name, file_path, file_size, 
			mime_type, extension, storage_type, bucket_name, is_public, metadata, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
	`

	_, err = exec.ExecContext(ctx, query,
		file.ID,
		file.UserID,
		file.OriginalName,
		file.StoredName,
		file.FilePath,
		file.FileSize,
		file.MimeType,
		file.Extension,
		file.StorageType,
		file.BucketName,
		file.IsPublic,
		metadataJSON,
		file.CreatedAt,
		file.UpdatedAt,
	)

	if err != nil {
		return errors.Wrap(err, errors.DatabaseInsertFailed)
	}

	return nil
}

func (r *fileRepository) GetByID(ctx context.Context, id string) (*FileUpload, error) {
	query := `
		SELECT id, user_id, original_name, stored_name, file_path, file_size, 
			mime_type, extension, storage_type, bucket_name, is_public, metadata, 
			created_at, updated_at, deleted_at
		FROM file_uploads
		WHERE id = $1 AND deleted_at IS NULL
	`

	var file FileUpload
	var userID sql.NullString
	var bucketName sql.NullString
	var extension sql.NullString
	var metadata []byte
	var deletedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, id).Scan(
		&file.ID,
		&userID,
		&file.OriginalName,
		&file.StoredName,
		&file.FilePath,
		&file.FileSize,
		&file.MimeType,
		&extension,
		&file.StorageType,
		&bucketName,
		&file.IsPublic,
		&metadata,
		&file.CreatedAt,
		&file.UpdatedAt,
		&deletedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New(errors.ResourceNotFound)
		}
		return nil, errors.Wrap(err, errors.DatabaseQueryFailed)
	}

	if userID.Valid {
		file.UserID = &userID.String
	}
	file.BucketName = bucketName.String
	file.Extension = extension.String
	if deletedAt.Valid {
		file.DeletedAt = &deletedAt.Time
	}
	if metadata != nil {
		json.Unmarshal(metadata, &file.Metadata)
	}

	return &file, nil
}

func (r *fileRepository) GetByUserID(ctx context.Context, userID string, page, pageSize int) ([]FileUpload, int64, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize

	// Count total
	countQuery := `SELECT COUNT(*) FROM file_uploads WHERE user_id = $1 AND deleted_at IS NULL`
	var total int64
	if err := r.db.QueryRowContext(ctx, countQuery, userID).Scan(&total); err != nil {
		return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}

	// Get files
	query := `
		SELECT id, user_id, original_name, stored_name, file_path, file_size, 
			mime_type, extension, storage_type, bucket_name, is_public, metadata, 
			created_at, updated_at
		FROM file_uploads
		WHERE user_id = $1 AND deleted_at IS NULL
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`

	rows, err := r.db.QueryContext(ctx, query, userID, pageSize, offset)
	if err != nil {
		return nil, 0, errors.Wrap(err, errors.DatabaseQueryFailed)
	}
	defer rows.Close()

	var files []FileUpload
	for rows.Next() {
		var file FileUpload
		var uid sql.NullString
		var bucketName sql.NullString
		var extension sql.NullString
		var metadata []byte

		err := rows.Scan(
			&file.ID,
			&uid,
			&file.OriginalName,
			&file.StoredName,
			&file.FilePath,
			&file.FileSize,
			&file.MimeType,
			&extension,
			&file.StorageType,
			&bucketName,
			&file.IsPublic,
			&metadata,
			&file.CreatedAt,
			&file.UpdatedAt,
		)
		if err != nil {
			return nil, 0, errors.Wrap(err, errors.DatabaseScanFailed)
		}

		if uid.Valid {
			file.UserID = &uid.String
		}
		file.BucketName = bucketName.String
		file.Extension = extension.String
		if metadata != nil {
			json.Unmarshal(metadata, &file.Metadata)
		}

		files = append(files, file)
	}

	return files, total, nil
}

func (r *fileRepository) Delete(ctx context.Context, id string) error {
	query := `DELETE FROM file_uploads WHERE id = $1`
	result, err := r.db.ExecContext(ctx, query, id)
	if err != nil {
		return errors.Wrap(err, errors.DatabaseDeleteFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	return nil
}

func (r *fileRepository) SoftDelete(ctx context.Context, id string) error {
	query := `UPDATE file_uploads SET deleted_at = $2 WHERE id = $1 AND deleted_at IS NULL`
	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	return nil
}

func (r *fileRepository) Restore(ctx context.Context, id string) error {
	query := `UPDATE file_uploads SET deleted_at = NULL, updated_at = $2 WHERE id = $1 AND deleted_at IS NOT NULL`
	result, err := r.db.ExecContext(ctx, query, id, time.Now())
	if err != nil {
		return errors.Wrap(err, errors.DatabaseUpdateFailed)
	}

	rowsAffected, _ := result.RowsAffected()
	if rowsAffected == 0 {
		return errors.New(errors.ResourceNotFound)
	}

	return nil
}

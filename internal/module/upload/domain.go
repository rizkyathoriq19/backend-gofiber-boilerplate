package upload

import (
	"context"
	"io"
	"mime/multipart"
)

// FileStorage defines the interface for file storage operations
type FileStorage interface {
	Upload(ctx context.Context, file io.Reader, filename string, opts *UploadOptions) (*FileUpload, error)
	Delete(ctx context.Context, filePath string) error
	GetURL(ctx context.Context, filePath string) (string, error)
	Exists(ctx context.Context, filePath string) (bool, error)
}

// FileRepository defines the data access layer for file uploads
type FileRepository interface {
	Create(ctx context.Context, file *FileUpload) error
	GetByID(ctx context.Context, id string) (*FileUpload, error)
	GetByUserID(ctx context.Context, userID string, page, pageSize int) ([]FileUpload, int64, error)
	Delete(ctx context.Context, id string) error
	SoftDelete(ctx context.Context, id string) error
	Restore(ctx context.Context, id string) error
}

// FileUseCase defines the business logic for file uploads
type FileUseCase interface {
	Upload(ctx context.Context, userID *string, fileHeader *multipart.FileHeader, opts *UploadOptions) (*FileUpload, error)
	GetByID(ctx context.Context, id string) (*FileUpload, error)
	GetByUserID(ctx context.Context, userID string, page, pageSize int) ([]FileUpload, int64, error)
	Delete(ctx context.Context, id string, hardDelete bool) error
	GetURL(ctx context.Context, id string) (string, error)
}

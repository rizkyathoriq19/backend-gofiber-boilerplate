package upload

import (
	"context"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"slices"

	"boilerplate-be/internal/pkg/errors"
)

type fileUseCase struct {
	repo    FileRepository
	storage FileStorage
}

// NewFileUseCase creates a new file use case
func NewFileUseCase(repo FileRepository, storage FileStorage) FileUseCase {
	return &fileUseCase{
		repo:    repo,
		storage: storage,
	}
}

func (uc *fileUseCase) Upload(ctx context.Context, userID *string, fileHeader *multipart.FileHeader, opts *UploadOptions) (*FileUpload, error) {
	if opts == nil {
		opts = DefaultUploadOptions()
	}

	// Validate file size
	if fileHeader.Size > opts.MaxSize {
		return nil, errors.New(errors.FileSizeExceeded)
	}

	// Open the file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, errors.Wrap(err, errors.InternalServerError)
	}
	defer file.Close()

	// Detect MIME type
	buffer := make([]byte, 512)
	_, err = file.Read(buffer)
	if err != nil {
		return nil, errors.Wrap(err, errors.InternalServerError)
	}
	mimeType := http.DetectContentType(buffer)

	// Reset file pointer
	file.Seek(0, 0)

	// Validate MIME type
	if len(opts.AllowedTypes) > 0 && !slices.Contains(opts.AllowedTypes, mimeType) {
		return nil, errors.New(errors.InvalidFileType)
	}

	// Upload to storage
	upload, err := uc.storage.Upload(ctx, file, fileHeader.Filename, opts)
	if err != nil {
		return nil, errors.Wrap(err, errors.InternalServerError)
	}

	// Set additional fields
	upload.UserID = userID
	upload.MimeType = mimeType
	upload.Extension = filepath.Ext(fileHeader.Filename)

	// Save to database
	if err := uc.repo.Create(ctx, upload); err != nil {
		// Rollback: delete uploaded file
		uc.storage.Delete(ctx, upload.FilePath)
		return nil, err
	}

	return upload, nil
}

func (uc *fileUseCase) GetByID(ctx context.Context, id string) (*FileUpload, error) {
	return uc.repo.GetByID(ctx, id)
}

func (uc *fileUseCase) GetByUserID(ctx context.Context, userID string, page, pageSize int) ([]FileUpload, int64, error) {
	return uc.repo.GetByUserID(ctx, userID, page, pageSize)
}

func (uc *fileUseCase) Delete(ctx context.Context, id string, hardDelete bool) error {
	// Get file first
	file, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if hardDelete {
		// Delete from storage
		if err := uc.storage.Delete(ctx, file.FilePath); err != nil {
			return errors.Wrap(err, errors.InternalServerError)
		}
		// Delete from database
		return uc.repo.Delete(ctx, id)
	}

	// Soft delete
	return uc.repo.SoftDelete(ctx, id)
}

func (uc *fileUseCase) GetURL(ctx context.Context, id string) (string, error) {
	file, err := uc.repo.GetByID(ctx, id)
	if err != nil {
		return "", err
	}

	return uc.storage.GetURL(ctx, file.FilePath)
}

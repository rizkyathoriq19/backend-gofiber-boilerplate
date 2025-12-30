package upload

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
)

// LocalStorage implements FileStorage for local filesystem
type LocalStorage struct {
	basePath string
	baseURL  string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(basePath, baseURL string) *LocalStorage {
	return &LocalStorage{
		basePath: basePath,
		baseURL:  baseURL,
	}
}

func (s *LocalStorage) Upload(ctx context.Context, file io.Reader, filename string, opts *UploadOptions) (*FileUpload, error) {
	// Generate unique filename
	ext := filepath.Ext(filename)
	id, _ := uuid.NewV7()
	storedName := fmt.Sprintf("%s%s", id.String(), ext)

	// Create directory if not exists
	dir := filepath.Join(s.basePath, opts.Directory)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create directory: %w", err)
	}

	// Create file path
	filePath := filepath.Join(opts.Directory, storedName)
	fullPath := filepath.Join(s.basePath, filePath)

	// Create destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	size, err := io.Copy(dst, file)
	if err != nil {
		os.Remove(fullPath) // Cleanup on error
		return nil, fmt.Errorf("failed to save file: %w", err)
	}

	return &FileUpload{
		ID:           id.String(),
		OriginalName: filename,
		StoredName:   storedName,
		FilePath:     filePath,
		FileSize:     size,
		Extension:    ext,
		StorageType:  StorageLocal,
		IsPublic:     opts.IsPublic,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}, nil
}

func (s *LocalStorage) Delete(ctx context.Context, filePath string) error {
	fullPath := filepath.Join(s.basePath, filePath)
	if err := os.Remove(fullPath); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

func (s *LocalStorage) GetURL(ctx context.Context, filePath string) (string, error) {
	return fmt.Sprintf("%s/%s", s.baseURL, filePath), nil
}

func (s *LocalStorage) Exists(ctx context.Context, filePath string) (bool, error) {
	fullPath := filepath.Join(s.basePath, filePath)
	_, err := os.Stat(fullPath)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, nil
}

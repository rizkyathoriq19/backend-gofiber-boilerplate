package upload

import (
	"time"
)

// StorageType represents the type of file storage
type StorageType string

const (
	StorageLocal StorageType = "local"
	StorageS3    StorageType = "s3"
	StorageMinio StorageType = "minio"
)

// FileUpload represents a file upload record
type FileUpload struct {
	ID           string                 `json:"id"`
	UserID       *string                `json:"user_id,omitempty"`
	OriginalName string                 `json:"original_name"`
	StoredName   string                 `json:"stored_name"`
	FilePath     string                 `json:"file_path"`
	FileSize     int64                  `json:"file_size"`
	MimeType     string                 `json:"mime_type"`
	Extension    string                 `json:"extension"`
	StorageType  StorageType            `json:"storage_type"`
	BucketName   string                 `json:"bucket_name,omitempty"`
	IsPublic     bool                   `json:"is_public"`
	Metadata     map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt    time.Time              `json:"created_at"`
	UpdatedAt    time.Time              `json:"updated_at"`
	DeletedAt    *time.Time             `json:"deleted_at,omitempty"`
}

// UploadOptions contains options for file upload
type UploadOptions struct {
	MaxSize      int64    // Maximum file size in bytes
	AllowedTypes []string // Allowed MIME types
	IsPublic     bool     // Whether the file should be publicly accessible
	Directory    string   // Subdirectory to store the file
}

// DefaultUploadOptions returns default upload options
func DefaultUploadOptions() *UploadOptions {
	return &UploadOptions{
		MaxSize:      10 * 1024 * 1024, // 10MB
		AllowedTypes: []string{"image/jpeg", "image/png", "image/gif", "application/pdf"},
		IsPublic:     false,
		Directory:    "uploads",
	}
}

// ImageUploadOptions returns options for image uploads
func ImageUploadOptions() *UploadOptions {
	return &UploadOptions{
		MaxSize:      5 * 1024 * 1024, // 5MB
		AllowedTypes: []string{"image/jpeg", "image/png", "image/gif", "image/webp"},
		IsPublic:     true,
		Directory:    "images",
	}
}

// DocumentUploadOptions returns options for document uploads
func DocumentUploadOptions() *UploadOptions {
	return &UploadOptions{
		MaxSize:      20 * 1024 * 1024, // 20MB
		AllowedTypes: []string{"application/pdf", "application/msword", "application/vnd.openxmlformats-officedocument.wordprocessingml.document", "application/vnd.ms-excel", "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet"},
		IsPublic:     false,
		Directory:    "documents",
	}
}

package storage

import (
	"context"
	"time"
)

// ObjectStorage defines the interface for cloud object storage operations
type ObjectStorage interface {
	// GenerateUploadURL creates a signed URL for uploading a file
	GenerateUploadURL(ctx context.Context, path string, contentType string, expiresIn time.Duration) (string, error)

	// GenerateDownloadURL creates a signed URL for downloading a file
	GenerateDownloadURL(ctx context.Context, path string, expiresIn time.Duration) (string, error)

	// Delete removes an object from storage
	Delete(ctx context.Context, path string) error

	// Exists checks if an object exists
	Exists(ctx context.Context, path string) (bool, error)
}

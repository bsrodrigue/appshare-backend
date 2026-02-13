package storage

import (
	"context"
	"io"
	"time"
)

// Storage defines the interface for file storage operations.
type Storage interface {
	// GenerateUploadURL generates a signed URL for uploading a file to the given path.
	// This usually returns a PUT URL that the client can use to upload the file directly.
	GenerateUploadURL(ctx context.Context, path string, expires time.Duration) (string, error)

	// Delete deletes a file from the given path.
	Delete(ctx context.Context, path string) error

	// GetPublicURL returns the public URL for a file at the given path.
	GetPublicURL(path string) string

	// Download returns a reader for the file at the given path.
	Download(ctx context.Context, path string) (io.ReadCloser, error)

	// ExtractStoragePath extracts the storage path from a URL.
	ExtractStoragePath(url string) (string, bool)
}

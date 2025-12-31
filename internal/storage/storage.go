package storage

import (
	"context"
	"io"
)

// FileStorage defines the interface for file storage operations
type FileStorage interface {
	// SaveFile saves a file and returns its public URL
	SaveFile(ctx context.Context, file io.Reader, filename string, contentType string) (string, error)
	// DeleteFile deletes a file by its URL
	DeleteFile(ctx context.Context, fileURL string) error
}

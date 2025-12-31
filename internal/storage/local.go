package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// LocalFileStorage implements FileStorage for local filesystem
type LocalFileStorage struct {
	basePath string
	baseURL  string
}

// NewLocalFileStorage creates a new local file storage
func NewLocalFileStorage(basePath, baseURL string) (*LocalFileStorage, error) {
	// Ensure directory exists
	if err := os.MkdirAll(basePath, 0755); err != nil {
		return nil, fmt.Errorf("failed to create storage directory: %w", err)
	}

	return &LocalFileStorage{
		basePath: basePath,
		baseURL:  strings.TrimRight(baseURL, "/"),
	}, nil
}

// SaveFile saves a file to local disk
func (s *LocalFileStorage) SaveFile(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	// Generate unique filename to prevent collisions
	ext := filepath.Ext(filename)
	if ext == "" {
		// Try to guess from content type (simplified)
		chunks := strings.Split(contentType, "/")
		if len(chunks) == 2 {
			ext = "." + chunks[1]
		}
	}

	newFilename := fmt.Sprintf("%s_%s%s", time.Now().Format("20060102"), uuid.New().String(), ext)
	fullPath := filepath.Join(s.basePath, newFilename)

	// Create file
	dst, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file on disk: %w", err)
	}
	defer dst.Close()

	// Copy content
	if _, err := io.Copy(dst, file); err != nil {
		return "", fmt.Errorf("failed to save file content: %w", err)
	}

	// Return public URL
	return fmt.Sprintf("%s/%s", s.baseURL, newFilename), nil
}

// DeleteFile deletes a file from local disk
func (s *LocalFileStorage) DeleteFile(ctx context.Context, fileURL string) error {
	// Extract filename from URL
	parts := strings.Split(fileURL, "/")
	filename := parts[len(parts)-1]

	fullPath := filepath.Join(s.basePath, filename)

	// Check if file exists within base path to prevent traversal (basic check)
	// In production, should be more robust

	// Check if exists
	if _, err := os.Stat(fullPath); os.IsNotExist(err) {
		return nil // Already gone
	}

	if err := os.Remove(fullPath); err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}

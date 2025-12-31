package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
	internalConfig "github.com/locolive/backend/internal/config"
)

type S3Storage struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

// NewS3Storage creates a new S3/R2 storage provider
func NewS3Storage(ctx context.Context, cfg internalConfig.StorageConfig) (*S3Storage, error) {
	// Create a custom endpoint resolver for R2
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: cfg.Endpoint,
		}, nil
	})

	awsCfg, err := config.LoadDefaultConfig(ctx,
		config.WithRegion(cfg.Region),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(cfg.AccessKeyID, cfg.SecretAccessKey, "")),
		config.WithEndpointResolverWithOptions(r2Resolver),
	)
	if err != nil {
		return nil, fmt.Errorf("unable to load SDK config, %v", err)
	}

	client := s3.NewFromConfig(awsCfg)

	return &S3Storage{
		client:    client,
		bucket:    cfg.Bucket,
		publicURL: cfg.PublicURL,
	}, nil
}

// SaveFile uploads a file to R2/S3
func (s *S3Storage) SaveFile(ctx context.Context, file io.Reader, filename string, contentType string) (string, error) {
	// Generate a unique filename to prevent collisions
	ext := filepath.Ext(filename)
	uniqueName := fmt.Sprintf("%s%s", uuid.New().String(), ext)

	// In a real app, you might want to organize by date or type, e.g., "stories/YYYY/MM/DD/uuid.ext"
	key := fmt.Sprintf("uploads/%s", uniqueName)

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        file,
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Construct public URL
	// If PublicURL is set (e.g., custom domain), use it.
	// Otherwise, R2 public bucket URL format is usually like https://pub-<hash>.r2.dev/<key>
	if s.publicURL != "" {
		// Ensure trailing slash handling if needed, but simple concatenation is usually fine if configured correctly
		return fmt.Sprintf("%s/%s", s.publicURL, key), nil
	}

	// Fallback/Warning: This might not work if not configured, but returns the key for reference
	return key, nil
}

// DeleteFile deletes a file from S3
func (s *S3Storage) DeleteFile(ctx context.Context, fileURL string) error {
	// Simple extraction of key from URL.
	// This assumes fileURL contains the key at the end.
	// A better approach depends on exact URL structure.
	// For now, let's assume valid key is passed or extracted manually if URL is full.

	// TODO: Robust key extraction
	key := fileURL

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}
	return nil
}

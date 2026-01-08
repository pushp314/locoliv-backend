package storage

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// R2Config holds Cloudflare R2 configuration
type R2Config struct {
	AccountID       string
	AccessKeyID     string
	AccessKeySecret string
	BucketName      string
	PublicURL       string // CDN URL like https://cdn.locolive.app or R2 public bucket URL
}

// R2Storage implements FileStorage for Cloudflare R2
type R2Storage struct {
	client    *s3.Client
	bucket    string
	publicURL string
}

// NewR2Storage creates a new R2 storage client
func NewR2Storage(cfg R2Config) (*R2Storage, error) {
	// R2 endpoint: https://<account_id>.r2.cloudflarestorage.com
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", cfg.AccountID)

	// Create custom resolver for R2
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL:               endpoint,
			SigningRegion:     "auto",
			HostnameImmutable: true,
		}, nil
	})

	// Load AWS config with R2 credentials
	awsCfg, err := config.LoadDefaultConfig(context.Background(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
			cfg.AccessKeyID,
			cfg.AccessKeySecret,
			"",
		)),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load R2 config: %w", err)
	}

	// Create S3 client (R2 is S3-compatible)
	client := s3.NewFromConfig(awsCfg, func(o *s3.Options) {
		o.UsePathStyle = true
	})

	return &R2Storage{
		client:    client,
		bucket:    cfg.BucketName,
		publicURL: strings.TrimSuffix(cfg.PublicURL, "/"),
	}, nil
}

// SaveFile uploads a file to R2 and returns the public URL
func (s *R2Storage) SaveFile(ctx context.Context, file io.Reader, filename, contentType string) (string, error) {
	// Generate unique key with date-based folder structure
	now := time.Now()
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = getExtFromContentType(contentType)
	}

	key := fmt.Sprintf("%d/%02d/%02d/%s%s",
		now.Year(), now.Month(), now.Day(),
		uuid.New().String(), ext)

	// Upload to R2
	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:       aws.String(s.bucket),
		Key:          aws.String(key),
		Body:         file,
		ContentType:  aws.String(contentType),
		CacheControl: aws.String("public, max-age=31536000"), // 1 year cache
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to R2: %w", err)
	}

	// Return public CDN URL
	return fmt.Sprintf("%s/%s", s.publicURL, key), nil
}

// DeleteFile removes a file from R2
func (s *R2Storage) DeleteFile(ctx context.Context, url string) error {
	// Extract key from URL
	key := strings.TrimPrefix(url, s.publicURL+"/")
	if key == url {
		// URL doesn't match our pattern, might be external
		return nil
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	return err
}

// GetSignedURL generates a presigned URL for private access
func (s *R2Storage) GetSignedURL(ctx context.Context, key string, expiry time.Duration) (string, error) {
	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignGetObject(ctx, &s3.GetObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	}, s3.WithPresignExpires(expiry))

	if err != nil {
		return "", fmt.Errorf("failed to create presigned URL: %w", err)
	}

	return request.URL, nil
}

// GenerateUploadURL creates a presigned URL for direct client upload
func (s *R2Storage) GenerateUploadURL(ctx context.Context, filename, contentType string, expiry time.Duration) (string, string, error) {
	now := time.Now()
	ext := filepath.Ext(filename)
	if ext == "" {
		ext = getExtFromContentType(contentType)
	}

	key := fmt.Sprintf("%d/%02d/%02d/%s%s",
		now.Year(), now.Month(), now.Day(),
		uuid.New().String(), ext)

	presignClient := s3.NewPresignClient(s.client)

	request, err := presignClient.PresignPutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		ContentType: aws.String(contentType),
	}, s3.WithPresignExpires(expiry))

	if err != nil {
		return "", "", fmt.Errorf("failed to create presigned upload URL: %w", err)
	}

	publicURL := fmt.Sprintf("%s/%s", s.publicURL, key)
	return request.URL, publicURL, nil
}

// ListFiles lists files in a prefix
func (s *R2Storage) ListFiles(ctx context.Context, prefix string, maxKeys int) ([]string, error) {
	result, err := s.client.ListObjectsV2(ctx, &s3.ListObjectsV2Input{
		Bucket:  aws.String(s.bucket),
		Prefix:  aws.String(prefix),
		MaxKeys: aws.Int32(int32(maxKeys)),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	urls := make([]string, 0, len(result.Contents))
	for _, obj := range result.Contents {
		urls = append(urls, fmt.Sprintf("%s/%s", s.publicURL, *obj.Key))
	}
	return urls, nil
}

// getExtFromContentType returns file extension from content type
func getExtFromContentType(contentType string) string {
	switch contentType {
	case "image/jpeg":
		return ".jpg"
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "video/mp4":
		return ".mp4"
	case "video/webm":
		return ".webm"
	case "audio/mpeg":
		return ".mp3"
	case "audio/wav":
		return ".wav"
	default:
		return ""
	}
}

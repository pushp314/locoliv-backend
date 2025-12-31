package domain

import (
	"context"
	"io"
	"time"

	"github.com/locolive/backend/internal/storage"
)

type StoryService struct {
	repo    StoryRepository
	storage storage.FileStorage
}

func NewStoryService(repo StoryRepository, storage storage.FileStorage) *StoryService {
	return &StoryService{
		repo:    repo,
		storage: storage,
	}
}

func (s *StoryService) CreateStory(ctx context.Context, params CreateStoryParams, file io.Reader, filename, contentType string) (*Story, error) {
	// Upload file
	url, err := s.storage.SaveFile(ctx, file, filename, contentType)
	if err != nil {
		return nil, err
	}
	params.MediaURL = url

	// Set default expiry to 24 hours if not set
	if params.ExpiresAt.IsZero() {
		params.ExpiresAt = time.Now().Add(24 * time.Hour)
	}
	return s.repo.CreateStory(ctx, params)
}

func (s *StoryService) GetFeed(ctx context.Context, page, limit int, lat, lng, radius *float64) ([]*Story, error) {
	if limit <= 0 {
		limit = 10
	}
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	if lat != nil && lng != nil && radius != nil {
		return s.repo.GetStoriesByLocation(ctx, *lat, *lng, *radius, limit, offset)
	}

	return s.repo.GetActiveStories(ctx, limit, offset)
}

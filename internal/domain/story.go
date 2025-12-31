package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Story struct {
	ID          uuid.UUID     `json:"id"`
	UserID      uuid.UUID     `json:"user_id"`
	MediaURL    string        `json:"media_url"`
	MediaType   string        `json:"media_type"` // "image" or "video"
	Caption     *string       `json:"caption,omitempty"`
	LocationLat *float64      `json:"location_lat,omitempty"`
	LocationLng *float64      `json:"location_lng,omitempty"`
	ExpiresAt   time.Time     `json:"expires_at"`
	CreatedAt   time.Time     `json:"created_at"`
	User        *UserResponse `json:"user,omitempty"` // For feed response
}

type CreateStoryParams struct {
	UserID      uuid.UUID
	MediaURL    string
	MediaType   string
	Caption     *string
	LocationLat *float64
	LocationLng *float64
	ExpiresAt   time.Time // Calculated by service usually
}

type StoryRepository interface {
	CreateStory(ctx context.Context, params CreateStoryParams) (*Story, error)
	GetActiveStories(ctx context.Context, limit, offset int) ([]*Story, error)
	GetStoriesByLocation(ctx context.Context, lat, lng, radius float64, limit, offset int) ([]*Story, error)
	DeleteExpiredStories(ctx context.Context) (int64, error)
}

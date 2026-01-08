package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// NearbyUser represents a user in discovery results
type NearbyUser struct {
	ID               uuid.UUID  `json:"id"`
	Name             string     `json:"name"`
	AvatarURL        *string    `json:"avatar_url,omitempty"`
	City             *string    `json:"city,omitempty"`
	DistanceKm       float64    `json:"distance_km"`
	TotalKarma       int        `json:"total_karma"`
	CurrentTier      string     `json:"current_tier"`
	IsVerified       bool       `json:"is_verified"`
	VerificationType *string    `json:"verification_type,omitempty"`
	CommonInterests  []string   `json:"common_interests,omitempty"`
	LastActiveAt     *time.Time `json:"last_active_at,omitempty"`
}

// NearbyStory represents a story in discovery results
type NearbyStory struct {
	ID         uuid.UUID `json:"id"`
	UserID     uuid.UUID `json:"user_id"`
	UserName   string    `json:"user_name"`
	UserAvatar *string   `json:"user_avatar,omitempty"`
	MediaURL   string    `json:"media_url"`
	MediaType  string    `json:"media_type"`
	Caption    *string   `json:"caption,omitempty"`
	City       *string   `json:"city,omitempty"`
	DistanceKm float64   `json:"distance_km"`
	CreatedAt  time.Time `json:"created_at"`
	ViewCount  int       `json:"view_count"`
}

// NearbyEvent represents an event in discovery results
type NearbyEvent struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description *string   `json:"description,omitempty"`
	EventDate   time.Time `json:"event_date"`
	Location    string    `json:"location"`
	City        string    `json:"city"`
	DistanceKm  float64   `json:"distance_km"`
	RSVPCount   int       `json:"rsvp_count"`
	CreatorName string    `json:"creator_name"`
}

// ExploreData is aggregated data for the explore screen
type ExploreData struct {
	NearbyUsers    []NearbyUser  `json:"nearby_users"`
	NearbyStories  []NearbyStory `json:"nearby_stories"`
	NearbyEvents   []NearbyEvent `json:"nearby_events"`
	TrendingTopics []string      `json:"trending_topics"`
	SuggestedUsers []NearbyUser  `json:"suggested_users"`
	LocalStats     LocalStats    `json:"local_stats"`
}

// LocalStats shows statistics for the user's area
type LocalStats struct {
	TotalUsers     int `json:"total_users"`
	ActiveToday    int `json:"active_today"`
	NewStories     int `json:"new_stories"`
	UpcomingEvents int `json:"upcoming_events"`
}

// UserProfile is a public user profile
type UserProfile struct {
	ID               uuid.UUID `json:"id"`
	Name             string    `json:"name"`
	Bio              *string   `json:"bio,omitempty"`
	AvatarURL        *string   `json:"avatar_url,omitempty"`
	City             *string   `json:"city,omitempty"`
	TotalKarma       int       `json:"total_karma"`
	CurrentTier      string    `json:"current_tier"`
	TierName         string    `json:"tier_name"`
	IsVerified       bool      `json:"is_verified"`
	VerificationType *string   `json:"verification_type,omitempty"`
	StoryCount       int       `json:"story_count"`
	ConnectionCount  int       `json:"connection_count"`
	BadgeCount       int       `json:"badge_count"`
	JoinedAt         time.Time `json:"joined_at"`
	IsConnected      bool      `json:"is_connected"`
	IsPending        bool      `json:"is_pending"`
}

// DiscoveryRepository interface for location-based queries
type DiscoveryRepository interface {
	GetNearbyUsers(ctx context.Context, excludeUserID uuid.UUID, lat, lng, radiusKm float64, limit int) ([]NearbyUser, error)
	GetNearbyStories(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]NearbyStory, error)
	GetNearbyEvents(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]NearbyEvent, error)
	UpdateUserLocation(ctx context.Context, userID uuid.UUID, lat, lng float64, city string) error
	GetLocalStats(ctx context.Context, lat, lng, radiusKm float64) (*LocalStats, error)
	GetUserProfile(ctx context.Context, userID uuid.UUID) (*UserProfile, error)
	GetConnectionStatus(ctx context.Context, userID, targetID uuid.UUID) (isConnected, isPending bool, err error)
}

// DiscoveryService handles location-based discovery
type DiscoveryService struct {
	repo DiscoveryRepository
}

// NewDiscoveryService creates a new discovery service
func NewDiscoveryService(repo DiscoveryRepository) *DiscoveryService {
	return &DiscoveryService{repo: repo}
}

// GetNearbyUsers returns users near a location
func (s *DiscoveryService) GetNearbyUsers(ctx context.Context, excludeUserID uuid.UUID, lat, lng, radiusKm float64, limit int) ([]NearbyUser, error) {
	return s.repo.GetNearbyUsers(ctx, excludeUserID, lat, lng, radiusKm, limit)
}

// GetNearbyStories returns stories near a location
func (s *DiscoveryService) GetNearbyStories(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]NearbyStory, error) {
	return s.repo.GetNearbyStories(ctx, lat, lng, radiusKm, limit)
}

// GetNearbyEvents returns events near a location
func (s *DiscoveryService) GetNearbyEvents(ctx context.Context, lat, lng, radiusKm float64, limit int) ([]NearbyEvent, error) {
	return s.repo.GetNearbyEvents(ctx, lat, lng, radiusKm, limit)
}

// UpdateUserLocation updates the user's current location
func (s *DiscoveryService) UpdateUserLocation(ctx context.Context, userID uuid.UUID, lat, lng float64, city string) error {
	return s.repo.UpdateUserLocation(ctx, userID, lat, lng, city)
}

// GetExploreData returns aggregated explore screen data
func (s *DiscoveryService) GetExploreData(ctx context.Context, userID uuid.UUID, lat, lng float64) (*ExploreData, error) {
	// Get nearby users (excluding self)
	users, err := s.repo.GetNearbyUsers(ctx, userID, lat, lng, 10, 10)
	if err != nil {
		return nil, err
	}

	// Get nearby stories
	stories, err := s.repo.GetNearbyStories(ctx, lat, lng, 10, 10)
	if err != nil {
		return nil, err
	}

	// Get nearby events
	events, err := s.repo.GetNearbyEvents(ctx, lat, lng, 25, 5)
	if err != nil {
		return nil, err
	}

	// Get local stats
	stats, err := s.repo.GetLocalStats(ctx, lat, lng, 10)
	if err != nil {
		stats = &LocalStats{}
	}

	return &ExploreData{
		NearbyUsers:    users,
		NearbyStories:  stories,
		NearbyEvents:   events,
		TrendingTopics: []string{"#लोकल", "#भारत", "#समुदाय"},
		SuggestedUsers: users[:min(5, len(users))],
		LocalStats:     *stats,
	}, nil
}

// GetUserProfile returns a user's public profile
func (s *DiscoveryService) GetUserProfile(ctx context.Context, targetID uuid.UUID) (*UserProfile, error) {
	return s.repo.GetUserProfile(ctx, targetID)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

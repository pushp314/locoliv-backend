package domain

import (
	"context"

	"github.com/google/uuid"
)

// StoryVisibility defines who can see a story
type StoryVisibility string

const (
	VisibilityPublic      StoryVisibility = "public"
	VisibilityConnections StoryVisibility = "connections"
	VisibilityFamily      StoryVisibility = "family"
)

// PrivacySettings for a user
type PrivacySettings struct {
	UserID                 uuid.UUID       `json:"user_id"`
	DefaultStoryVisibility StoryVisibility `json:"default_story_visibility"`
	ShowInDiscovery        bool            `json:"show_in_discovery"`
	ShowLocation           bool            `json:"show_location"`
	WhoCanMessage          string          `json:"who_can_message"` // everyone, connections, nobody
	IncognitoMode          bool            `json:"incognito_mode"`
}

// BlockedUser represents a blocked relationship
type BlockedUser struct {
	ID            uuid.UUID `json:"id"`
	UserID        uuid.UUID `json:"user_id"`
	BlockedUserID uuid.UUID `json:"blocked_user_id"`
	Reason        *string   `json:"reason,omitempty"`
}

// CloseFriend represents a close friend/family member
type CloseFriend struct {
	ID       uuid.UUID `json:"id"`
	UserID   uuid.UUID `json:"user_id"`
	FriendID uuid.UUID `json:"friend_id"`
	ListType string    `json:"list_type"` // family, close_friends
}

// PrivacyRepository interface
type PrivacyRepository interface {
	// Settings
	GetPrivacySettings(ctx context.Context, userID uuid.UUID) (*PrivacySettings, error)
	CreatePrivacySettings(ctx context.Context, userID uuid.UUID) (*PrivacySettings, error)
	UpdatePrivacySettings(ctx context.Context, settings *PrivacySettings) error

	// Blocked users
	BlockUser(ctx context.Context, userID, blockedID uuid.UUID, reason *string) error
	UnblockUser(ctx context.Context, userID, blockedID uuid.UUID) error
	GetBlockedUsers(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error)
	IsBlocked(ctx context.Context, userID, otherUserID uuid.UUID) (bool, error)

	// Close friends
	AddCloseFriend(ctx context.Context, userID, friendID uuid.UUID, listType string) error
	RemoveCloseFriend(ctx context.Context, userID, friendID uuid.UUID) error
	GetCloseFriends(ctx context.Context, userID uuid.UUID, listType string) ([]uuid.UUID, error)
	IsCloseFriend(ctx context.Context, userID, friendID uuid.UUID) (bool, error)
}

// PrivacyService handles user privacy settings
type PrivacyService struct {
	repo PrivacyRepository
}

// NewPrivacyService creates a new privacy service
func NewPrivacyService(repo PrivacyRepository) *PrivacyService {
	return &PrivacyService{repo: repo}
}

// GetSettings returns user's privacy settings
func (s *PrivacyService) GetSettings(ctx context.Context, userID uuid.UUID) (*PrivacySettings, error) {
	settings, err := s.repo.GetPrivacySettings(ctx, userID)
	if err != nil {
		// Create default settings
		return s.repo.CreatePrivacySettings(ctx, userID)
	}
	return settings, nil
}

// UpdateSettings updates privacy settings
func (s *PrivacyService) UpdateSettings(ctx context.Context, settings *PrivacySettings) error {
	return s.repo.UpdatePrivacySettings(ctx, settings)
}

// BlockUser blocks another user
func (s *PrivacyService) BlockUser(ctx context.Context, userID, blockedID uuid.UUID, reason *string) error {
	return s.repo.BlockUser(ctx, userID, blockedID, reason)
}

// UnblockUser unblocks a user
func (s *PrivacyService) UnblockUser(ctx context.Context, userID, blockedID uuid.UUID) error {
	return s.repo.UnblockUser(ctx, userID, blockedID)
}

// GetBlockedUsers returns list of blocked user IDs
func (s *PrivacyService) GetBlockedUsers(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return s.repo.GetBlockedUsers(ctx, userID)
}

// CanViewStory checks if a user can view another's story
func (s *PrivacyService) CanViewStory(ctx context.Context, viewerID, authorID uuid.UUID, visibility StoryVisibility) (bool, error) {
	// Check if blocked
	blocked, err := s.repo.IsBlocked(ctx, authorID, viewerID)
	if err != nil {
		return false, err
	}
	if blocked {
		return false, nil
	}

	switch visibility {
	case VisibilityPublic:
		return true, nil
	case VisibilityConnections:
		// TODO: Check if they are connected
		return true, nil // Placeholder
	case VisibilityFamily:
		return s.repo.IsCloseFriend(ctx, authorID, viewerID)
	default:
		return true, nil
	}
}

// AddToFamily adds a user to family/close friends list
func (s *PrivacyService) AddToFamily(ctx context.Context, userID, friendID uuid.UUID) error {
	return s.repo.AddCloseFriend(ctx, userID, friendID, "family")
}

// GetFamilyList returns user's family list
func (s *PrivacyService) GetFamilyList(ctx context.Context, userID uuid.UUID) ([]uuid.UUID, error) {
	return s.repo.GetCloseFriends(ctx, userID, "family")
}

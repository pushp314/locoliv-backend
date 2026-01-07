package domain

import (
	"context"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// DailyMatch represents a suggested connection
type DailyMatch struct {
	ID            uuid.UUID     `json:"id"`
	UserID        uuid.UUID     `json:"user_id"`
	MatchedUserID uuid.UUID     `json:"matched_user_id"`
	MatchDate     string        `json:"match_date"`
	MatchReason   string        `json:"match_reason"`
	IsAccepted    *bool         `json:"is_accepted,omitempty"`
	AcceptedAt    *time.Time    `json:"accepted_at,omitempty"`
	KarmaAwarded  bool          `json:"karma_awarded"`
	MatchedUser   *UserResponse `json:"matched_user,omitempty"`
}

// UserInterest represents a user's interest
type UserInterest struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Interest  string    `json:"interest"`
	Category  string    `json:"category"`
	CreatedAt time.Time `json:"created_at"`
}

// MatchReasons for explaining why users matched
var MatchReasons = []string{
	"Similar interests 🎯",
	"Same city 📍",
	"Both love Cricket 🏏",
	"Music enthusiasts 🎵",
	"Foodies! 🍕",
	"Travel lovers ✈️",
	"Tech enthusiasts 💻",
	"Fitness buddies 💪",
}

// IcebreakerRepository interface
type IcebreakerRepository interface {
	// Daily Matches
	GetTodaysMatch(ctx context.Context, userID uuid.UUID, date string) (*DailyMatch, error)
	CreateMatch(ctx context.Context, userID, matchedUserID uuid.UUID, reason string) (*DailyMatch, error)
	RespondToMatch(ctx context.Context, matchID uuid.UUID, accept bool) error
	MarkKarmaAwarded(ctx context.Context, matchID uuid.UUID) error

	// Find potential matches
	FindPotentialMatches(ctx context.Context, userID uuid.UUID, city *string, limit int) ([]uuid.UUID, error)

	// Interests
	GetUserInterests(ctx context.Context, userID uuid.UUID) ([]UserInterest, error)
	AddInterest(ctx context.Context, userID uuid.UUID, interest, category string) error
	RemoveInterest(ctx context.Context, userID uuid.UUID, interest string) error
	GetPopularInterests(ctx context.Context, limit int) ([]struct {
		Interest string
		Category string
		Count    int
	}, error)
}

// IcebreakerService handles daily matching
type IcebreakerService struct {
	repo         IcebreakerRepository
	karmaService *KarmaService
}

// NewIcebreakerService creates a new icebreaker service
func NewIcebreakerService(repo IcebreakerRepository, karmaService *KarmaService) *IcebreakerService {
	return &IcebreakerService{
		repo:         repo,
		karmaService: karmaService,
	}
}

// GetTodaysMatch returns or creates today's match for a user
func (s *IcebreakerService) GetTodaysMatch(ctx context.Context, userID uuid.UUID, city *string) (*DailyMatch, error) {
	today := time.Now().Format("2006-01-02")

	// Check if match already exists
	match, err := s.repo.GetTodaysMatch(ctx, userID, today)
	if err == nil && match != nil {
		return match, nil
	}

	// Find a new match
	potentialMatches, err := s.repo.FindPotentialMatches(ctx, userID, city, 10)
	if err != nil || len(potentialMatches) == 0 {
		return nil, nil // No match available
	}

	// Pick random match
	matchedUserID := potentialMatches[rand.Intn(len(potentialMatches))]
	reason := MatchReasons[rand.Intn(len(MatchReasons))]

	return s.repo.CreateMatch(ctx, userID, matchedUserID, reason)
}

// RespondToMatch handles accept/decline
func (s *IcebreakerService) RespondToMatch(ctx context.Context, userID, matchID uuid.UUID, accept bool) (int, error) {
	if err := s.repo.RespondToMatch(ctx, matchID, accept); err != nil {
		return 0, err
	}

	karmaEarned := 0
	if accept {
		// Award karma for connecting
		refType := "icebreaker"
		_ = s.karmaService.EarnKarma(ctx, userID, "icebreaker_connect", &refType, &matchID)
		_ = s.repo.MarkKarmaAwarded(ctx, matchID)
		karmaEarned = 25
	}

	return karmaEarned, nil
}

// GetUserInterests returns user's interests
func (s *IcebreakerService) GetUserInterests(ctx context.Context, userID uuid.UUID) ([]UserInterest, error) {
	return s.repo.GetUserInterests(ctx, userID)
}

// AddInterest adds an interest
func (s *IcebreakerService) AddInterest(ctx context.Context, userID uuid.UUID, interest, category string) error {
	return s.repo.AddInterest(ctx, userID, interest, category)
}

// RemoveInterest removes an interest
func (s *IcebreakerService) RemoveInterest(ctx context.Context, userID uuid.UUID, interest string) error {
	return s.repo.RemoveInterest(ctx, userID, interest)
}

// GetSuggestedInterests returns popular interests
func (s *IcebreakerService) GetSuggestedInterests(ctx context.Context) ([]struct {
	Interest string
	Category string
	Count    int
}, error) {
	return s.repo.GetPopularInterests(ctx, 20)
}

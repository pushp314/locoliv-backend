package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// KarmaRepository interface for karma data access
type KarmaRepository interface {
	// Karma
	GetUserKarma(ctx context.Context, userID uuid.UUID) (*UserKarma, error)
	CreateUserKarma(ctx context.Context, userID uuid.UUID) (*UserKarma, error)
	UpdateUserKarma(ctx context.Context, userID uuid.UUID, totalKarma int, tier KarmaTier, prabhav float64) error
	UpdateLoginStreak(ctx context.Context, userID uuid.UUID, streak int, loginDate string) error

	// Transactions
	CreateKarmaTransaction(ctx context.Context, userID uuid.UUID, action KarmaAction, points int, refType *string, refID *uuid.UUID, desc *string) error
	GetKarmaTransactions(ctx context.Context, userID uuid.UUID, limit, offset int) ([]KarmaTransaction, error)

	// Badges
	GetAllBadges(ctx context.Context) ([]Badge, error)
	GetUserBadges(ctx context.Context, userID uuid.UUID) ([]UserBadge, error)
	AwardBadge(ctx context.Context, userID uuid.UUID, badgeCode string) error

	// Subscriptions
	GetActiveSubscription(ctx context.Context, userID uuid.UUID) (*Subscription, error)
	CreateSubscription(ctx context.Context, sub *Subscription) error
	UpdateSubscriptionStatus(ctx context.Context, subID uuid.UUID, status SubscriptionStatus) error

	// Boosts
	CreateStoryBoost(ctx context.Context, boost *StoryBoost) error
	GetActiveBoosts(ctx context.Context, storyID uuid.UUID) ([]StoryBoost, error)
	GetUserBoostCount(ctx context.Context, userID uuid.UUID, since time.Time) (int, error)

	// Stats for Prabhav calculation
	GetUserStoryStats(ctx context.Context, userID uuid.UUID) (totalStories, totalViews, totalReactions, totalReplies int, err error)
}

// KarmaService handles karma business logic
type KarmaService struct {
	repo KarmaRepository
}

// NewKarmaService creates a new karma service
func NewKarmaService(repo KarmaRepository) *KarmaService {
	return &KarmaService{repo: repo}
}

// EarnKarma adds karma for an action
func (s *KarmaService) EarnKarma(ctx context.Context, userID uuid.UUID, action KarmaAction, refType *string, refID *uuid.UUID) error {
	points := KarmaPoints[action]
	if points == 0 {
		return nil // Unknown action
	}

	// Get or create karma record
	karma, err := s.repo.GetUserKarma(ctx, userID)
	if err != nil {
		karma, err = s.repo.CreateUserKarma(ctx, userID)
		if err != nil {
			return err
		}
	}

	// Record transaction
	desc := string(action) + " earned"
	if err := s.repo.CreateKarmaTransaction(ctx, userID, action, points, refType, refID, &desc); err != nil {
		return err
	}

	// Update total and tier
	newTotal := karma.TotalKarma + points
	newTier := CalculateTier(newTotal)

	// Check if tier changed and award badge
	if newTier != karma.CurrentTier {
		_ = s.repo.AwardBadge(ctx, userID, string(newTier))
	}

	return s.repo.UpdateUserKarma(ctx, userID, newTotal, newTier, karma.PrabhavScore)
}

// RecordDailyLogin handles login streak
func (s *KarmaService) RecordDailyLogin(ctx context.Context, userID uuid.UUID) error {
	karma, err := s.repo.GetUserKarma(ctx, userID)
	if err != nil {
		karma, err = s.repo.CreateUserKarma(ctx, userID)
		if err != nil {
			return err
		}
	}

	today := time.Now().Format("2006-01-02")

	// Check if already logged in today
	if karma.LastLoginDate != nil && *karma.LastLoginDate == today {
		return nil // Already recorded
	}

	// Calculate streak
	newStreak := 1
	if karma.LastLoginDate != nil {
		yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")
		if *karma.LastLoginDate == yesterday {
			newStreak = karma.LoginStreak + 1
		}
	}

	// Cap streak bonus at 10 days
	if newStreak <= 10 {
		points := KarmaPoints[KarmaActionDailyLogin]
		desc := "Daily login streak"
		refType := "login"
		_ = s.repo.CreateKarmaTransaction(ctx, userID, KarmaActionDailyLogin, points, &refType, nil, &desc)

		// Update total
		newTotal := karma.TotalKarma + points
		newTier := CalculateTier(newTotal)
		_ = s.repo.UpdateUserKarma(ctx, userID, newTotal, newTier, karma.PrabhavScore)
	}

	return s.repo.UpdateLoginStreak(ctx, userID, newStreak, today)
}

// CalculatePrabhav computes impact score
func (s *KarmaService) CalculatePrabhav(ctx context.Context, userID uuid.UUID) (float64, error) {
	totalStories, totalViews, totalReactions, totalReplies, err := s.repo.GetUserStoryStats(ctx, userID)
	if err != nil {
		return 0, err
	}

	if totalStories == 0 {
		return 0, nil
	}

	// Prabhav formula: (Views*0.5 + Reactions*2 + Replies*5) / TotalStories
	prabhav := (float64(totalViews)*0.5 + float64(totalReactions)*2 + float64(totalReplies)*5) / float64(totalStories)

	// Update stored value
	karma, err := s.repo.GetUserKarma(ctx, userID)
	if err == nil {
		_ = s.repo.UpdateUserKarma(ctx, userID, karma.TotalKarma, karma.CurrentTier, prabhav)
	}

	return prabhav, nil
}

// GetKarmaStatus returns full karma status for a user
func (s *KarmaService) GetKarmaStatus(ctx context.Context, userID uuid.UUID) (*KarmaResponse, error) {
	karma, err := s.repo.GetUserKarma(ctx, userID)
	if err != nil {
		// Create if doesn't exist
		karma, err = s.repo.CreateUserKarma(ctx, userID)
		if err != nil {
			return nil, err
		}
	}

	badges, _ := s.repo.GetUserBadges(ctx, userID)
	recent, _ := s.repo.GetKarmaTransactions(ctx, userID, 10, 0)

	resp := &KarmaResponse{
		TotalKarma:     karma.TotalKarma,
		CurrentTier:    string(karma.CurrentTier),
		TierName:       TierDisplayName(karma.CurrentTier),
		PrabhavScore:   karma.PrabhavScore,
		LoginStreak:    karma.LoginStreak,
		Badges:         badges,
		RecentActivity: recent,
	}

	// Calculate next tier info
	nextTier, pointsNeeded := s.getNextTierInfo(karma.TotalKarma)
	if nextTier != nil {
		tier := string(*nextTier)
		resp.NextTier = &tier
		resp.PointsToNext = &pointsNeeded
	}

	return resp, nil
}

func (s *KarmaService) getNextTierInfo(currentKarma int) (*KarmaTier, int) {
	tiers := []KarmaTier{TierNavasadhak, TierPrayatnasheel, TierSaathi, TierMargdarshak, TierPrabhavshali, TierLokpriya, TierYogi}

	for _, tier := range tiers {
		threshold := TierThresholds[tier]
		if currentKarma < threshold {
			needed := threshold - currentKarma
			return &tier, needed
		}
	}
	return nil, 0 // Already at max
}

// RedeemKarmaForPremium converts karma to premium days
func (s *KarmaService) RedeemKarmaForPremium(ctx context.Context, userID uuid.UUID) (*Subscription, error) {
	karma, err := s.repo.GetUserKarma(ctx, userID)
	if err != nil {
		return nil, err
	}

	// Check eligibility
	var days int
	var cost int
	switch {
	case karma.TotalKarma >= 10000:
		days = 30
		cost = 10000
	case karma.TotalKarma >= 5000:
		days = 7
		cost = 5000
	default:
		return nil, ErrInsufficientKarma
	}

	// Deduct karma
	newTotal := karma.TotalKarma - cost
	newTier := CalculateTier(newTotal)
	desc := "Redeemed for premium"
	refType := "premium"
	_ = s.repo.CreateKarmaTransaction(ctx, userID, "karma_redeem", -cost, &refType, nil, &desc)
	_ = s.repo.UpdateUserKarma(ctx, userID, newTotal, newTier, karma.PrabhavScore)

	// Create subscription
	provider := "karma"
	sub := &Subscription{
		ID:              uuid.New(),
		UserID:          userID,
		Plan:            PlanKarmaFree,
		Status:          SubscriptionActive,
		StartedAt:       time.Now(),
		ExpiresAt:       time.Now().AddDate(0, 0, days),
		PaymentProvider: &provider,
	}

	if err := s.repo.CreateSubscription(ctx, sub); err != nil {
		return nil, err
	}

	return sub, nil
}

// GetPremiumStatus checks if user has active premium
func (s *KarmaService) GetPremiumStatus(ctx context.Context, userID uuid.UUID) (*PremiumStatus, error) {
	status := &PremiumStatus{
		IsPremium:        false,
		CanRedeemKarma:   false,
		KarmaForFreeDays: 0,
	}

	// Check subscription
	sub, err := s.repo.GetActiveSubscription(ctx, userID)
	if err == nil && sub != nil && sub.ExpiresAt.After(time.Now()) {
		status.IsPremium = true
		plan := string(sub.Plan)
		status.Plan = &plan
		status.ExpiresAt = &sub.ExpiresAt
	}

	// Check karma eligibility
	karma, err := s.repo.GetUserKarma(ctx, userID)
	if err == nil {
		if karma.TotalKarma >= 10000 {
			status.CanRedeemKarma = true
			status.KarmaForFreeDays = 30
		} else if karma.TotalKarma >= 5000 {
			status.CanRedeemKarma = true
			status.KarmaForFreeDays = 7
		}
	}

	return status, nil
}

// Custom errors
var (
	ErrInsufficientKarma = &AppError{Code: "INSUFFICIENT_KARMA", Message: "Not enough karma points"}
)

type AppError struct {
	Code    string
	Message string
}

func (e *AppError) Error() string {
	return e.Message
}

package domain

import (
	"time"

	"github.com/google/uuid"
)

// KarmaAction represents types of actions that earn/lose karma
type KarmaAction string

const (
	KarmaActionStoryPost       KarmaAction = "story_post"
	KarmaActionStoryView       KarmaAction = "story_view"
	KarmaActionStoryReaction   KarmaAction = "story_reaction"
	KarmaActionStoryReply      KarmaAction = "story_reply"
	KarmaActionConnection      KarmaAction = "connection"
	KarmaActionDailyLogin      KarmaAction = "daily_login"
	KarmaActionProfileComplete KarmaAction = "profile_complete"
	KarmaActionFirstStoryWeek  KarmaAction = "first_story_week"
	KarmaActionThanks          KarmaAction = "thanks"
	KarmaActionReported        KarmaAction = "reported"
	KarmaActionBlocked         KarmaAction = "blocked"
	KarmaActionSpam            KarmaAction = "spam"
)

// KarmaPoints defines points for each action
var KarmaPoints = map[KarmaAction]int{
	KarmaActionStoryPost:       5,
	KarmaActionStoryView:       1,
	KarmaActionStoryReaction:   2,
	KarmaActionStoryReply:      3,
	KarmaActionConnection:      10,
	KarmaActionDailyLogin:      5,
	KarmaActionProfileComplete: 25,
	KarmaActionFirstStoryWeek:  15,
	KarmaActionThanks:          20,
	KarmaActionReported:        -50,
	KarmaActionBlocked:         -25,
	KarmaActionSpam:            -100,
}

// KarmaTier represents user tiers based on karma
type KarmaTier string

const (
	TierNavasadhak    KarmaTier = "navasadhak"
	TierPrayatnasheel KarmaTier = "prayatnasheel"
	TierSaathi        KarmaTier = "saathi"
	TierMargdarshak   KarmaTier = "margdarshak"
	TierPrabhavshali  KarmaTier = "prabhavshali"
	TierLokpriya      KarmaTier = "lokpriya"
	TierYogi          KarmaTier = "yogi"
)

// TierThresholds maps tiers to minimum karma required
var TierThresholds = map[KarmaTier]int{
	TierNavasadhak:    0,
	TierPrayatnasheel: 100,
	TierSaathi:        500,
	TierMargdarshak:   1000,
	TierPrabhavshali:  2500,
	TierLokpriya:      5000,
	TierYogi:          10000,
}

// UserKarma represents a user's karma state
type UserKarma struct {
	UserID           uuid.UUID `json:"user_id"`
	TotalKarma       int       `json:"total_karma"`
	CurrentTier      KarmaTier `json:"current_tier"`
	PrabhavScore     float64   `json:"prabhav_score"`
	LoginStreak      int       `json:"login_streak"`
	LastLoginDate    *string   `json:"last_login_date,omitempty"`
	LastCalculatedAt time.Time `json:"last_calculated_at"`
}

// KarmaTransaction represents a single karma change
type KarmaTransaction struct {
	ID            uuid.UUID   `json:"id"`
	UserID        uuid.UUID   `json:"user_id"`
	Action        KarmaAction `json:"action"`
	Points        int         `json:"points"`
	ReferenceType *string     `json:"reference_type,omitempty"`
	ReferenceID   *uuid.UUID  `json:"reference_id,omitempty"`
	Description   *string     `json:"description,omitempty"`
	CreatedAt     time.Time   `json:"created_at"`
}

// Badge represents an achievement badge
type Badge struct {
	Code         string  `json:"code"`
	NameSanskrit string  `json:"name_sanskrit"`
	NameHindi    string  `json:"name_hindi"`
	Description  *string `json:"description,omitempty"`
	MinKarma     int     `json:"min_karma"`
	IconURL      *string `json:"icon_url,omitempty"`
	SortOrder    int     `json:"sort_order"`
}

// UserBadge represents a badge earned by a user
type UserBadge struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	BadgeCode string    `json:"badge_code"`
	Badge     *Badge    `json:"badge,omitempty"`
	EarnedAt  time.Time `json:"earned_at"`
}

// SubscriptionPlan types
type SubscriptionPlan string

const (
	PlanPlusMonthly SubscriptionPlan = "plus_monthly"
	PlanPlusAnnual  SubscriptionPlan = "plus_annual"
	PlanKarmaFree   SubscriptionPlan = "karma_free"
)

// SubscriptionStatus types
type SubscriptionStatus string

const (
	SubscriptionActive    SubscriptionStatus = "active"
	SubscriptionCancelled SubscriptionStatus = "cancelled"
	SubscriptionExpired   SubscriptionStatus = "expired"
)

// Subscription represents a premium subscription
type Subscription struct {
	ID              uuid.UUID          `json:"id"`
	UserID          uuid.UUID          `json:"user_id"`
	Plan            SubscriptionPlan   `json:"plan"`
	Status          SubscriptionStatus `json:"status"`
	StartedAt       time.Time          `json:"started_at"`
	ExpiresAt       time.Time          `json:"expires_at"`
	PaymentProvider *string            `json:"payment_provider,omitempty"`
	PaymentID       *string            `json:"payment_id,omitempty"`
	AmountPaid      *int               `json:"amount_paid,omitempty"`
}

// StoryBoost represents a boosted story
type StoryBoost struct {
	ID              uuid.UUID `json:"id"`
	StoryID         uuid.UUID `json:"story_id"`
	UserID          uuid.UUID `json:"user_id"`
	BoostType       string    `json:"boost_type"`
	BoostMultiplier float64   `json:"boost_multiplier"`
	BoostedAt       time.Time `json:"boosted_at"`
	ExpiresAt       time.Time `json:"expires_at"`
}

// KarmaResponse for API responses
type KarmaResponse struct {
	TotalKarma     int                `json:"total_karma"`
	CurrentTier    string             `json:"current_tier"`
	TierName       string             `json:"tier_name"`
	PrabhavScore   float64            `json:"prabhav_score"`
	LoginStreak    int                `json:"login_streak"`
	NextTier       *string            `json:"next_tier,omitempty"`
	PointsToNext   *int               `json:"points_to_next,omitempty"`
	Badges         []UserBadge        `json:"badges"`
	RecentActivity []KarmaTransaction `json:"recent_activity,omitempty"`
}

// PremiumStatus for checking subscription
type PremiumStatus struct {
	IsPremium        bool       `json:"is_premium"`
	Plan             *string    `json:"plan,omitempty"`
	ExpiresAt        *time.Time `json:"expires_at,omitempty"`
	CanRedeemKarma   bool       `json:"can_redeem_karma"`
	KarmaForFreeDays int        `json:"karma_for_free_days"`
}

// CalculateTier determines tier from karma
func CalculateTier(karma int) KarmaTier {
	switch {
	case karma >= 10000:
		return TierYogi
	case karma >= 5000:
		return TierLokpriya
	case karma >= 2500:
		return TierPrabhavshali
	case karma >= 1000:
		return TierMargdarshak
	case karma >= 500:
		return TierSaathi
	case karma >= 100:
		return TierPrayatnasheel
	default:
		return TierNavasadhak
	}
}

// TierDisplayName returns Hindi display name
func TierDisplayName(tier KarmaTier) string {
	names := map[KarmaTier]string{
		TierNavasadhak:    "नवसाधक",
		TierPrayatnasheel: "प्रयत्नशील",
		TierSaathi:        "साथी",
		TierMargdarshak:   "मार्गदर्शक",
		TierPrabhavshali:  "प्रभावशाली",
		TierLokpriya:      "लोकप्रिय",
		TierYogi:          "योगी",
	}
	return names[tier]
}

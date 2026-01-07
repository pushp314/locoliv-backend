package domain

import (
	"context"
	"math/rand"
	"time"

	"github.com/google/uuid"
)

// Challenge represents a daily challenge for a user
type Challenge struct {
	ID          uuid.UUID  `json:"id"`
	UserID      uuid.UUID  `json:"user_id"`
	Type        string     `json:"challenge_type"`
	Title       string     `json:"title"`
	TitleHindi  *string    `json:"title_hindi,omitempty"`
	Description *string    `json:"description,omitempty"`
	RewardKarma int        `json:"reward_karma"`
	IsCompleted bool       `json:"is_completed"`
	Progress    int        `json:"progress"`
	Target      int        `json:"target"`
	AssignedAt  string     `json:"assigned_date"`
	CompletedAt *time.Time `json:"completed_at,omitempty"`
	ExpiresAt   time.Time  `json:"expires_at"`
}

// ChallengeTemplate for random challenge assignment
type ChallengeTemplate struct {
	ID               uuid.UUID `json:"id"`
	Type             string    `json:"type"`
	Title            string    `json:"title"`
	TitleHindi       *string   `json:"title_hindi,omitempty"`
	Description      *string   `json:"description,omitempty"`
	DescriptionHindi *string   `json:"description_hindi,omitempty"`
	RewardKarma      int       `json:"reward_karma"`
	Target           int       `json:"target"`
	Difficulty       string    `json:"difficulty"`
}

// FestivalEvent for bonus karma during festivals
type FestivalEvent struct {
	ID              uuid.UUID `json:"id"`
	Name            string    `json:"name"`
	NameHindi       *string   `json:"name_hindi,omitempty"`
	Description     *string   `json:"description,omitempty"`
	KarmaMultiplier float64   `json:"karma_multiplier"`
	BonusKarma      int       `json:"bonus_karma"`
	StartDate       time.Time `json:"start_date"`
	EndDate         time.Time `json:"end_date"`
	IsActive        bool      `json:"is_active"`
}

// UserStreak tracks user challenge streaks
type UserStreak struct {
	UserID                   uuid.UUID `json:"user_id"`
	CurrentStreak            int       `json:"current_streak"`
	LongestStreak            int       `json:"longest_streak"`
	LastChallengeDate        *string   `json:"last_challenge_date,omitempty"`
	TotalChallengesCompleted int       `json:"total_challenges_completed"`
}

// ChallengeRepository interface
type ChallengeRepository interface {
	// Challenges
	GetUserChallenges(ctx context.Context, userID uuid.UUID, date string) ([]Challenge, error)
	CreateChallenge(ctx context.Context, c *Challenge) error
	UpdateChallengeProgress(ctx context.Context, challengeID uuid.UUID, progress int) error
	CompleteChallenge(ctx context.Context, challengeID uuid.UUID) error

	// Templates
	GetActiveTemplates(ctx context.Context) ([]ChallengeTemplate, error)
	GetRandomTemplates(ctx context.Context, count int, excludeTypes []string) ([]ChallengeTemplate, error)

	// Festivals
	GetActiveFestival(ctx context.Context) (*FestivalEvent, error)
	GetUpcomingFestivals(ctx context.Context, limit int) ([]FestivalEvent, error)

	// Streaks
	GetUserStreak(ctx context.Context, userID uuid.UUID) (*UserStreak, error)
	UpdateUserStreak(ctx context.Context, userID uuid.UUID, streak int, lastDate string, total int) error
}

// ChallengeService handles daily challenges
type ChallengeService struct {
	challengeRepo ChallengeRepository
	karmaService  *KarmaService
}

// NewChallengeService creates a new challenge service
func NewChallengeService(repo ChallengeRepository, karmaService *KarmaService) *ChallengeService {
	return &ChallengeService{
		challengeRepo: repo,
		karmaService:  karmaService,
	}
}

// GetTodaysChallenges returns user's challenges for today, creating if needed
func (s *ChallengeService) GetTodaysChallenges(ctx context.Context, userID uuid.UUID) ([]Challenge, *FestivalEvent, error) {
	today := time.Now().Format("2006-01-02")

	// Check existing challenges
	challenges, err := s.challengeRepo.GetUserChallenges(ctx, userID, today)
	if err != nil {
		return nil, nil, err
	}

	// If no challenges for today, assign new ones
	if len(challenges) == 0 {
		challenges, err = s.assignDailyChallenges(ctx, userID)
		if err != nil {
			return nil, nil, err
		}
	}

	// Check for active festival
	festival, _ := s.challengeRepo.GetActiveFestival(ctx)

	return challenges, festival, nil
}

// assignDailyChallenges creates 3 random challenges for the user
func (s *ChallengeService) assignDailyChallenges(ctx context.Context, userID uuid.UUID) ([]Challenge, error) {
	templates, err := s.challengeRepo.GetRandomTemplates(ctx, 3, nil)
	if err != nil || len(templates) == 0 {
		// Fallback to default challenges
		templates = s.getDefaultTemplates()
	}

	today := time.Now().Format("2006-01-02")
	expiresAt := time.Now().Add(24 * time.Hour)

	var challenges []Challenge
	for _, t := range templates {
		c := &Challenge{
			ID:          uuid.New(),
			UserID:      userID,
			Type:        t.Type,
			Title:       t.Title,
			TitleHindi:  t.TitleHindi,
			Description: t.Description,
			RewardKarma: t.RewardKarma,
			IsCompleted: false,
			Progress:    0,
			Target:      t.Target,
			AssignedAt:  today,
			ExpiresAt:   expiresAt,
		}

		if err := s.challengeRepo.CreateChallenge(ctx, c); err != nil {
			continue
		}
		challenges = append(challenges, *c)
	}

	return challenges, nil
}

// UpdateProgress updates challenge progress based on user action
func (s *ChallengeService) UpdateProgress(ctx context.Context, userID uuid.UUID, challengeType string, increment int) error {
	today := time.Now().Format("2006-01-02")
	challenges, err := s.challengeRepo.GetUserChallenges(ctx, userID, today)
	if err != nil {
		return err
	}

	for _, c := range challenges {
		if c.Type == challengeType && !c.IsCompleted {
			newProgress := c.Progress + increment
			if newProgress >= c.Target {
				// Complete the challenge
				if err := s.challengeRepo.CompleteChallenge(ctx, c.ID); err != nil {
					return err
				}

				// Award karma with festival multiplier
				reward := c.RewardKarma
				festival, _ := s.challengeRepo.GetActiveFestival(ctx)
				if festival != nil {
					reward = int(float64(reward) * festival.KarmaMultiplier)
				}

				refType := "challenge"
				_ = s.karmaService.EarnKarma(ctx, userID, "challenge_complete", &refType, &c.ID)

				// Update streak
				s.updateStreak(ctx, userID)
			} else {
				_ = s.challengeRepo.UpdateChallengeProgress(ctx, c.ID, newProgress)
			}
			break
		}
	}

	return nil
}

// updateStreak updates user's challenge completion streak
func (s *ChallengeService) updateStreak(ctx context.Context, userID uuid.UUID) {
	streak, err := s.challengeRepo.GetUserStreak(ctx, userID)
	if err != nil {
		streak = &UserStreak{UserID: userID}
	}

	today := time.Now().Format("2006-01-02")
	yesterday := time.Now().AddDate(0, 0, -1).Format("2006-01-02")

	newStreak := 1
	if streak.LastChallengeDate != nil && *streak.LastChallengeDate == yesterday {
		newStreak = streak.CurrentStreak + 1
	} else if streak.LastChallengeDate != nil && *streak.LastChallengeDate == today {
		return // Already updated today
	}

	longest := streak.LongestStreak
	if newStreak > longest {
		longest = newStreak
	}

	_ = s.challengeRepo.UpdateUserStreak(ctx, userID, newStreak, today, streak.TotalChallengesCompleted+1)

	// Streak bonuses
	switch newStreak {
	case 7:
		refType := "streak"
		_ = s.karmaService.EarnKarma(ctx, userID, "streak_7_day", &refType, nil)
	case 30:
		refType := "streak"
		_ = s.karmaService.EarnKarma(ctx, userID, "streak_30_day", &refType, nil)
	}
}

// GetFestivalStatus returns current festival if any
func (s *ChallengeService) GetFestivalStatus(ctx context.Context) (*FestivalEvent, error) {
	return s.challengeRepo.GetActiveFestival(ctx)
}

// getDefaultTemplates returns fallback templates
func (s *ChallengeService) getDefaultTemplates() []ChallengeTemplate {
	hindi1 := "अपना दिन शेयर करो"
	hindi2 := "प्यार बांटो"
	hindi3 := "किसी की स्टोरी पर रिप्लाई करो"

	return []ChallengeTemplate{
		{Type: "post_story", Title: "Share Your Day", TitleHindi: &hindi1, RewardKarma: 15, Target: 1},
		{Type: "react_stories", Title: "Spread Love", TitleHindi: &hindi2, RewardKarma: 20, Target: 5},
		{Type: "reply_story", Title: "Start a Conversation", TitleHindi: &hindi3, RewardKarma: 25, Target: 1},
	}
}

func init() {
	rand.Seed(time.Now().UnixNano())
}

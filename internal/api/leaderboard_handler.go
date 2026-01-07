package api

import (
	"net/http"

	"github.com/locolive/backend/internal/middleware"
	"github.com/locolive/backend/pkg/response"
	"go.uber.org/zap"
)

type LeaderboardHandler struct {
	logger *zap.Logger
	// Would need karma repository for actual implementation
}

func NewLeaderboardHandler(logger *zap.Logger) *LeaderboardHandler {
	return &LeaderboardHandler{logger: logger}
}

// LeaderboardEntry represents a user on the leaderboard
type LeaderboardEntry struct {
	Rank       int    `json:"rank"`
	UserID     string `json:"user_id"`
	Name       string `json:"name"`
	AvatarURL  string `json:"avatar_url,omitempty"`
	TotalKarma int    `json:"total_karma"`
	Tier       string `json:"tier"`
	TierName   string `json:"tier_name"` // Hindi name
}

// GetCityLeaderboard returns top karma earners in a city
func (h *LeaderboardHandler) GetCityLeaderboard(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	city := r.URL.Query().Get("city")
	if city == "" {
		city = "Your City" // Default
	}

	// TODO: Implement actual DB query
	// SELECT u.id, u.name, uk.total_karma, uk.current_tier
	// FROM user_karma uk
	// JOIN users u ON uk.user_id = u.id
	// JOIN profiles p ON u.id = p.user_id
	// WHERE p.city = $1
	// ORDER BY uk.total_karma DESC
	// LIMIT 10

	// Mock leaderboard for now
	mockLeaderboard := []LeaderboardEntry{
		{Rank: 1, UserID: "1", Name: "राहुल शर्मा", TotalKarma: 15420, Tier: "yogi", TierName: "योगी"},
		{Rank: 2, UserID: "2", Name: "Priya Patel", TotalKarma: 12350, Tier: "yogi", TierName: "योगी"},
		{Rank: 3, UserID: "3", Name: "अमित वर्मा", TotalKarma: 8900, Tier: "lokpriya", TierName: "लोकप्रिय"},
		{Rank: 4, UserID: "4", Name: "Sneha Gupta", TotalKarma: 6780, Tier: "lokpriya", TierName: "लोकप्रिय"},
		{Rank: 5, UserID: "5", Name: "विकास सिंह", TotalKarma: 5200, Tier: "lokpriya", TierName: "लोकप्रिय"},
		{Rank: 6, UserID: "6", Name: "Neha Sharma", TotalKarma: 4100, Tier: "prabhavshali", TierName: "प्रभावशाली"},
		{Rank: 7, UserID: "7", Name: "राज कुमार", TotalKarma: 3500, Tier: "prabhavshali", TierName: "प्रभावशाली"},
		{Rank: 8, UserID: "8", Name: "Ananya Roy", TotalKarma: 2800, Tier: "prabhavshali", TierName: "प्रभावशाली"},
		{Rank: 9, UserID: "9", Name: "करण मेहता", TotalKarma: 1900, Tier: "margdarshak", TierName: "मार्गदर्शक"},
		{Rank: 10, UserID: "10", Name: "Pooja Jain", TotalKarma: 1200, Tier: "margdarshak", TierName: "मार्गदर्शक"},
	}

	response.OK(w, map[string]interface{}{
		"city":        city,
		"leaderboard": mockLeaderboard,
		"updated_at":  "2026-01-05T23:00:00Z",
	})
}

// GetMyRank returns the current user's rank
func (h *LeaderboardHandler) GetMyRank(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	city := r.URL.Query().Get("city")
	if city == "" {
		city = "Your City"
	}

	// TODO: Implement actual rank calculation
	response.OK(w, map[string]interface{}{
		"city":        city,
		"rank":        42,
		"total_users": 156,
		"user_id":     userID.String(),
		"message":     "You're in the top 27% of karma earners in your city!",
	})
}

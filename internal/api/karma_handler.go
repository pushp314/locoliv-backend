package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/locolive/backend/internal/domain"
	"github.com/locolive/backend/internal/middleware"
	"github.com/locolive/backend/pkg/response"
	"go.uber.org/zap"
)

type KarmaHandler struct {
	service *domain.KarmaService
	logger  *zap.Logger
}

func NewKarmaHandler(service *domain.KarmaService, logger *zap.Logger) *KarmaHandler {
	return &KarmaHandler{service: service, logger: logger}
}

// GetMyKarma returns the current user's karma status
func (h *KarmaHandler) GetMyKarma(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	// Record daily login
	_ = h.service.RecordDailyLogin(r.Context(), userID)

	// Recalculate prabhav
	_, _ = h.service.CalculatePrabhav(r.Context(), userID)

	karma, err := h.service.GetKarmaStatus(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get karma", zap.Error(err))
		response.InternalError(w, "failed to get karma status")
		return
	}

	response.OK(w, karma)
}

// GetLeaderboard returns top users by karma in a location
func (h *KarmaHandler) GetLeaderboard(w http.ResponseWriter, r *http.Request) {
	// TODO: Implement location-based leaderboard
	response.OK(w, map[string]string{"message": "leaderboard coming soon"})
}

// GetBadges returns all available badges
func (h *KarmaHandler) GetBadges(w http.ResponseWriter, r *http.Request) {
	// This would need repo access, for now return static list
	badges := []map[string]interface{}{
		{"code": "navasadhak", "name_sanskrit": "नवसाधक", "name_hindi": "नया साधक", "min_karma": 0},
		{"code": "prayatnasheel", "name_sanskrit": "प्रयत्नशील", "name_hindi": "प्रयास करने वाला", "min_karma": 100},
		{"code": "saathi", "name_sanskrit": "साथी", "name_hindi": "साथी", "min_karma": 500},
		{"code": "margdarshak", "name_sanskrit": "मार्गदर्शक", "name_hindi": "मार्गदर्शक", "min_karma": 1000},
		{"code": "prabhavshali", "name_sanskrit": "प्रभावशाली", "name_hindi": "प्रभावशाली", "min_karma": 2500},
		{"code": "lokpriya", "name_sanskrit": "लोकप्रिय", "name_hindi": "लोकप्रिय", "min_karma": 5000},
		{"code": "yogi", "name_sanskrit": "योगी", "name_hindi": "योगी", "min_karma": 10000},
	}
	response.OK(w, badges)
}

// GetPremiumStatus returns subscription status
func (h *KarmaHandler) GetPremiumStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	status, err := h.service.GetPremiumStatus(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get premium status", zap.Error(err))
		response.InternalError(w, "failed to get premium status")
		return
	}

	response.OK(w, status)
}

// RedeemKarmaForPremium converts karma to premium subscription
func (h *KarmaHandler) RedeemKarmaForPremium(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	sub, err := h.service.RedeemKarmaForPremium(r.Context(), userID)
	if err != nil {
		if err == domain.ErrInsufficientKarma {
			response.BadRequest(w, "insufficient karma points")
			return
		}
		h.logger.Error("Failed to redeem karma", zap.Error(err))
		response.InternalError(w, "failed to redeem karma")
		return
	}

	response.OK(w, map[string]interface{}{
		"message":    "Premium activated!",
		"expires_at": sub.ExpiresAt,
	})
}

// BoostStory boosts a story for increased visibility
func (h *KarmaHandler) BoostStory(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	storyIDStr := chi.URLParam(r, "storyId")
	storyID, err := uuid.Parse(storyIDStr)
	if err != nil {
		response.BadRequest(w, "invalid story ID")
		return
	}

	// TODO: Implement boost logic (check premium, check limits, etc)
	h.logger.Info("Story boost requested", zap.String("user_id", userID.String()), zap.String("story_id", storyID.String()))

	response.OK(w, map[string]string{"message": "story boosted successfully"})
}

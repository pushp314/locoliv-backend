package api

import (
	"net/http"

	"github.com/locolive/backend/internal/domain"
	"github.com/locolive/backend/internal/middleware"
	"github.com/locolive/backend/pkg/response"
	"go.uber.org/zap"
)

type ChallengeHandler struct {
	service *domain.ChallengeService
	logger  *zap.Logger
}

func NewChallengeHandler(service *domain.ChallengeService, logger *zap.Logger) *ChallengeHandler {
	return &ChallengeHandler{service: service, logger: logger}
}

// GetDailyChallenges returns today's challenges
func (h *ChallengeHandler) GetDailyChallenges(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	challenges, festival, err := h.service.GetTodaysChallenges(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get challenges", zap.Error(err))
		response.InternalError(w, "failed to get challenges")
		return
	}

	resp := map[string]interface{}{
		"challenges": challenges,
		"date":       challenges[0].AssignedAt,
	}

	if festival != nil {
		resp["festival"] = map[string]interface{}{
			"name":             festival.Name,
			"name_hindi":       festival.NameHindi,
			"description":      festival.Description,
			"karma_multiplier": festival.KarmaMultiplier,
			"bonus_karma":      festival.BonusKarma,
			"ends_at":          festival.EndDate,
		}
	}

	response.OK(w, resp)
}

// GetFestivalStatus returns active festival if any
func (h *ChallengeHandler) GetFestivalStatus(w http.ResponseWriter, r *http.Request) {
	festival, err := h.service.GetFestivalStatus(r.Context())
	if err != nil || festival == nil {
		response.OK(w, map[string]interface{}{
			"active_festival": nil,
			"message":         "No active festival",
		})
		return
	}

	response.OK(w, map[string]interface{}{
		"active_festival": festival,
	})
}

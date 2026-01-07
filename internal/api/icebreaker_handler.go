package api

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/locolive/backend/internal/domain"
	"github.com/locolive/backend/internal/middleware"
	"github.com/locolive/backend/pkg/response"
	"go.uber.org/zap"
)

type IcebreakerHandler struct {
	service *domain.IcebreakerService
	logger  *zap.Logger
}

func NewIcebreakerHandler(service *domain.IcebreakerService, logger *zap.Logger) *IcebreakerHandler {
	return &IcebreakerHandler{service: service, logger: logger}
}

// GetTodaysMatch returns today's match suggestion
func (h *IcebreakerHandler) GetTodaysMatch(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	city := r.URL.Query().Get("city")
	var cityPtr *string
	if city != "" {
		cityPtr = &city
	}

	match, err := h.service.GetTodaysMatch(r.Context(), userID, cityPtr)
	if err != nil {
		h.logger.Error("Failed to get match", zap.Error(err))
		response.InternalError(w, "failed to get match")
		return
	}

	if match == nil {
		response.OK(w, map[string]interface{}{
			"match":   nil,
			"message": "No match available today",
		})
		return
	}

	response.OK(w, map[string]interface{}{
		"match": match,
	})
}

type RespondRequest struct {
	Accept bool `json:"accept"`
}

// RespondToMatch handles accept/decline
func (h *IcebreakerHandler) RespondToMatch(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	matchIDStr := chi.URLParam(r, "matchId")
	matchID, err := uuid.Parse(matchIDStr)
	if err != nil {
		response.BadRequest(w, "invalid match ID")
		return
	}

	var req RespondRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	karmaEarned, err := h.service.RespondToMatch(r.Context(), userID, matchID, req.Accept)
	if err != nil {
		h.logger.Error("Failed to respond to match", zap.Error(err))
		response.InternalError(w, "failed to respond")
		return
	}

	message := "Maybe next time!"
	if req.Accept {
		message = "Connected successfully!"
	}

	response.OK(w, map[string]interface{}{
		"message":      message,
		"karma_earned": karmaEarned,
	})
}

// GetMyInterests returns user's interests
func (h *IcebreakerHandler) GetMyInterests(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	interests, err := h.service.GetUserInterests(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get interests", zap.Error(err))
		response.InternalError(w, "failed to get interests")
		return
	}

	response.OK(w, map[string]interface{}{
		"interests": interests,
	})
}

type AddInterestRequest struct {
	Interest string `json:"interest"`
	Category string `json:"category"`
}

// AddInterest adds an interest
func (h *IcebreakerHandler) AddInterest(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	var req AddInterestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Interest == "" {
		response.BadRequest(w, "interest is required")
		return
	}

	if err := h.service.AddInterest(r.Context(), userID, req.Interest, req.Category); err != nil {
		h.logger.Error("Failed to add interest", zap.Error(err))
		response.InternalError(w, "failed to add interest")
		return
	}

	response.OK(w, map[string]string{"message": "interest added"})
}

// RemoveInterest removes an interest
func (h *IcebreakerHandler) RemoveInterest(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	interest := chi.URLParam(r, "interest")
	if interest == "" {
		response.BadRequest(w, "interest is required")
		return
	}

	if err := h.service.RemoveInterest(r.Context(), userID, interest); err != nil {
		h.logger.Error("Failed to remove interest", zap.Error(err))
		response.InternalError(w, "failed to remove interest")
		return
	}

	response.OK(w, map[string]string{"message": "interest removed"})
}

// GetSuggestedInterests returns popular interests
func (h *IcebreakerHandler) GetSuggestedInterests(w http.ResponseWriter, r *http.Request) {
	suggestions, err := h.service.GetSuggestedInterests(r.Context())
	if err != nil {
		h.logger.Error("Failed to get suggestions", zap.Error(err))
		response.InternalError(w, "failed to get suggestions")
		return
	}

	response.OK(w, map[string]interface{}{
		"suggestions": suggestions,
	})
}

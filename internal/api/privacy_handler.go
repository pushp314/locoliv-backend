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

type PrivacyHandler struct {
	service *domain.PrivacyService
	logger  *zap.Logger
}

func NewPrivacyHandler(service *domain.PrivacyService, logger *zap.Logger) *PrivacyHandler {
	return &PrivacyHandler{service: service, logger: logger}
}

// GetPrivacySettings returns user's privacy settings
func (h *PrivacyHandler) GetSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	settings, err := h.service.GetSettings(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get privacy settings", zap.Error(err))
		response.InternalError(w, "failed to get settings")
		return
	}

	response.OK(w, settings)
}

type UpdateSettingsRequest struct {
	DefaultStoryVisibility string `json:"default_story_visibility"` // public, connections, family
	ShowInDiscovery        *bool  `json:"show_in_discovery,omitempty"`
	ShowLocation           *bool  `json:"show_location,omitempty"`
	WhoCanMessage          string `json:"who_can_message,omitempty"` // everyone, connections, nobody
	IncognitoMode          *bool  `json:"incognito_mode,omitempty"`
}

// UpdateSettings updates privacy settings
func (h *PrivacyHandler) UpdateSettings(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	var req UpdateSettingsRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Get current settings
	settings, err := h.service.GetSettings(r.Context(), userID)
	if err != nil {
		response.InternalError(w, "failed to get settings")
		return
	}

	// Apply updates
	if req.DefaultStoryVisibility != "" {
		settings.DefaultStoryVisibility = domain.StoryVisibility(req.DefaultStoryVisibility)
	}
	if req.ShowInDiscovery != nil {
		settings.ShowInDiscovery = *req.ShowInDiscovery
	}
	if req.ShowLocation != nil {
		settings.ShowLocation = *req.ShowLocation
	}
	if req.WhoCanMessage != "" {
		settings.WhoCanMessage = req.WhoCanMessage
	}
	if req.IncognitoMode != nil {
		settings.IncognitoMode = *req.IncognitoMode
	}

	if err := h.service.UpdateSettings(r.Context(), settings); err != nil {
		h.logger.Error("Failed to update settings", zap.Error(err))
		response.InternalError(w, "failed to update settings")
		return
	}

	response.OK(w, settings)
}

type BlockUserRequest struct {
	UserID uuid.UUID `json:"user_id"`
	Reason *string   `json:"reason,omitempty"`
}

// BlockUser blocks another user
func (h *PrivacyHandler) BlockUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	var req BlockUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.UserID == uuid.Nil {
		response.BadRequest(w, "user_id is required")
		return
	}

	if err := h.service.BlockUser(r.Context(), userID, req.UserID, req.Reason); err != nil {
		h.logger.Error("Failed to block user", zap.Error(err))
		response.InternalError(w, "failed to block user")
		return
	}

	response.OK(w, map[string]string{"message": "user blocked"})
}

// UnblockUser unblocks a user
func (h *PrivacyHandler) UnblockUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	blockedIDStr := chi.URLParam(r, "userId")
	blockedID, err := uuid.Parse(blockedIDStr)
	if err != nil {
		response.BadRequest(w, "invalid user ID")
		return
	}

	if err := h.service.UnblockUser(r.Context(), userID, blockedID); err != nil {
		h.logger.Error("Failed to unblock user", zap.Error(err))
		response.InternalError(w, "failed to unblock user")
		return
	}

	response.OK(w, map[string]string{"message": "user unblocked"})
}

// GetBlockedUsers returns list of blocked users
func (h *PrivacyHandler) GetBlockedUsers(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	blockedIDs, err := h.service.GetBlockedUsers(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get blocked users", zap.Error(err))
		response.InternalError(w, "failed to get blocked users")
		return
	}

	response.OK(w, map[string]interface{}{"blocked_users": blockedIDs})
}

// AddToFamily adds a user to family list
func (h *PrivacyHandler) AddToFamily(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	friendIDStr := chi.URLParam(r, "userId")
	friendID, err := uuid.Parse(friendIDStr)
	if err != nil {
		response.BadRequest(w, "invalid user ID")
		return
	}

	if err := h.service.AddToFamily(r.Context(), userID, friendID); err != nil {
		h.logger.Error("Failed to add to family", zap.Error(err))
		response.InternalError(w, "failed to add to family")
		return
	}

	response.OK(w, map[string]string{"message": "added to family list"})
}

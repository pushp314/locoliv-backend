package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/locolive/backend/internal/domain"
	"github.com/locolive/backend/internal/middleware"
	"github.com/locolive/backend/pkg/response"
	"go.uber.org/zap"
)

type NotificationHandler struct {
	service *domain.NotificationService
	logger  *zap.Logger
}

func NewNotificationHandler(service *domain.NotificationService, logger *zap.Logger) *NotificationHandler {
	return &NotificationHandler{
		service: service,
		logger:  logger,
	}
}

func (h *NotificationHandler) GetNotifications(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if page < 1 {
		page = 1
	}
	offset := (page - 1) * limit

	notifs, err := h.service.GetNotifications(r.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error("failed to get notifications", zap.Error(err))
		response.InternalError(w, "failed to fetch notifications")
		return
	}

	response.OK(w, notifs)
}

func (h *NotificationHandler) MarkRead(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	idStr := chi.URLParam(r, "id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		response.BadRequest(w, "invalid notification id")
		return
	}

	if err := h.service.MarkRead(r.Context(), userID, id); err != nil {
		h.logger.Error("failed to mark notification read", zap.Error(err))
		response.InternalError(w, "failed to update notification")
		return
	}

	response.OK(w, map[string]string{"status": "success"})
}

func (h *NotificationHandler) UpdateFCMToken(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	sessionID, ok := middleware.GetSessionID(r.Context())
	if !ok {
		response.Unauthorized(w, "no session")
		return
	}

	var req struct {
		FCMToken string `json:"fcm_token"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request")
		return
	}

	if err := h.service.UpdateFCMToken(r.Context(), sessionID, req.FCMToken); err != nil {
		h.logger.Error("failed to update fcm token", zap.String("user_id", userID.String()), zap.Error(err))
		response.InternalError(w, "failed to update token")
		return
	}

	response.OK(w, map[string]string{"status": "success"})
}

package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/locolive/backend/internal/domain"
	"github.com/locolive/backend/internal/middleware"
	"github.com/locolive/backend/pkg/response"
	"go.uber.org/zap"
)

type ConnectionHandler struct {
	connService *domain.ConnectionService
	logger      *zap.Logger
}

func NewConnectionHandler(connService *domain.ConnectionService, logger *zap.Logger) *ConnectionHandler {
	return &ConnectionHandler{
		connService: connService,
		logger:      logger,
	}
}

// SendRequest handles POST /connections/request
func (h *ConnectionHandler) SendRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	var req struct {
		TargetUserID string `json:"target_user_id"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request")
		return
	}

	targetID, err := uuid.Parse(req.TargetUserID)
	if err != nil {
		response.BadRequest(w, "invalid target user id")
		return
	}

	conn, err := h.connService.SendRequest(r.Context(), userID, targetID)
	if err != nil {
		h.logger.Error("failed to send connection request", zap.Error(err))
		response.InternalError(w, "failed to send request")
		return
	}

	response.OK(w, conn)
}

// RespondRequest handles POST /connections/respond
func (h *ConnectionHandler) RespondRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	var req struct {
		ConnectionID string `json:"connection_id"`
		Accept       bool   `json:"accept"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request")
		return
	}

	connID, err := uuid.Parse(req.ConnectionID)
	if err != nil {
		response.BadRequest(w, "invalid connection id")
		return
	}

	conn, err := h.connService.RespondToRequest(r.Context(), userID, connID, req.Accept)
	if err != nil {
		h.logger.Error("failed to respond to request", zap.Error(err))
		response.InternalError(w, "failed to respond")
		return
	}

	response.OK(w, conn)
}

// GetConnections handles GET /connections
func (h *ConnectionHandler) GetConnections(w http.ResponseWriter, r *http.Request) {
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

	conns, err := h.connService.GetConnections(r.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error("failed to get connections", zap.Error(err))
		response.InternalError(w, "failed to get connections")
		return
	}

	response.OK(w, conns)
}

// GetRequests handles GET /connections/requests
func (h *ConnectionHandler) GetRequests(w http.ResponseWriter, r *http.Request) {
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

	conns, err := h.connService.GetPendingRequests(r.Context(), userID, limit, offset)
	if err != nil {
		h.logger.Error("failed to get requests", zap.Error(err))
		response.InternalError(w, "failed to get requests")
		return
	}

	response.OK(w, conns)
}

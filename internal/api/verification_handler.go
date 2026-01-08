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

// VerificationHandler handles verification requests
type VerificationHandler struct {
	service *domain.VerificationService
	logger  *zap.Logger
}

// NewVerificationHandler creates a new verification handler
func NewVerificationHandler(service *domain.VerificationService, logger *zap.Logger) *VerificationHandler {
	return &VerificationHandler{
		service: service,
		logger:  logger,
	}
}

// SubmitRequest handles verification request submission
func (h *VerificationHandler) SubmitRequest(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	type SubmitReq struct {
		RequestType  string  `json:"request_type"`
		DocumentURL  *string `json:"document_url,omitempty"`
		DocumentType *string `json:"document_type,omitempty"`
	}

	var req SubmitReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate request type
	validTypes := map[string]bool{
		"blue": true, "celebrity": true, "official": true, "business": true,
	}
	if !validTypes[req.RequestType] {
		response.BadRequest(w, "invalid request type")
		return
	}

	result, err := h.service.SubmitRequest(
		r.Context(),
		userID,
		domain.VerificationType(req.RequestType),
		req.DocumentURL,
		req.DocumentType,
	)
	if err != nil {
		h.logger.Error("Failed to submit verification request", zap.Error(err))
		response.InternalError(w, "failed to submit request")
		return
	}

	response.Created(w, result)
}

// GetMyRequests returns user's verification requests
func (h *VerificationHandler) GetMyRequests(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	requests, err := h.service.GetMyRequests(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get verification requests", zap.Error(err))
		response.InternalError(w, "failed to get requests")
		return
	}

	response.OK(w, requests)
}

// GetMyStatus returns user's verification status
func (h *VerificationHandler) GetMyStatus(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	status, err := h.service.GetUserVerification(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get verification status", zap.Error(err))
		response.InternalError(w, "failed to get status")
		return
	}

	response.OK(w, status)
}

// GetUserVerification returns a specific user's verification status
func (h *VerificationHandler) GetUserVerification(w http.ResponseWriter, r *http.Request) {
	userIDStr := chi.URLParam(r, "userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		response.BadRequest(w, "invalid user id")
		return
	}

	status, err := h.service.GetUserVerification(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get user verification", zap.Error(err))
		response.InternalError(w, "failed to get verification")
		return
	}

	response.OK(w, status)
}

// GetPendingRequests returns pending requests (admin only)
func (h *VerificationHandler) GetPendingRequests(w http.ResponseWriter, r *http.Request) {
	// TODO: Add admin authorization check

	limit := 50
	offset := 0

	requests, err := h.service.GetPendingRequests(r.Context(), limit, offset)
	if err != nil {
		h.logger.Error("Failed to get pending requests", zap.Error(err))
		response.InternalError(w, "failed to get requests")
		return
	}

	response.OK(w, requests)
}

// ApproveRequest approves a verification request (admin only)
func (h *VerificationHandler) ApproveRequest(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	// TODO: Add admin authorization check

	reqIDStr := chi.URLParam(r, "requestId")
	reqID, err := uuid.Parse(reqIDStr)
	if err != nil {
		response.BadRequest(w, "invalid request id")
		return
	}

	if err := h.service.ApproveRequest(r.Context(), reqID, adminID); err != nil {
		h.logger.Error("Failed to approve request", zap.Error(err))
		response.InternalError(w, "failed to approve request")
		return
	}

	response.OK(w, map[string]string{"message": "request approved"})
}

// RejectRequest rejects a verification request (admin only)
func (h *VerificationHandler) RejectRequest(w http.ResponseWriter, r *http.Request) {
	adminID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	// TODO: Add admin authorization check

	reqIDStr := chi.URLParam(r, "requestId")
	reqID, err := uuid.Parse(reqIDStr)
	if err != nil {
		response.BadRequest(w, "invalid request id")
		return
	}

	type RejectReq struct {
		Reason string `json:"reason"`
	}
	var req RejectReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if err := h.service.RejectRequest(r.Context(), reqID, adminID, req.Reason); err != nil {
		h.logger.Error("Failed to reject request", zap.Error(err))
		response.InternalError(w, "failed to reject request")
		return
	}

	response.OK(w, map[string]string{"message": "request rejected"})
}

// RegisterVerificationRoutes registers verification routes
func RegisterVerificationRoutes(r chi.Router, h *VerificationHandler) {
	r.Route("/verification", func(r chi.Router) {
		r.Get("/status", h.GetMyStatus)
		r.Get("/requests", h.GetMyRequests)
		r.Post("/request", h.SubmitRequest)
	})

	// Admin routes
	r.Route("/admin/verification", func(r chi.Router) {
		r.Get("/pending", h.GetPendingRequests)
		r.Post("/requests/{requestId}/approve", h.ApproveRequest)
		r.Post("/requests/{requestId}/reject", h.RejectRequest)
	})
}

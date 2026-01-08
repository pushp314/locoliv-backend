package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/locolive/backend/internal/middleware"
	"github.com/locolive/backend/internal/storage"
	"github.com/locolive/backend/pkg/response"
	"go.uber.org/zap"
)

// MediaHandler handles media upload requests
type MediaHandler struct {
	storage *storage.R2Storage
	logger  *zap.Logger
}

// NewMediaHandler creates a new media handler
func NewMediaHandler(storage *storage.R2Storage, logger *zap.Logger) *MediaHandler {
	return &MediaHandler{
		storage: storage,
		logger:  logger,
	}
}

// PresignRequest is the request body for generating a presigned URL
type PresignRequest struct {
	Filename    string `json:"filename"`
	ContentType string `json:"content_type"`
	MediaType   string `json:"media_type"` // story, avatar, chat, event, document
}

// PresignResponse is the response containing presigned URLs
type PresignResponse struct {
	UploadURL string    `json:"upload_url"`
	PublicURL string    `json:"public_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// GeneratePresignedURL creates a presigned URL for direct client upload
func (h *MediaHandler) GeneratePresignedURL(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	if h.storage == nil {
		response.InternalError(w, "storage not configured")
		return
	}

	var req PresignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Filename == "" || req.ContentType == "" {
		response.BadRequest(w, "filename and content_type are required")
		return
	}

	// Validate content type
	if !isAllowedContentType(req.ContentType) {
		response.BadRequest(w, "unsupported content type")
		return
	}

	// Generate presigned URL (valid for 15 minutes)
	expiry := 15 * time.Minute
	uploadURL, publicURL, err := h.storage.GenerateUploadURL(r.Context(), req.Filename, req.ContentType, expiry)
	if err != nil {
		h.logger.Error("Failed to generate presigned URL", zap.Error(err))
		response.InternalError(w, "failed to generate upload URL")
		return
	}

	response.OK(w, PresignResponse{
		UploadURL: uploadURL,
		PublicURL: publicURL,
		ExpiresAt: time.Now().Add(expiry),
	})
}

// DeleteMedia deletes a media file
func (h *MediaHandler) DeleteMedia(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	if h.storage == nil {
		response.InternalError(w, "storage not configured")
		return
	}

	type DeleteRequest struct {
		URL string `json:"url"`
	}

	var req DeleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.URL == "" {
		response.BadRequest(w, "url is required")
		return
	}

	if err := h.storage.DeleteFile(r.Context(), req.URL); err != nil {
		h.logger.Error("Failed to delete file", zap.Error(err), zap.String("url", req.URL))
		response.InternalError(w, "failed to delete file")
		return
	}

	response.NoContent(w)
}

// isAllowedContentType checks if the content type is allowed for upload
func isAllowedContentType(contentType string) bool {
	allowed := map[string]bool{
		"image/jpeg":      true,
		"image/png":       true,
		"image/gif":       true,
		"image/webp":      true,
		"video/mp4":       true,
		"video/quicktime": true,
		"video/webm":      true,
		"application/pdf": true,
	}
	return allowed[contentType]
}

// RegisterMediaRoutes registers media routes
func RegisterMediaRoutes(r chi.Router, h *MediaHandler) {
	r.Route("/media", func(r chi.Router) {
		r.Post("/presign", h.GeneratePresignedURL)
		r.Delete("/", h.DeleteMedia)
	})
}

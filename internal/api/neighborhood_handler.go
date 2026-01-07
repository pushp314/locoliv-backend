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

type NeighborhoodHandler struct {
	service *domain.NeighborhoodService
	logger  *zap.Logger
}

func NewNeighborhoodHandler(service *domain.NeighborhoodService, logger *zap.Logger) *NeighborhoodHandler {
	return &NeighborhoodHandler{service: service, logger: logger}
}

type CreatePostRequest struct {
	PostType string  `json:"post_type"` // question, recommendation, help, event
	Title    string  `json:"title"`
	Content  *string `json:"content,omitempty"`
	City     *string `json:"city,omitempty"`
	RadiusKm int     `json:"radius_km"`
}

// CreatePost creates a new neighborhood post
func (h *NeighborhoodHandler) CreatePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	var req CreatePostRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Title == "" {
		response.BadRequest(w, "title is required")
		return
	}

	postType := domain.PostTypeQuestion
	switch req.PostType {
	case "recommendation":
		postType = domain.PostTypeRecommendation
	case "help":
		postType = domain.PostTypeHelp
	case "event":
		postType = domain.PostTypeEvent
	}

	radiusKm := req.RadiusKm
	if radiusKm == 0 {
		radiusKm = 5
	}

	post, err := h.service.CreatePost(r.Context(), domain.CreateNeighborhoodPostParams{
		UserID:   userID,
		PostType: postType,
		Title:    req.Title,
		Content:  req.Content,
		City:     req.City,
		RadiusKm: radiusKm,
	})
	if err != nil {
		h.logger.Error("Failed to create post", zap.Error(err))
		response.InternalError(w, "failed to create post")
		return
	}

	response.Created(w, post)
}

// GetPosts returns nearby posts
func (h *NeighborhoodHandler) GetPosts(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	postType := r.URL.Query().Get("type")

	if city == "" {
		city = "Unknown" // Default
	}

	var pt *string
	if postType != "" {
		pt = &postType
	}

	posts, err := h.service.GetNearbyPosts(r.Context(), city, pt, 20, 0)
	if err != nil {
		h.logger.Error("Failed to get posts", zap.Error(err))
		response.InternalError(w, "failed to get posts")
		return
	}

	response.OK(w, map[string]interface{}{
		"posts": posts,
		"city":  city,
	})
}

type CreateAnswerRequest struct {
	Content string `json:"content"`
}

// AnswerPost adds an answer to a post
func (h *NeighborhoodHandler) AnswerPost(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	postIDStr := chi.URLParam(r, "postId")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		response.BadRequest(w, "invalid post ID")
		return
	}

	var req CreateAnswerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Content == "" {
		response.BadRequest(w, "content is required")
		return
	}

	answer, err := h.service.AnswerPost(r.Context(), postID, userID, req.Content)
	if err != nil {
		h.logger.Error("Failed to create answer", zap.Error(err))
		response.InternalError(w, "failed to create answer")
		return
	}

	response.Created(w, answer)
}

// UpvotePost upvotes a post
func (h *NeighborhoodHandler) UpvotePost(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	postIDStr := chi.URLParam(r, "postId")
	postID, err := uuid.Parse(postIDStr)
	if err != nil {
		response.BadRequest(w, "invalid post ID")
		return
	}

	if err := h.service.UpvotePost(r.Context(), postID, userID); err != nil {
		h.logger.Error("Failed to upvote", zap.Error(err))
		response.InternalError(w, "failed to upvote")
		return
	}

	response.OK(w, map[string]string{"message": "upvoted"})
}

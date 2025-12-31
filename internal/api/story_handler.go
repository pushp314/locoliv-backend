package api

import (
	"net/http"
	"strconv"

	"github.com/locolive/backend/internal/domain"
	"github.com/locolive/backend/internal/middleware"
	"github.com/locolive/backend/pkg/response"
	"go.uber.org/zap"
)

type StoryHandler struct {
	storyService *domain.StoryService
	logger       *zap.Logger
}

func NewStoryHandler(storyService *domain.StoryService, logger *zap.Logger) *StoryHandler {
	return &StoryHandler{
		storyService: storyService,
		logger:       logger,
	}
}

// CreateStory handles creating a new story
func (h *StoryHandler) CreateStory(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	// Parse multipart form (max 10MB)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.BadRequest(w, "invalid form data")
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		response.BadRequest(w, "missing file")
		return
	}
	defer file.Close()

	caption := r.FormValue("caption")
	mediaType := r.FormValue("media_type")
	if mediaType == "" {
		mediaType = "image" // Default
	}

	var lat, lng *float64
	if latStr := r.FormValue("lat"); latStr != "" {
		if val, err := strconv.ParseFloat(latStr, 64); err == nil {
			lat = &val
		}
	}
	if lngStr := r.FormValue("lng"); lngStr != "" {
		if val, err := strconv.ParseFloat(lngStr, 64); err == nil {
			lng = &val
		}
	}

	params := domain.CreateStoryParams{
		UserID:      userID,
		MediaType:   mediaType,
		Caption:     &caption,
		LocationLat: lat,
		LocationLng: lng,
	}

	story, err := h.storyService.CreateStory(r.Context(), params, file, header.Filename, header.Header.Get("Content-Type"))
	if err != nil {
		h.logger.Error("create story failed", zap.Error(err))
		response.InternalError(w, "failed to create story")
		return
	}

	response.Created(w, story)
}

// GetFeed handles fetching the story feed
func (h *StoryHandler) GetFeed(w http.ResponseWriter, r *http.Request) {
	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)

	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)

	var lat, lng, radius *float64
	if latStr := r.URL.Query().Get("lat"); latStr != "" {
		if val, err := strconv.ParseFloat(latStr, 64); err == nil {
			lat = &val
		}
	}
	if lngStr := r.URL.Query().Get("lng"); lngStr != "" {
		if val, err := strconv.ParseFloat(lngStr, 64); err == nil {
			lng = &val
		}
	}
	if radiumStr := r.URL.Query().Get("radius"); radiumStr != "" {
		if val, err := strconv.ParseFloat(radiumStr, 64); err == nil {
			radius = &val
		}
	} else if lat != nil && lng != nil {
		// Default radius 5km if location provided but no radius
		r := 5000.0
		radius = &r
	}

	stories, err := h.storyService.GetFeed(r.Context(), page, limit, lat, lng, radius)
	if err != nil {
		h.logger.Error("get feed failed", zap.Error(err))
		response.InternalError(w, "failed to get feed")
		return
	}

	response.OK(w, stories)
}

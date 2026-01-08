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

// DiscoveryHandler handles location-based discovery
type DiscoveryHandler struct {
	service *domain.DiscoveryService
	logger  *zap.Logger
}

// NewDiscoveryHandler creates a new discovery handler
func NewDiscoveryHandler(service *domain.DiscoveryService, logger *zap.Logger) *DiscoveryHandler {
	return &DiscoveryHandler{
		service: service,
		logger:  logger,
	}
}

// NearbyUsersRequest is the request for nearby users
type NearbyUsersRequest struct {
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	RadiusKm  float64 `json:"radius_km"`
	Limit     int     `json:"limit"`
}

// GetNearbyUsers returns users near a location
func (h *DiscoveryHandler) GetNearbyUsers(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	// Get query parameters
	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	radiusKm, _ := strconv.ParseFloat(r.URL.Query().Get("radius"), 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if lat == 0 || lng == 0 {
		response.BadRequest(w, "latitude and longitude are required")
		return
	}

	if radiusKm <= 0 || radiusKm > 50 {
		radiusKm = 10 // Default 10km
	}
	if limit <= 0 || limit > 100 {
		limit = 20 // Default 20 users
	}

	users, err := h.service.GetNearbyUsers(r.Context(), userID, lat, lng, radiusKm, limit)
	if err != nil {
		h.logger.Error("Failed to get nearby users", zap.Error(err))
		response.InternalError(w, "failed to get nearby users")
		return
	}

	response.OK(w, users)
}

// GetNearbyStories returns stories near a location
func (h *DiscoveryHandler) GetNearbyStories(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	radiusKm, _ := strconv.ParseFloat(r.URL.Query().Get("radius"), 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if lat == 0 || lng == 0 {
		response.BadRequest(w, "latitude and longitude are required")
		return
	}

	if radiusKm <= 0 || radiusKm > 50 {
		radiusKm = 10
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	stories, err := h.service.GetNearbyStories(r.Context(), lat, lng, radiusKm, limit)
	if err != nil {
		h.logger.Error("Failed to get nearby stories", zap.Error(err))
		response.InternalError(w, "failed to get nearby stories")
		return
	}

	response.OK(w, stories)
}

// GetNearbyEvents returns events near a location
func (h *DiscoveryHandler) GetNearbyEvents(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)
	radiusKm, _ := strconv.ParseFloat(r.URL.Query().Get("radius"), 64)
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))

	if lat == 0 || lng == 0 {
		response.BadRequest(w, "latitude and longitude are required")
		return
	}

	if radiusKm <= 0 || radiusKm > 100 {
		radiusKm = 25
	}
	if limit <= 0 || limit > 50 {
		limit = 20
	}

	events, err := h.service.GetNearbyEvents(r.Context(), lat, lng, radiusKm, limit)
	if err != nil {
		h.logger.Error("Failed to get nearby events", zap.Error(err))
		response.InternalError(w, "failed to get nearby events")
		return
	}

	response.OK(w, events)
}

// UpdateLocation updates user's current location
func (h *DiscoveryHandler) UpdateLocation(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	type LocationUpdate struct {
		Latitude  float64 `json:"latitude"`
		Longitude float64 `json:"longitude"`
		City      string  `json:"city,omitempty"`
	}

	var req LocationUpdate
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Latitude == 0 || req.Longitude == 0 {
		response.BadRequest(w, "latitude and longitude are required")
		return
	}

	if err := h.service.UpdateUserLocation(r.Context(), userID, req.Latitude, req.Longitude, req.City); err != nil {
		h.logger.Error("Failed to update location", zap.Error(err))
		response.InternalError(w, "failed to update location")
		return
	}

	response.OK(w, map[string]string{"message": "location updated"})
}

// GetExploreData returns aggregated explore data
func (h *DiscoveryHandler) GetExploreData(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	lat, _ := strconv.ParseFloat(r.URL.Query().Get("lat"), 64)
	lng, _ := strconv.ParseFloat(r.URL.Query().Get("lng"), 64)

	if lat == 0 || lng == 0 {
		response.BadRequest(w, "latitude and longitude are required")
		return
	}

	data, err := h.service.GetExploreData(r.Context(), userID, lat, lng)
	if err != nil {
		h.logger.Error("Failed to get explore data", zap.Error(err))
		response.InternalError(w, "failed to get explore data")
		return
	}

	response.OK(w, data)
}

// GetUserProfile returns a user's public profile
func (h *DiscoveryHandler) GetUserProfile(w http.ResponseWriter, r *http.Request) {
	_, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	targetIDStr := chi.URLParam(r, "userId")
	targetID, err := uuid.Parse(targetIDStr)
	if err != nil {
		response.BadRequest(w, "invalid user id")
		return
	}

	profile, err := h.service.GetUserProfile(r.Context(), targetID)
	if err != nil {
		h.logger.Error("Failed to get user profile", zap.Error(err))
		response.InternalError(w, "failed to get user profile")
		return
	}

	response.OK(w, profile)
}

// RegisterDiscoveryRoutes registers discovery routes
func RegisterDiscoveryRoutes(r chi.Router, h *DiscoveryHandler) {
	r.Route("/discover", func(r chi.Router) {
		r.Get("/nearby/users", h.GetNearbyUsers)
		r.Get("/nearby/stories", h.GetNearbyStories)
		r.Get("/nearby/events", h.GetNearbyEvents)
		r.Get("/explore", h.GetExploreData)
		r.Post("/location", h.UpdateLocation)
	})

	r.Get("/users/{userId}/profile", h.GetUserProfile)
}

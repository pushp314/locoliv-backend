package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/locolive/backend/internal/domain"
	"github.com/locolive/backend/internal/middleware"
	"github.com/locolive/backend/pkg/response"
	"go.uber.org/zap"
)

type EventHandler struct {
	service *domain.EventService
	logger  *zap.Logger
}

func NewEventHandler(service *domain.EventService, logger *zap.Logger) *EventHandler {
	return &EventHandler{service: service, logger: logger}
}

type CreateEventRequest struct {
	Title        string  `json:"title"`
	TitleHindi   *string `json:"title_hindi,omitempty"`
	Description  *string `json:"description,omitempty"`
	EventType    string  `json:"event_type"`
	Venue        *string `json:"venue,omitempty"`
	City         *string `json:"city,omitempty"`
	EventDate    string  `json:"event_date"` // ISO 8601
	EndDate      *string `json:"end_date,omitempty"`
	MaxAttendees *int    `json:"max_attendees,omitempty"`
	IsPublic     *bool   `json:"is_public,omitempty"`
}

// CreateEvent creates a new local event
func (h *EventHandler) CreateEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	var req CreateEventRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Title == "" || req.EventDate == "" {
		response.BadRequest(w, "title and event_date are required")
		return
	}

	eventDate, err := time.Parse(time.RFC3339, req.EventDate)
	if err != nil {
		response.BadRequest(w, "invalid event_date format")
		return
	}

	var endDate *time.Time
	if req.EndDate != nil {
		ed, err := time.Parse(time.RFC3339, *req.EndDate)
		if err == nil {
			endDate = &ed
		}
	}

	isPublic := true
	if req.IsPublic != nil {
		isPublic = *req.IsPublic
	}

	event, err := h.service.CreateEvent(r.Context(), domain.CreateEventParams{
		CreatorID:    userID,
		Title:        req.Title,
		TitleHindi:   req.TitleHindi,
		Description:  req.Description,
		EventType:    domain.EventType(req.EventType),
		Venue:        req.Venue,
		City:         req.City,
		EventDate:    eventDate,
		EndDate:      endDate,
		MaxAttendees: req.MaxAttendees,
		IsPublic:     isPublic,
	})
	if err != nil {
		h.logger.Error("Failed to create event", zap.Error(err))
		response.InternalError(w, "failed to create event")
		return
	}

	response.Created(w, event)
}

// GetEvents returns events by city/type
func (h *EventHandler) GetEvents(w http.ResponseWriter, r *http.Request) {
	city := r.URL.Query().Get("city")
	eventType := r.URL.Query().Get("type")

	var et *string
	if eventType != "" {
		et = &eventType
	}

	events, err := h.service.GetEvents(r.Context(), city, et, 20, 0)
	if err != nil {
		h.logger.Error("Failed to get events", zap.Error(err))
		response.InternalError(w, "failed to get events")
		return
	}

	response.OK(w, map[string]interface{}{
		"events": events,
		"city":   city,
	})
}

// GetEvent returns single event details
func (h *EventHandler) GetEvent(w http.ResponseWriter, r *http.Request) {
	userID, _ := middleware.GetUserID(r.Context())

	eventIDStr := chi.URLParam(r, "eventId")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		response.BadRequest(w, "invalid event ID")
		return
	}

	event, err := h.service.GetEvent(r.Context(), eventID, userID)
	if err != nil {
		h.logger.Error("Failed to get event", zap.Error(err))
		response.NotFound(w, "event not found")
		return
	}

	response.OK(w, event)
}

type RSVPRequest struct {
	Status string `json:"status"` // going, interested, not_going
}

// RSVPEvent marks user's attendance
func (h *EventHandler) RSVPEvent(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	eventIDStr := chi.URLParam(r, "eventId")
	eventID, err := uuid.Parse(eventIDStr)
	if err != nil {
		response.BadRequest(w, "invalid event ID")
		return
	}

	var req RSVPRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Status != "going" && req.Status != "interested" && req.Status != "not_going" {
		response.BadRequest(w, "status must be going, interested, or not_going")
		return
	}

	if err := h.service.RSVPEvent(r.Context(), eventID, userID, req.Status); err != nil {
		h.logger.Error("Failed to RSVP", zap.Error(err))
		response.InternalError(w, "failed to RSVP")
		return
	}

	response.OK(w, map[string]string{
		"message": "RSVP recorded",
		"status":  req.Status,
	})
}

// GetMyEvents returns user's created and attending events
func (h *EventHandler) GetMyEvents(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "user not authenticated")
		return
	}

	created, attending, err := h.service.GetMyEvents(r.Context(), userID)
	if err != nil {
		h.logger.Error("Failed to get my events", zap.Error(err))
		response.InternalError(w, "failed to get events")
		return
	}

	response.OK(w, map[string]interface{}{
		"created":   created,
		"attending": attending,
	})
}

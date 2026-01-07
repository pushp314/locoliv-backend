package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// EventType categorizes events
type EventType string

const (
	EventTypeGeneral   EventType = "general"
	EventTypeSports    EventType = "sports"
	EventTypeReligious EventType = "religious"
	EventTypeSocial    EventType = "social"
	EventTypeParty     EventType = "party"
)

// LocalEvent represents a community event
type LocalEvent struct {
	ID            uuid.UUID     `json:"id"`
	CreatorID     uuid.UUID     `json:"creator_id"`
	Title         string        `json:"title"`
	TitleHindi    *string       `json:"title_hindi,omitempty"`
	Description   *string       `json:"description,omitempty"`
	EventType     EventType     `json:"event_type"`
	LocationLat   *float64      `json:"location_lat,omitempty"`
	LocationLng   *float64      `json:"location_lng,omitempty"`
	Venue         *string       `json:"venue,omitempty"`
	City          *string       `json:"city,omitempty"`
	EventDate     time.Time     `json:"event_date"`
	EndDate       *time.Time    `json:"end_date,omitempty"`
	MaxAttendees  *int          `json:"max_attendees,omitempty"`
	IsPublic      bool          `json:"is_public"`
	CoverImageURL *string       `json:"cover_image_url,omitempty"`
	KarmaReward   int           `json:"karma_reward"`
	CreatedAt     time.Time     `json:"created_at"`
	Creator       *UserResponse `json:"creator,omitempty"`
	RSVPCount     int           `json:"rsvp_count"`
	UserRSVP      *string       `json:"user_rsvp_status,omitempty"`
}

// EventRSVP represents a user's RSVP to an event
type EventRSVP struct {
	ID        uuid.UUID `json:"id"`
	EventID   uuid.UUID `json:"event_id"`
	UserID    uuid.UUID `json:"user_id"`
	Status    string    `json:"status"` // going, interested, not_going
	CreatedAt time.Time `json:"created_at"`
}

// CreateEventParams for creating events
type CreateEventParams struct {
	CreatorID    uuid.UUID
	Title        string
	TitleHindi   *string
	Description  *string
	EventType    EventType
	Venue        *string
	City         *string
	EventDate    time.Time
	EndDate      *time.Time
	MaxAttendees *int
	IsPublic     bool
}

// EventRepository interface
type EventRepository interface {
	CreateEvent(ctx context.Context, params CreateEventParams) (*LocalEvent, error)
	GetEvent(ctx context.Context, eventID uuid.UUID) (*LocalEvent, error)
	GetEventsByCity(ctx context.Context, city string, eventType *string, limit, offset int) ([]LocalEvent, error)
	GetUpcomingEvents(ctx context.Context, limit int) ([]LocalEvent, error)
	GetUserEvents(ctx context.Context, userID uuid.UUID) (created []LocalEvent, attending []LocalEvent, err error)
	UpdateEvent(ctx context.Context, eventID uuid.UUID, params CreateEventParams) error
	DeleteEvent(ctx context.Context, eventID uuid.UUID) error

	// RSVPs
	RSVPEvent(ctx context.Context, eventID, userID uuid.UUID, status string) error
	GetEventRSVPs(ctx context.Context, eventID uuid.UUID) ([]EventRSVP, error)
	GetUserRSVP(ctx context.Context, eventID, userID uuid.UUID) (*string, error)
	GetRSVPCount(ctx context.Context, eventID uuid.UUID) (int, error)
}

// EventService handles event operations
type EventService struct {
	repo         EventRepository
	karmaService *KarmaService
}

// NewEventService creates a new event service
func NewEventService(repo EventRepository, karmaService *KarmaService) *EventService {
	return &EventService{
		repo:         repo,
		karmaService: karmaService,
	}
}

// CreateEvent creates a new local event
func (s *EventService) CreateEvent(ctx context.Context, params CreateEventParams) (*LocalEvent, error) {
	event, err := s.repo.CreateEvent(ctx, params)
	if err != nil {
		return nil, err
	}

	// Award karma for creating event
	refType := "event"
	_ = s.karmaService.EarnKarma(ctx, params.CreatorID, "create_event", &refType, &event.ID)

	return event, nil
}

// GetEvents returns events by city/type
func (s *EventService) GetEvents(ctx context.Context, city string, eventType *string, limit, offset int) ([]LocalEvent, error) {
	if city == "" {
		return s.repo.GetUpcomingEvents(ctx, limit)
	}
	return s.repo.GetEventsByCity(ctx, city, eventType, limit, offset)
}

// GetEvent returns a single event with user's RSVP status
func (s *EventService) GetEvent(ctx context.Context, eventID, userID uuid.UUID) (*LocalEvent, error) {
	event, err := s.repo.GetEvent(ctx, eventID)
	if err != nil {
		return nil, err
	}

	// Get RSVP count and user status
	event.RSVPCount, _ = s.repo.GetRSVPCount(ctx, eventID)
	event.UserRSVP, _ = s.repo.GetUserRSVP(ctx, eventID, userID)

	return event, nil
}

// RSVPEvent marks user's attendance
func (s *EventService) RSVPEvent(ctx context.Context, eventID, userID uuid.UUID, status string) error {
	if err := s.repo.RSVPEvent(ctx, eventID, userID, status); err != nil {
		return err
	}

	// Award karma for attending
	if status == "going" {
		refType := "event_rsvp"
		_ = s.karmaService.EarnKarma(ctx, userID, "event_rsvp", &refType, &eventID)
	}

	return nil
}

// GetMyEvents returns user's created and attending events
func (s *EventService) GetMyEvents(ctx context.Context, userID uuid.UUID) (created []LocalEvent, attending []LocalEvent, err error) {
	return s.repo.GetUserEvents(ctx, userID)
}

package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Notification struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	Type      string    `json:"type"`
	Title     string    `json:"title"`
	Body      string    `json:"body"`
	Data      Map       `json:"data"` // leveraging the Map type or map[string]interface{}
	IsRead    bool      `json:"is_read"`
	CreatedAt time.Time `json:"created_at"`
}

// Map alias for JSONB data
type Map map[string]interface{}

type NotificationRepository interface {
	CreateNotification(ctx context.Context, userID uuid.UUID, typeStr, title, body string, data map[string]interface{}) error
	GetNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Notification, error)
	MarkNotificationRead(ctx context.Context, notificationID uuid.UUID) error
	UpdateSessionFCMToken(ctx context.Context, sessionID uuid.UUID, fcmToken string) error
	GetFCMTokens(ctx context.Context, userID uuid.UUID) ([]string, error)
}

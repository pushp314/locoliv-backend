package domain

import (
	"context"
	"fmt"
	"log"

	"github.com/google/uuid"
	"github.com/locolive/backend/internal/fcm"
)

type NotificationService struct {
	repo      NotificationRepository
	fcmClient *fcm.Client
}

func NewNotificationService(repo NotificationRepository, fcmClient *fcm.Client) *NotificationService {
	return &NotificationService{
		repo:      repo,
		fcmClient: fcmClient,
	}
}

func (s *NotificationService) GetNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Notification, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.GetNotifications(ctx, userID, limit, offset)
}

func (s *NotificationService) MarkRead(ctx context.Context, userID, notificationID uuid.UUID) error {
	return s.repo.MarkNotificationRead(ctx, notificationID)
}

func (s *NotificationService) SendNotification(ctx context.Context, userID uuid.UUID, typeStr, title, body string, data map[string]interface{}) error {
	// 1. Create in DB
	err := s.repo.CreateNotification(ctx, userID, typeStr, title, body, data)
	if err != nil {
		return err
	}

	// 2. Send push if client available
	if s.fcmClient != nil {
		// Convert map[string]interface{} to map[string]string for FCM
		strData := make(map[string]string)
		for k, v := range data {
			strData[k] = fmt.Sprintf("%v", v)
		}
		strData["type"] = typeStr

		tokens, err := s.repo.GetFCMTokens(ctx, userID)
		if err != nil {
			log.Printf("failed to get fcm tokens: %v", err)
			return nil // Don't fail the operation
		}

		for _, token := range tokens {
			if token == "" {
				continue
			}
			go func(t string) {
				_ = s.fcmClient.Send(context.Background(), t, title, body, strData)
			}(token)
		}
	}
	return nil
}

func (s *NotificationService) UpdateFCMToken(ctx context.Context, sessionID uuid.UUID, token string) error {
	return s.repo.UpdateSessionFCMToken(ctx, sessionID, token)
}

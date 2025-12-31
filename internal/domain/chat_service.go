package domain

import (
	"context"

	"github.com/google/uuid"
)

type ChatService struct {
	repo         ChatRepository
	notifService *NotificationService
}

func NewChatService(repo ChatRepository, notifService *NotificationService) *ChatService {
	return &ChatService{
		repo:         repo,
		notifService: notifService,
	}
}

func (s *ChatService) CreateChat(ctx context.Context, user1ID, user2ID uuid.UUID) (*Chat, error) {
	if user1ID == user2ID {
		// return nil, errors.New("cannot chat with self")
		// Or handle appropriately. For now let repository handle or fail.
	}
	return s.repo.CreateChat(ctx, user1ID, user2ID)
}

func (s *ChatService) GetUserChats(ctx context.Context, userID uuid.UUID) ([]*Chat, error) {
	return s.repo.GetChatsByUserID(ctx, userID)
}

func (s *ChatService) GetChat(ctx context.Context, chatID uuid.UUID) (*Chat, error) {
	return s.repo.GetChatByID(ctx, chatID)
}

func (s *ChatService) SendMessage(ctx context.Context, chatID, senderID uuid.UUID, content string) (*Message, error) {
	msg, err := s.repo.CreateMessage(ctx, chatID, senderID, content)
	if err != nil {
		return nil, err
	}

	// Send notification asynchronously
	go func() {
		// We need to find the OTHER user in the chat to notify them
		// Get participants
		chat, err := s.repo.GetChatByID(context.Background(), chatID)
		if err != nil {
			return
		}

		var receiverID uuid.UUID
		var senderName string

		for _, u := range chat.Users {
			if u.ID != senderID {
				receiverID = u.ID
			} else {
				senderName = u.Name
			}
		}

		if receiverID != uuid.Nil {
			_ = s.notifService.SendNotification(
				context.Background(),
				receiverID,
				"message",
				senderName,
				content, // In prod, truncate this
				map[string]interface{}{
					"chat_id": chatID.String(),
				},
			)
		}
	}()

	return msg, nil
}

func (s *ChatService) GetMessages(ctx context.Context, chatID uuid.UUID, limit, offset int) ([]*Message, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetMessages(ctx, chatID, limit, offset)
}

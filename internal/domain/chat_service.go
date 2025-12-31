package domain

import (
	"context"

	"github.com/google/uuid"
)

type ChatService struct {
	repo ChatRepository
}

func NewChatService(repo ChatRepository) *ChatService {
	return &ChatService{
		repo: repo,
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
	return s.repo.CreateMessage(ctx, chatID, senderID, content)
}

func (s *ChatService) GetMessages(ctx context.Context, chatID uuid.UUID, limit, offset int) ([]*Message, error) {
	if limit <= 0 {
		limit = 50
	}
	return s.repo.GetMessages(ctx, chatID, limit, offset)
}

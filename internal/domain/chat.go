package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type Chat struct {
	ID          uuid.UUID       `json:"id"`
	Users       []*UserResponse `json:"users,omitempty"`
	LastMessage *Message        `json:"last_message,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type Message struct {
	ID        uuid.UUID  `json:"id"`
	ChatID    uuid.UUID  `json:"chat_id"`
	SenderID  uuid.UUID  `json:"sender_id"`
	Content   string     `json:"content"`
	ReadAt    *time.Time `json:"read_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

type ChatRepository interface {
	CreateChat(ctx context.Context, user1ID, user2ID uuid.UUID) (*Chat, error)
	GetChatByID(ctx context.Context, chatID uuid.UUID) (*Chat, error)
	GetChatsByUserID(ctx context.Context, userID uuid.UUID) ([]*Chat, error)
	CreateMessage(ctx context.Context, chatID, senderID uuid.UUID, content string) (*Message, error)
	GetMessages(ctx context.Context, chatID uuid.UUID, limit, offset int) ([]*Message, error)
}

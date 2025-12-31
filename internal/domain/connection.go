package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type ConnectionStatus string

const (
	ConnectionStatusPending  ConnectionStatus = "pending"
	ConnectionStatusAccepted ConnectionStatus = "accepted"
	ConnectionStatusRejected ConnectionStatus = "rejected"
)

type Connection struct {
	ID          uuid.UUID        `json:"id"`
	RequesterID uuid.UUID        `json:"requester_id"`
	ReceiverID  uuid.UUID        `json:"receiver_id"`
	Status      ConnectionStatus `json:"status"`
	CreatedAt   time.Time        `json:"created_at"`
	UpdatedAt   time.Time        `json:"updated_at"`

	// For API responses
	User *UserResponse `json:"user,omitempty"`
}

type ConnectionRepository interface {
	CreateConnectionRequest(ctx context.Context, requesterID, receiverID uuid.UUID) (*Connection, error)
	UpdateConnectionStatus(ctx context.Context, connectionID uuid.UUID, status ConnectionStatus) (*Connection, error)
	GetConnectionByID(ctx context.Context, connectionID uuid.UUID) (*Connection, error)
	GetConnections(ctx context.Context, userID uuid.UUID, status ConnectionStatus, limit, offset int) ([]*Connection, error)
	DeleteConnection(ctx context.Context, connectionID uuid.UUID) error
}

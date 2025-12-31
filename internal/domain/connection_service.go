package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type ConnectionService struct {
	repo ConnectionRepository
}

func NewConnectionService(repo ConnectionRepository) *ConnectionService {
	return &ConnectionService{
		repo: repo,
	}
}

func (s *ConnectionService) SendRequest(ctx context.Context, requesterID, receiverID uuid.UUID) (*Connection, error) {
	if requesterID == receiverID {
		return nil, errors.New("cannot connect with self")
	}
	return s.repo.CreateConnectionRequest(ctx, requesterID, receiverID)
}

func (s *ConnectionService) RespondToRequest(ctx context.Context, userID, connectionID uuid.UUID, accept bool) (*Connection, error) {
	conn, err := s.repo.GetConnectionByID(ctx, connectionID)
	if err != nil {
		return nil, err
	}

	if conn.ReceiverID != userID {
		return nil, errors.New("unauthorized to respond to this request")
	}

	if conn.Status != ConnectionStatusPending {
		return nil, errors.New("connection is not pending")
	}

	status := ConnectionStatusRejected
	if accept {
		status = ConnectionStatusAccepted
	}

	return s.repo.UpdateConnectionStatus(ctx, connectionID, status)
}

func (s *ConnectionService) GetConnections(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Connection, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.GetConnections(ctx, userID, ConnectionStatusAccepted, limit, offset)
}

func (s *ConnectionService) GetPendingRequests(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*Connection, error) {
	if limit <= 0 {
		limit = 20
	}
	return s.repo.GetConnections(ctx, userID, ConnectionStatusPending, limit, offset)
}

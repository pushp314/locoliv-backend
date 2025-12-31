package domain

import (
	"context"
	"errors"

	"github.com/google/uuid"
)

type ConnectionService struct {
	repo         ConnectionRepository
	notifService *NotificationService
}

func NewConnectionService(repo ConnectionRepository, notifService *NotificationService) *ConnectionService {
	return &ConnectionService{
		repo:         repo,
		notifService: notifService,
	}
}

func (s *ConnectionService) SendRequest(ctx context.Context, requesterID, receiverID uuid.UUID) (*Connection, error) {
	if requesterID == receiverID {
		return nil, errors.New("cannot connect with self")
	}
	conn, err := s.repo.CreateConnectionRequest(ctx, requesterID, receiverID)
	if err != nil {
		return nil, err
	}

	// Notify receiver
	go func() {
		// Need requester name. Ideally service should look it up or accept it.
		// For now simple generic message or fetch user
		// Not injecting UserRepo here to avoid bloat, assuming just "New Request" is enough for now or id lookup inside
		// Actually, let's just say "New Connection Request"
		_ = s.notifService.SendNotification(
			context.Background(),
			receiverID,
			"connection_request",
			"New Connection Request",
			"Someone wants to connect with you",
			map[string]interface{}{
				"requester_id": requesterID.String(),
			},
		)
	}()

	return conn, nil
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

	updatedConn, err := s.repo.UpdateConnectionStatus(ctx, connectionID, status)
	if err != nil {
		return nil, err
	}

	if accept {
		// Notify original requester
		go func() {
			_ = s.notifService.SendNotification(
				context.Background(),
				conn.RequesterID,
				"connection_accepted",
				"Connection Accepted",
				"You are now connected!",
				map[string]interface{}{
					"accepter_id": userID.String(),
				},
			)
		}()
	}

	return updatedConn, nil
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

package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/locolive/backend/internal/domain"
)

// VerificationRepository implements domain.VerificationRepository
type VerificationRepository struct {
	db *pgxpool.Pool
}

func NewVerificationRepository(db *pgxpool.Pool) *VerificationRepository {
	return &VerificationRepository{db: db}
}

func (r *VerificationRepository) CreateRequest(ctx context.Context, req *domain.VerificationRequest) error {
	query := `
		INSERT INTO verification_requests (id, user_id, request_type, document_url, document_type, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
	`
	_, err := r.db.Exec(ctx, query,
		req.ID, req.UserID, req.RequestType, req.DocumentURL, req.DocumentType,
		req.Status, req.CreatedAt, req.UpdatedAt,
	)
	return err
}

func (r *VerificationRepository) GetRequest(ctx context.Context, reqID uuid.UUID) (*domain.VerificationRequest, error) {
	query := `
		SELECT id, user_id, request_type, document_url, document_type, status, 
		       rejection_reason, reviewed_by, reviewed_at, created_at, updated_at
		FROM verification_requests WHERE id = $1
	`
	req := &domain.VerificationRequest{}
	err := r.db.QueryRow(ctx, query, reqID).Scan(
		&req.ID, &req.UserID, &req.RequestType, &req.DocumentURL, &req.DocumentType,
		&req.Status, &req.RejectionReason, &req.ReviewedBy, &req.ReviewedAt,
		&req.CreatedAt, &req.UpdatedAt,
	)
	return req, err
}

func (r *VerificationRepository) GetUserRequests(ctx context.Context, userID uuid.UUID) ([]domain.VerificationRequest, error) {
	query := `
		SELECT id, user_id, request_type, document_url, document_type, status, 
		       rejection_reason, reviewed_by, reviewed_at, created_at, updated_at
		FROM verification_requests WHERE user_id = $1 ORDER BY created_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []domain.VerificationRequest
	for rows.Next() {
		var req domain.VerificationRequest
		if err := rows.Scan(
			&req.ID, &req.UserID, &req.RequestType, &req.DocumentURL, &req.DocumentType,
			&req.Status, &req.RejectionReason, &req.ReviewedBy, &req.ReviewedAt,
			&req.CreatedAt, &req.UpdatedAt,
		); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

func (r *VerificationRepository) GetPendingRequests(ctx context.Context, limit, offset int) ([]domain.VerificationRequest, error) {
	query := `
		SELECT id, user_id, request_type, document_url, document_type, status, 
		       rejection_reason, reviewed_by, reviewed_at, created_at, updated_at
		FROM verification_requests 
		WHERE status = 'pending' 
		ORDER BY created_at ASC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var requests []domain.VerificationRequest
	for rows.Next() {
		var req domain.VerificationRequest
		if err := rows.Scan(
			&req.ID, &req.UserID, &req.RequestType, &req.DocumentURL, &req.DocumentType,
			&req.Status, &req.RejectionReason, &req.ReviewedBy, &req.ReviewedAt,
			&req.CreatedAt, &req.UpdatedAt,
		); err != nil {
			return nil, err
		}
		requests = append(requests, req)
	}
	return requests, nil
}

func (r *VerificationRepository) UpdateRequest(ctx context.Context, req *domain.VerificationRequest) error {
	query := `
		UPDATE verification_requests 
		SET status = $1, rejection_reason = $2, reviewed_by = $3, reviewed_at = $4, updated_at = $5
		WHERE id = $6
	`
	_, err := r.db.Exec(ctx, query,
		req.Status, req.RejectionReason, req.ReviewedBy, req.ReviewedAt, req.UpdatedAt, req.ID,
	)
	return err
}

func (r *VerificationRepository) VerifyUser(ctx context.Context, userID uuid.UUID, vType domain.VerificationType, verifiedBy uuid.UUID) error {
	query := `
		UPDATE users 
		SET is_verified = TRUE, verification_type = $1, verified_at = $2, verified_by = $3
		WHERE id = $4
	`
	_, err := r.db.Exec(ctx, query, vType, time.Now(), verifiedBy, userID)
	return err
}

func (r *VerificationRepository) UnverifyUser(ctx context.Context, userID uuid.UUID) error {
	query := `
		UPDATE users 
		SET is_verified = FALSE, verification_type = NULL, verified_at = NULL, verified_by = NULL
		WHERE id = $1
	`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

func (r *VerificationRepository) GetUserVerification(ctx context.Context, userID uuid.UUID) (*domain.UserVerification, error) {
	query := `
		SELECT is_verified, verification_type, verified_at
		FROM users WHERE id = $1
	`
	v := &domain.UserVerification{}
	var vType *string
	err := r.db.QueryRow(ctx, query, userID).Scan(&v.IsVerified, &vType, &v.VerifiedAt)
	if err != nil {
		return nil, err
	}
	if vType != nil {
		t := domain.VerificationType(*vType)
		v.VerificationType = &t
	}
	return v, nil
}

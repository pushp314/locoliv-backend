package domain

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// VerificationType represents types of verification badges
type VerificationType string

const (
	VerificationBlue      VerificationType = "blue"      // Standard verified user
	VerificationCelebrity VerificationType = "celebrity" // Celebrity/Influencer
	VerificationOfficial  VerificationType = "official"  // Official government/organization
	VerificationBusiness  VerificationType = "business"  // Verified business
)

// VerificationRequestStatus represents request status
type VerificationRequestStatus string

const (
	VerificationPending  VerificationRequestStatus = "pending"
	VerificationApproved VerificationRequestStatus = "approved"
	VerificationRejected VerificationRequestStatus = "rejected"
)

// VerificationRequest represents a user's request for verification
type VerificationRequest struct {
	ID              uuid.UUID                 `json:"id"`
	UserID          uuid.UUID                 `json:"user_id"`
	RequestType     VerificationType          `json:"request_type"`
	DocumentURL     *string                   `json:"document_url,omitempty"`
	DocumentType    *string                   `json:"document_type,omitempty"` // aadhaar, pan, passport, etc.
	Status          VerificationRequestStatus `json:"status"`
	RejectionReason *string                   `json:"rejection_reason,omitempty"`
	ReviewedBy      *uuid.UUID                `json:"reviewed_by,omitempty"`
	ReviewedAt      *time.Time                `json:"reviewed_at,omitempty"`
	CreatedAt       time.Time                 `json:"created_at"`
	UpdatedAt       time.Time                 `json:"updated_at"`
}

// UserVerification represents a user's verification status
type UserVerification struct {
	IsVerified       bool              `json:"is_verified"`
	VerificationType *VerificationType `json:"verification_type,omitempty"`
	VerifiedAt       *time.Time        `json:"verified_at,omitempty"`
}

// VerificationRepository interface
type VerificationRepository interface {
	CreateRequest(ctx context.Context, req *VerificationRequest) error
	GetRequest(ctx context.Context, reqID uuid.UUID) (*VerificationRequest, error)
	GetUserRequests(ctx context.Context, userID uuid.UUID) ([]VerificationRequest, error)
	GetPendingRequests(ctx context.Context, limit, offset int) ([]VerificationRequest, error)
	UpdateRequest(ctx context.Context, req *VerificationRequest) error

	VerifyUser(ctx context.Context, userID uuid.UUID, vType VerificationType, verifiedBy uuid.UUID) error
	UnverifyUser(ctx context.Context, userID uuid.UUID) error
	GetUserVerification(ctx context.Context, userID uuid.UUID) (*UserVerification, error)
}

// VerificationService handles verification logic
type VerificationService struct {
	repo VerificationRepository
}

// NewVerificationService creates a new verification service
func NewVerificationService(repo VerificationRepository) *VerificationService {
	return &VerificationService{repo: repo}
}

// SubmitRequest submits a new verification request
func (s *VerificationService) SubmitRequest(ctx context.Context, userID uuid.UUID, reqType VerificationType, documentURL, documentType *string) (*VerificationRequest, error) {
	req := &VerificationRequest{
		ID:           uuid.New(),
		UserID:       userID,
		RequestType:  reqType,
		DocumentURL:  documentURL,
		DocumentType: documentType,
		Status:       VerificationPending,
		CreatedAt:    time.Now(),
		UpdatedAt:    time.Now(),
	}

	if err := s.repo.CreateRequest(ctx, req); err != nil {
		return nil, err
	}

	return req, nil
}

// GetMyRequests gets all verification requests for a user
func (s *VerificationService) GetMyRequests(ctx context.Context, userID uuid.UUID) ([]VerificationRequest, error) {
	return s.repo.GetUserRequests(ctx, userID)
}

// GetPendingRequests gets all pending requests (admin)
func (s *VerificationService) GetPendingRequests(ctx context.Context, limit, offset int) ([]VerificationRequest, error) {
	return s.repo.GetPendingRequests(ctx, limit, offset)
}

// ApproveRequest approves a verification request (admin)
func (s *VerificationService) ApproveRequest(ctx context.Context, reqID, adminID uuid.UUID) error {
	req, err := s.repo.GetRequest(ctx, reqID)
	if err != nil {
		return err
	}

	now := time.Now()
	req.Status = VerificationApproved
	req.ReviewedBy = &adminID
	req.ReviewedAt = &now
	req.UpdatedAt = now

	if err := s.repo.UpdateRequest(ctx, req); err != nil {
		return err
	}

	// Verify the user
	return s.repo.VerifyUser(ctx, req.UserID, req.RequestType, adminID)
}

// RejectRequest rejects a verification request (admin)
func (s *VerificationService) RejectRequest(ctx context.Context, reqID, adminID uuid.UUID, reason string) error {
	req, err := s.repo.GetRequest(ctx, reqID)
	if err != nil {
		return err
	}

	now := time.Now()
	req.Status = VerificationRejected
	req.RejectionReason = &reason
	req.ReviewedBy = &adminID
	req.ReviewedAt = &now
	req.UpdatedAt = now

	return s.repo.UpdateRequest(ctx, req)
}

// GetUserVerification gets a user's verification status
func (s *VerificationService) GetUserVerification(ctx context.Context, userID uuid.UUID) (*UserVerification, error) {
	return s.repo.GetUserVerification(ctx, userID)
}

// RevokeVerification removes verification from a user (admin)
func (s *VerificationService) RevokeVerification(ctx context.Context, userID uuid.UUID) error {
	return s.repo.UnverifyUser(ctx, userID)
}

// GetVerificationBadgeEmoji returns emoji for verification type
func GetVerificationBadgeEmoji(vType VerificationType) string {
	switch vType {
	case VerificationBlue:
		return "✓"
	case VerificationCelebrity:
		return "⭐"
	case VerificationOfficial:
		return "🏛️"
	case VerificationBusiness:
		return "🏢"
	default:
		return ""
	}
}

// GetVerificationBadgeColor returns color for verification type
func GetVerificationBadgeColor(vType VerificationType) string {
	switch vType {
	case VerificationBlue:
		return "#1DA1F2" // Twitter blue
	case VerificationCelebrity:
		return "#FFD700" // Gold
	case VerificationOfficial:
		return "#4CAF50" // Green
	case VerificationBusiness:
		return "#9C27B0" // Purple
	default:
		return ""
	}
}

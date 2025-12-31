package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the domain layer
type User struct {
	ID            uuid.UUID  `json:"id"`
	Email         *string    `json:"email,omitempty"`
	Phone         *string    `json:"phone,omitempty"`
	Name          string     `json:"name"`
	AvatarURL     *string    `json:"avatar_url,omitempty"`
	Bio           *string    `json:"bio,omitempty"`
	Gender        *string    `json:"gender,omitempty"`
	DateOfBirth   *time.Time `json:"date_of_birth,omitempty"`
	Visibility    string     `json:"visibility"`
	GoogleID      *string    `json:"-"`
	EmailVerified bool       `json:"email_verified"`
	PhoneVerified bool       `json:"phone_verified"`
	IsActive      bool       `json:"is_active"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

// UserResponse is the public representation of a user
type UserResponse struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email,omitempty"`
	Phone         string    `json:"phone,omitempty"`
	Name          string    `json:"name"`
	AvatarURL     string    `json:"avatar_url,omitempty"`
	Bio           string    `json:"bio,omitempty"`
	Gender        string    `json:"gender,omitempty"`
	DateOfBirth   string    `json:"date_of_birth,omitempty"`
	Visibility    string    `json:"visibility,omitempty"`
	EmailVerified bool      `json:"email_verified"`
	PhoneVerified bool      `json:"phone_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

// ToResponse converts a User to a UserResponse
func (u *User) ToResponse() *UserResponse {
	response := &UserResponse{
		ID:            u.ID,
		Name:          u.Name,
		Visibility:    u.Visibility,
		EmailVerified: u.EmailVerified,
		PhoneVerified: u.PhoneVerified,
		CreatedAt:     u.CreatedAt,
	}

	if u.Email != nil {
		response.Email = *u.Email
	}
	if u.Phone != nil {
		response.Phone = *u.Phone
	}
	if u.AvatarURL != nil {
		response.AvatarURL = *u.AvatarURL
	}
	if u.Bio != nil {
		response.Bio = *u.Bio
	}
	if u.Gender != nil {
		response.Gender = *u.Gender
	}
	if u.DateOfBirth != nil {
		response.DateOfBirth = u.DateOfBirth.Format("2006-01-02")
	}

	return response
}

// Session represents a user session
type Session struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	DeviceInfo     *string   `json:"device_info,omitempty"`
	IPAddress      *string   `json:"ip_address,omitempty"`
	UserAgent      *string   `json:"user_agent,omitempty"`
	FCMToken       *string   `json:"fcm_token,omitempty"`
	IsActive       bool      `json:"is_active"`
	CreatedAt      time.Time `json:"created_at"`
	ExpiresAt      time.Time `json:"expires_at"`
	LastActivityAt time.Time `json:"last_activity_at"`
}

// RefreshToken represents a stored refresh token
type RefreshToken struct {
	ID        uuid.UUID  `json:"id"`
	UserID    uuid.UUID  `json:"user_id"`
	SessionID *uuid.UUID `json:"session_id,omitempty"`
	TokenHash string     `json:"-"`
	ExpiresAt time.Time  `json:"expires_at"`
	Revoked   bool       `json:"revoked"`
	RevokedAt *time.Time `json:"revoked_at,omitempty"`
	CreatedAt time.Time  `json:"created_at"`
}

// PasswordResetToken represents a password reset token
type PasswordResetToken struct {
	ID        uuid.UUID `json:"id"`
	UserID    uuid.UUID `json:"user_id"`
	TokenHash string    `json:"-"`
	ExpiresAt time.Time `json:"expires_at"`
	Used      bool      `json:"used"`
	CreatedAt time.Time `json:"created_at"`
}

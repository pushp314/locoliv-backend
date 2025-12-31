package domain

import (
	"time"

	"github.com/google/uuid"
)

// User represents a user in the domain layer
type User struct {
	ID            uuid.UUID `json:"id"`
	Email         *string   `json:"email,omitempty"`
	Phone         *string   `json:"phone,omitempty"`
	Name          string    `json:"name"`
	AvatarURL     *string   `json:"avatar_url,omitempty"`
	GoogleID      *string   `json:"-"`
	EmailVerified bool      `json:"email_verified"`
	PhoneVerified bool      `json:"phone_verified"`
	IsActive      bool      `json:"is_active"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

// UserResponse is the public representation of a user
type UserResponse struct {
	ID            uuid.UUID `json:"id"`
	Email         string    `json:"email,omitempty"`
	Phone         string    `json:"phone,omitempty"`
	Name          string    `json:"name"`
	AvatarURL     string    `json:"avatar_url,omitempty"`
	EmailVerified bool      `json:"email_verified"`
	PhoneVerified bool      `json:"phone_verified"`
	CreatedAt     time.Time `json:"created_at"`
}

// ToResponse converts a User to a UserResponse
func (u *User) ToResponse() *UserResponse {
	response := &UserResponse{
		ID:            u.ID,
		Name:          u.Name,
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

	return response
}

// Session represents a user session
type Session struct {
	ID             uuid.UUID `json:"id"`
	UserID         uuid.UUID `json:"user_id"`
	DeviceInfo     *string   `json:"device_info,omitempty"`
	IPAddress      *string   `json:"ip_address,omitempty"`
	UserAgent      *string   `json:"user_agent,omitempty"`
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

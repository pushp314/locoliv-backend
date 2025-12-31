package domain

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/locolive/backend/internal/auth"
)

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrTokenRevoked       = errors.New("token has been revoked")
	ErrSessionExpired     = errors.New("session has expired")
	ErrInvalidToken       = errors.New("invalid token")
	ErrTokenExpired       = errors.New("token has expired")
)

// AuthRepository defines the interface for auth data access
type AuthRepository interface {
	// User operations
	CreateUser(ctx context.Context, params CreateUserParams) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByPhone(ctx context.Context, phone string) (*User, error)
	GetUserByGoogleID(ctx context.Context, googleID string) (*User, error)
	UpdateUser(ctx context.Context, userID uuid.UUID, params UpdateUserParams) (*User, error)
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	UpdateUserEmail(ctx context.Context, userID uuid.UUID, email string) error
	LinkGoogleAccount(ctx context.Context, userID uuid.UUID, googleID string) (*User, error)
	UserExistsByEmail(ctx context.Context, email string) (bool, error)
	UserExistsByPhone(ctx context.Context, phone string) (bool, error)
	VerifyUserPassword(ctx context.Context, email, password string) (*User, error)

	// Session operations
	CreateSession(ctx context.Context, params CreateSessionParams) (*Session, error)
	GetSessionByID(ctx context.Context, id uuid.UUID) (*Session, error)
	DeactivateSession(ctx context.Context, id uuid.UUID) error
	DeactivateUserSessions(ctx context.Context, userID uuid.UUID) error

	// Refresh token operations
	CreateRefreshToken(ctx context.Context, params CreateRefreshTokenParams) (*RefreshToken, error)
	GetRefreshTokenByHash(ctx context.Context, hash string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, id uuid.UUID) error
	RevokeRefreshTokenByHash(ctx context.Context, hash string) error
	RevokeUserRefreshTokens(ctx context.Context, userID uuid.UUID) error

	// Password reset token operations
	CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	GetPasswordResetToken(ctx context.Context, tokenHash string) (*PasswordResetToken, error)
	MarkPasswordResetTokenUsed(ctx context.Context, id uuid.UUID) error
}

// CreateUserParams holds parameters for user creation
type CreateUserParams struct {
	Email         *string
	Phone         *string
	PasswordHash  *string
	Name          string
	GoogleID      *string
	EmailVerified bool
}

// UpdateUserParams holds parameters for user update
type UpdateUserParams struct {
	Name        *string    `json:"name"`
	Bio         *string    `json:"bio"`
	Gender      *string    `json:"gender"`
	DateOfBirth *time.Time `json:"date_of_birth"`
	Visibility  *string    `json:"visibility"`
	AvatarURL   *string    `json:"avatar_url"`
}

// CreateSessionParams holds parameters for session creation
type CreateSessionParams struct {
	UserID     uuid.UUID
	DeviceInfo *string
	IPAddress  *string
	UserAgent  *string
	ExpiresAt  time.Time
}

// CreateRefreshTokenParams holds parameters for refresh token creation
type CreateRefreshTokenParams struct {
	UserID    uuid.UUID
	SessionID *uuid.UUID
	TokenHash string
	ExpiresAt time.Time
}

// AuthService handles authentication business logic
type AuthService struct {
	repo   AuthRepository
	jwt    *auth.JWTManager
	google *auth.GoogleAuthVerifier
}

// NewAuthService creates a new auth service
func NewAuthService(repo AuthRepository, jwt *auth.JWTManager, google *auth.GoogleAuthVerifier) *AuthService {
	return &AuthService{
		repo:   repo,
		jwt:    jwt,
		google: google,
	}
}

// RegisterResult represents the result of registration
type RegisterResult struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
}

// Register creates a new user with email/password
func (s *AuthService) Register(ctx context.Context, email, password, name string) (*RegisterResult, error) {
	// Check if user exists
	exists, err := s.repo.UserExistsByEmail(ctx, email)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, ErrUserAlreadyExists
	}

	// Hash password
	passwordHash, err := auth.HashPassword(password)
	if err != nil {
		return nil, err
	}

	// Create user
	user, err := s.repo.CreateUser(ctx, CreateUserParams{
		Email:        &email,
		PasswordHash: &passwordHash,
		Name:         name,
	})
	if err != nil {
		return nil, err
	}

	// Create session
	session, err := s.repo.CreateSession(ctx, CreateSessionParams{
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour), // 30 days
		// Device info could be passed in context or params, but for now defaults
	})
	if err != nil {
		return nil, err
	}

	// Generate tokens
	tokenPair, err := s.jwt.GenerateTokenPair(user.ID, session.ID, email)
	if err != nil {
		return nil, err
	}

	// Store refresh token
	tokenHash := auth.HashToken(tokenPair.RefreshToken)
	_, err = s.repo.CreateRefreshToken(ctx, CreateRefreshTokenParams{
		UserID:    user.ID,
		SessionID: &session.ID,
		TokenHash: tokenHash,
		ExpiresAt: tokenPair.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}

	return &RegisterResult{
		User:         user.ToResponse(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// LoginResult represents the result of login
type LoginResult struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
}

// Login authenticates a user with email/password
func (s *AuthService) Login(ctx context.Context, email, password string) (*LoginResult, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// User must have a password (not OAuth-only)
	if user.Email == nil {
		return nil, ErrInvalidCredentials
	}

	// Verify password
	_, err = s.repo.VerifyUserPassword(ctx, *user.Email, password)
	if err != nil {
		return nil, ErrInvalidCredentials
	}

	// Create session
	session, err := s.repo.CreateSession(ctx, CreateSessionParams{
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	})
	if err != nil {
		return nil, err
	}

	// Generate tokens
	tokenPair, err := s.jwt.GenerateTokenPair(user.ID, session.ID, *user.Email)
	if err != nil {
		return nil, err
	}

	// Store refresh token
	tokenHash := auth.HashToken(tokenPair.RefreshToken)
	_, err = s.repo.CreateRefreshToken(ctx, CreateRefreshTokenParams{
		UserID:    user.ID,
		SessionID: &session.ID,
		TokenHash: tokenHash,
		ExpiresAt: tokenPair.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}

	return &LoginResult{
		User:         user.ToResponse(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// RefreshResult represents the result of token refresh
type RefreshResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RefreshToken validates and rotates a refresh token
func (s *AuthService) RefreshToken(ctx context.Context, refreshToken string) (*RefreshResult, error) {
	// Validate the JWT refresh token
	claims, err := s.jwt.ValidateRefreshToken(refreshToken)
	if err != nil {
		return nil, err
	}

	// Get the stored token
	tokenHash := auth.HashToken(refreshToken)
	storedToken, err := s.repo.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		return nil, ErrTokenRevoked
	}

	if storedToken.Revoked {
		// Token reuse detected - revoke all user tokens
		_ = s.repo.RevokeUserRefreshTokens(ctx, claims.UserID)
		return nil, ErrTokenRevoked
	}

	// Revoke the old token
	_ = s.repo.RevokeRefreshToken(ctx, storedToken.ID)

	// Get user for email
	user, err := s.repo.GetUserByID(ctx, claims.UserID)
	if err != nil {
		return nil, ErrUserNotFound
	}

	email := ""
	if user.Email != nil {
		email = *user.Email
	}

	// Handle session
	var sessionID uuid.UUID
	if storedToken.SessionID != nil {
		sessionID = *storedToken.SessionID
	} else {
		// Legacy token without session, create one
		session, err := s.repo.CreateSession(ctx, CreateSessionParams{
			UserID:    claims.UserID,
			ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
		})
		if err != nil {
			return nil, err
		}
		sessionID = session.ID
	}

	// Generate new token pair
	tokenPair, err := s.jwt.GenerateTokenPair(claims.UserID, sessionID, email)
	if err != nil {
		return nil, err
	}

	// Store new refresh token
	newTokenHash := auth.HashToken(tokenPair.RefreshToken)
	_, err = s.repo.CreateRefreshToken(ctx, CreateRefreshTokenParams{
		UserID:    claims.UserID,
		SessionID: &sessionID,
		TokenHash: newTokenHash,
		ExpiresAt: tokenPair.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}

	return &RefreshResult{
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
	}, nil
}

// Logout revokes a refresh token
func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	tokenHash := auth.HashToken(refreshToken)
	return s.repo.RevokeRefreshTokenByHash(ctx, tokenHash)
}

// LogoutAll revokes all refresh tokens for a user
func (s *AuthService) LogoutAll(ctx context.Context, userID uuid.UUID) error {
	return s.repo.RevokeUserRefreshTokens(ctx, userID)
}

// GoogleLoginResult represents the result of Google OAuth login
type GoogleLoginResult struct {
	User         *UserResponse `json:"user"`
	AccessToken  string        `json:"access_token"`
	RefreshToken string        `json:"refresh_token"`
	IsNewUser    bool          `json:"is_new_user"`
}

// GoogleLogin handles Google OAuth login
func (s *AuthService) GoogleLogin(ctx context.Context, idToken string) (*GoogleLoginResult, error) {
	// Verify Google ID token
	googleUser, err := s.google.VerifyIDToken(ctx, idToken)
	if err != nil {
		return nil, err
	}

	var user *User
	isNewUser := false

	// Try to find existing user by Google ID
	user, err = s.repo.GetUserByGoogleID(ctx, googleUser.GoogleID)
	if err != nil {
		// Try to find by email
		user, err = s.repo.GetUserByEmail(ctx, googleUser.Email)
		if err != nil {
			// Create new user
			googleID := googleUser.GoogleID
			avatarURL := googleUser.Picture

			user, err = s.repo.CreateUser(ctx, CreateUserParams{
				Email:         &googleUser.Email,
				Name:          googleUser.Name,
				GoogleID:      &googleID,
				EmailVerified: googleUser.EmailVerified,
			})
			if err != nil {
				return nil, err
			}

			// Set avatar if provided
			if avatarURL != "" {
				user.AvatarURL = &avatarURL
			}

			isNewUser = true
		} else {
			// Link Google account to existing user
			user, err = s.repo.LinkGoogleAccount(ctx, user.ID, googleUser.GoogleID)
			if err != nil {
				return nil, err
			}
		}
	}

	// Create session
	session, err := s.repo.CreateSession(ctx, CreateSessionParams{
		UserID:    user.ID,
		ExpiresAt: time.Now().Add(30 * 24 * time.Hour),
	})
	if err != nil {
		return nil, err
	}

	// Generate tokens
	tokenPair, err := s.jwt.GenerateTokenPair(user.ID, session.ID, googleUser.Email)
	if err != nil {
		return nil, err
	}

	// Store refresh token
	tokenHash := auth.HashToken(tokenPair.RefreshToken)
	_, err = s.repo.CreateRefreshToken(ctx, CreateRefreshTokenParams{
		UserID:    user.ID,
		SessionID: &session.ID,
		TokenHash: tokenHash,
		ExpiresAt: tokenPair.ExpiresAt,
	})
	if err != nil {
		return nil, err
	}

	return &GoogleLoginResult{
		User:         user.ToResponse(),
		AccessToken:  tokenPair.AccessToken,
		RefreshToken: tokenPair.RefreshToken,
		IsNewUser:    isNewUser,
	}, nil
}

// GetUserByID retrieves a user by ID
func (s *AuthService) GetUserByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.GetUserByID(ctx, id)
}

// InitiatePasswordReset creates a password reset token
func (s *AuthService) InitiatePasswordReset(ctx context.Context, email string) (string, error) {
	user, err := s.repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", ErrUserNotFound
	}

	// Generate reset token
	token := auth.GenerateRandomToken(32)
	tokenHash := auth.HashToken(token)
	expiresAt := time.Now().Add(1 * time.Hour)

	err = s.repo.CreatePasswordResetToken(ctx, user.ID, tokenHash, expiresAt)
	if err != nil {
		return "", err
	}

	return token, nil
}

// ResetPassword resets password using a reset token
func (s *AuthService) ResetPassword(ctx context.Context, token, newPassword string) error {
	tokenHash := auth.HashToken(token)

	// Get and validate token
	resetToken, err := s.repo.GetPasswordResetToken(ctx, tokenHash)
	if err != nil {
		return ErrInvalidToken
	}

	if time.Now().After(resetToken.ExpiresAt) {
		return ErrTokenExpired
	}

	if resetToken.Used {
		return ErrInvalidToken
	}

	// Hash new password
	passwordHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	err = s.repo.UpdateUserPassword(ctx, resetToken.UserID, passwordHash)
	if err != nil {
		return err
	}

	// Mark token as used
	_ = s.repo.MarkPasswordResetTokenUsed(ctx, resetToken.ID)

	// Revoke all refresh tokens for security
	_ = s.repo.RevokeUserRefreshTokens(ctx, resetToken.UserID)

	return nil
}

// UpdatePassword changes password for authenticated user
func (s *AuthService) UpdatePassword(ctx context.Context, userID uuid.UUID, currentPassword, newPassword string) error {
	// Get user with password
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.Email == nil {
		return ErrInvalidCredentials
	}

	// Verify current password using repository method
	_, err = s.repo.VerifyUserPassword(ctx, *user.Email, currentPassword)
	if err != nil {
		return ErrInvalidCredentials
	}

	// Hash new password
	passwordHash, err := auth.HashPassword(newPassword)
	if err != nil {
		return err
	}

	// Update password
	return s.repo.UpdateUserPassword(ctx, userID, passwordHash)
}

// UpdateEmail changes email for authenticated user
func (s *AuthService) UpdateEmail(ctx context.Context, userID uuid.UUID, newEmail, password string) error {
	// Get user
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return ErrUserNotFound
	}

	if user.Email == nil {
		return ErrInvalidCredentials
	}

	// Verify password
	_, err = s.repo.VerifyUserPassword(ctx, *user.Email, password)
	if err != nil {
		return ErrInvalidCredentials
	}

	// Check if new email exists
	exists, err := s.repo.UserExistsByEmail(ctx, newEmail)
	if err != nil {
		return err
	}
	if exists {
		return ErrUserAlreadyExists
	}

	// Update email
	return s.repo.UpdateUserEmail(ctx, userID, newEmail)
}

// UpdateProfile updates the authenticated user's profile
func (s *AuthService) UpdateProfile(ctx context.Context, userID uuid.UUID, params UpdateUserParams) (*UserResponse, error) {
	// Update user in repo
	user, err := s.repo.UpdateUser(ctx, userID, params)
	if err != nil {
		return nil, err
	}

	return user.ToResponse(), nil
}

// GetUser retrieves a user by ID
func (s *AuthService) GetUser(ctx context.Context, userID uuid.UUID) (*UserResponse, error) {
	user, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	return user.ToResponse(), nil
}

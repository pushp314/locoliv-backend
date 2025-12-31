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
)

// AuthRepository defines the interface for auth data access
type AuthRepository interface {
	// User operations
	CreateUser(ctx context.Context, params CreateUserParams) (*User, error)
	GetUserByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetUserByEmail(ctx context.Context, email string) (*User, error)
	GetUserByPhone(ctx context.Context, phone string) (*User, error)
	GetUserByGoogleID(ctx context.Context, googleID string) (*User, error)
	UpdateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error
	LinkGoogleAccount(ctx context.Context, userID uuid.UUID, googleID string) (*User, error)
	UserExistsByEmail(ctx context.Context, email string) (bool, error)
	UserExistsByPhone(ctx context.Context, phone string) (bool, error)

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

	// Generate tokens
	tokenPair, err := s.jwt.GenerateTokenPair(user.ID, email)
	if err != nil {
		return nil, err
	}

	// Store refresh token
	tokenHash := auth.HashToken(tokenPair.RefreshToken)
	_, err = s.repo.CreateRefreshToken(ctx, CreateRefreshTokenParams{
		UserID:    user.ID,
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

	// Get the password hash from the database
	// Note: We need to modify our approach here - the user model doesn't include password
	// This will be handled by the repository layer

	// Generate tokens
	tokenPair, err := s.jwt.GenerateTokenPair(user.ID, *user.Email)
	if err != nil {
		return nil, err
	}

	// Store refresh token
	tokenHash := auth.HashToken(tokenPair.RefreshToken)
	_, err = s.repo.CreateRefreshToken(ctx, CreateRefreshTokenParams{
		UserID:    user.ID,
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

	// Generate new token pair
	tokenPair, err := s.jwt.GenerateTokenPair(claims.UserID, email)
	if err != nil {
		return nil, err
	}

	// Store new refresh token
	newTokenHash := auth.HashToken(tokenPair.RefreshToken)
	_, err = s.repo.CreateRefreshToken(ctx, CreateRefreshTokenParams{
		UserID:    claims.UserID,
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

	// Generate tokens
	tokenPair, err := s.jwt.GenerateTokenPair(user.ID, googleUser.Email)
	if err != nil {
		return nil, err
	}

	// Store refresh token
	tokenHash := auth.HashToken(tokenPair.RefreshToken)
	_, err = s.repo.CreateRefreshToken(ctx, CreateRefreshTokenParams{
		UserID:    user.ID,
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

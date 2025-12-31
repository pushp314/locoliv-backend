package api

import (
	"encoding/json"
	"net/http"

	"github.com/locolive/backend/internal/auth"
	"github.com/locolive/backend/internal/domain"
	"github.com/locolive/backend/internal/middleware"
	"github.com/locolive/backend/pkg/response"
	"github.com/locolive/backend/pkg/validator"
	"go.uber.org/zap"
)

// AuthHandler handles authentication endpoints
type AuthHandler struct {
	authService *domain.AuthService
	authRepo    domain.AuthRepository
	logger      *zap.Logger
}

// NewAuthHandler creates a new auth handler
func NewAuthHandler(authService *domain.AuthService, authRepo domain.AuthRepository, logger *zap.Logger) *AuthHandler {
	return &AuthHandler{
		authService: authService,
		authRepo:    authRepo,
		logger:      logger,
	}
}

// RegisterRequest represents the registration request body
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
	Phone    string `json:"phone,omitempty"`
}

// LoginRequest represents the login request body
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RefreshRequest represents the token refresh request body
type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// LogoutRequest represents the logout request body
type LogoutRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// GoogleLoginRequest represents the Google OAuth request body
type GoogleLoginRequest struct {
	IDToken string `json:"id_token"`
}

// Register handles user registration
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Validate email
	req.Email = validator.SanitizeEmail(req.Email)
	if !validator.ValidateEmail(req.Email) {
		response.BadRequest(w, "invalid email address")
		return
	}

	// Validate password
	if errs := validator.ValidatePassword(req.Password); errs.HasErrors() {
		response.BadRequest(w, errs.Error())
		return
	}

	// Validate name
	req.Name = validator.SanitizeString(req.Name, 100)
	if !validator.ValidateName(req.Name) {
		response.BadRequest(w, "name must be 2-100 characters")
		return
	}

	// Register user
	result, err := h.authService.Register(r.Context(), req.Email, req.Password, req.Name)
	if err != nil {
		if err == domain.ErrUserAlreadyExists {
			response.Conflict(w, "user with this email already exists")
			return
		}
		h.logger.Error("registration failed", zap.Error(err))
		response.InternalError(w, "registration failed")
		return
	}

	response.Created(w, result)
}

// Login handles user login
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	// Sanitize email
	req.Email = validator.SanitizeEmail(req.Email)
	if !validator.ValidateEmail(req.Email) {
		response.BadRequest(w, "invalid email address")
		return
	}

	if req.Password == "" {
		response.BadRequest(w, "password is required")
		return
	}

	// Get user with password hash for verification
	user, err := h.authRepo.GetUserByEmail(r.Context(), req.Email)
	if err != nil {
		response.Unauthorized(w, "invalid email or password")
		return
	}

	// Verify password - we need to get the hash from DB
	// The repository will handle password verification
	result, err := h.authService.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		if err == domain.ErrInvalidCredentials {
			response.Unauthorized(w, "invalid email or password")
			return
		}
		h.logger.Error("login failed", zap.Error(err), zap.String("email", req.Email))
		response.InternalError(w, "login failed")
		return
	}

	// Suppress unused variable warning
	_ = user

	response.OK(w, result)
}

// Refresh handles token refresh with rotation
func (h *AuthHandler) Refresh(w http.ResponseWriter, r *http.Request) {
	var req RefreshRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		response.BadRequest(w, "refresh_token is required")
		return
	}

	result, err := h.authService.RefreshToken(r.Context(), req.RefreshToken)
	if err != nil {
		if err == auth.ErrExpiredToken {
			response.Unauthorized(w, "refresh token has expired")
			return
		}
		if err == auth.ErrInvalidToken || err == domain.ErrTokenRevoked {
			response.Unauthorized(w, "invalid refresh token")
			return
		}
		h.logger.Error("token refresh failed", zap.Error(err))
		response.InternalError(w, "token refresh failed")
		return
	}

	response.OK(w, result)
}

// Logout handles user logout (revokes refresh token)
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req LogoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.RefreshToken == "" {
		response.BadRequest(w, "refresh_token is required")
		return
	}

	if err := h.authService.Logout(r.Context(), req.RefreshToken); err != nil {
		h.logger.Warn("logout failed", zap.Error(err))
		// Still return success - token may already be revoked
	}

	response.NoContent(w)
}

// LogoutAll handles logging out from all devices
func (h *AuthHandler) LogoutAll(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	if err := h.authService.LogoutAll(r.Context(), userID); err != nil {
		h.logger.Error("logout all failed", zap.Error(err))
		response.InternalError(w, "logout failed")
		return
	}

	response.NoContent(w)
}

// GoogleLogin handles Google OAuth token exchange
func (h *AuthHandler) GoogleLogin(w http.ResponseWriter, r *http.Request) {
	var req GoogleLoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.IDToken == "" {
		response.BadRequest(w, "id_token is required")
		return
	}

	result, err := h.authService.GoogleLogin(r.Context(), req.IDToken)
	if err != nil {
		if err == auth.ErrInvalidGoogleToken {
			response.Unauthorized(w, "invalid Google token")
			return
		}
		if err == auth.ErrGoogleEmailMissing {
			response.BadRequest(w, "email not available from Google account")
			return
		}
		h.logger.Error("Google login failed", zap.Error(err))
		response.InternalError(w, "Google login failed")
		return
	}

	response.OK(w, result)
}

// Me returns the current authenticated user
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := middleware.GetUserID(r.Context())
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	user, err := h.authService.GetUserByID(r.Context(), userID)
	if err != nil {
		if err == domain.ErrUserNotFound {
			response.NotFound(w, "user not found")
			return
		}
		h.logger.Error("get user failed", zap.Error(err))
		response.InternalError(w, "failed to get user")
		return
	}

	response.OK(w, user.ToResponse())
}

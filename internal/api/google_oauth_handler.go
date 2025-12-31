package api

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/locolive/backend/internal/auth"
	"github.com/locolive/backend/internal/config"
	"github.com/locolive/backend/internal/domain"
	"go.uber.org/zap"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

// GoogleOAuthHandler handles browser-based Google OAuth flow
type GoogleOAuthHandler struct {
	config      *oauth2.Config
	authService *domain.AuthService
	verifier    *auth.GoogleAuthVerifier
	logger      *zap.Logger
	appScheme   string // App deep link scheme (e.g., "locoliveapp")
}

// NewGoogleOAuthHandler creates a new Google OAuth handler
func NewGoogleOAuthHandler(
	cfg *config.Config,
	authService *domain.AuthService,
	verifier *auth.GoogleAuthVerifier,
	logger *zap.Logger,
) *GoogleOAuthHandler {

	// Use the first configured client ID for the web flow, or empty if none
	clientID := ""
	if len(cfg.Google.ClientIDs) > 0 {
		clientID = cfg.Google.ClientIDs[0]
	}

	conf := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: cfg.Google.ClientSecret,
		RedirectURL:  "https://launchit.co.in/auth/google/callback", // TODO: Make configurable
		Scopes: []string{
			"https://www.googleapis.com/auth/userinfo.email",
			"https://www.googleapis.com/auth/userinfo.profile",
		},
		Endpoint: google.Endpoint,
	}

	return &GoogleOAuthHandler{
		config:      conf,
		authService: authService,
		verifier:    verifier,
		logger:      logger,
		appScheme:   "locoliveapp",
	}
}

// GoogleOAuthLogin initiates the Google OAuth flow by redirecting to Google
func (h *GoogleOAuthHandler) GoogleOAuthLogin(w http.ResponseWriter, r *http.Request) {
	// Generate state for CSRF protection (in production, store this in session)
	state := "random-state-string" // TODO: Generate and store proper state

	// Generate the Google OAuth URL
	authURL := h.config.AuthCodeURL(state, oauth2.AccessTypeOffline)

	h.logger.Info("Redirecting to Google OAuth", zap.String("url", authURL))

	// Redirect user to Google login
	http.Redirect(w, r, authURL, http.StatusTemporaryRedirect)
}

// GoogleOAuthCallback handles the callback from Google after authentication
func (h *GoogleOAuthHandler) GoogleOAuthCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	// Get the authorization code from the query params
	code := r.URL.Query().Get("code")
	if code == "" {
		h.logger.Error("No code in callback")
		h.redirectWithError(w, r, "Authorization code missing")
		return
	}

	// TODO: Verify state parameter for CSRF protection

	// Exchange the code for tokens
	token, err := h.config.Exchange(ctx, code)
	if err != nil {
		h.logger.Error("Failed to exchange code for token", zap.Error(err))
		h.redirectWithError(w, r, "Failed to authenticate with Google")
		return
	}

	// Get the ID token from the response
	idToken, ok := token.Extra("id_token").(string)
	if !ok {
		h.logger.Error("No ID token in response")
		h.redirectWithError(w, r, "Failed to get user info from Google")
		return
	}

	// Use the existing GoogleLogin service method to create/login user
	result, err := h.authService.GoogleLogin(ctx, idToken)
	if err != nil {
		h.logger.Error("Failed to login user", zap.Error(err))
		h.redirectWithError(w, r, "Failed to create user account")
		return
	}

	// Redirect back to the app with the tokens as query params
	h.redirectWithSuccess(w, r, result.AccessToken, result.RefreshToken, result.User.ID.String())
}

// redirectWithSuccess redirects to the app with auth tokens
func (h *GoogleOAuthHandler) redirectWithSuccess(w http.ResponseWriter, r *http.Request, accessToken, refreshToken, userID string) {
	// Create deep link URL: locoliveapp://auth/callback?access_token=xxx&refresh_token=yyy
	appURL := fmt.Sprintf("%s://auth/callback?access_token=%s&refresh_token=%s&user_id=%s",
		h.appScheme,
		url.QueryEscape(accessToken),
		url.QueryEscape(refreshToken),
		url.QueryEscape(userID),
	)

	h.logger.Info("Redirecting to app with tokens", zap.String("scheme", h.appScheme))

	http.Redirect(w, r, appURL, http.StatusTemporaryRedirect)
}

// redirectWithError redirects to the app with an error message
func (h *GoogleOAuthHandler) redirectWithError(w http.ResponseWriter, r *http.Request, errorMsg string) {
	appURL := fmt.Sprintf("%s://auth/callback?error=%s",
		h.appScheme,
		url.QueryEscape(errorMsg),
	)

	h.logger.Error("Redirecting to app with error", zap.String("error", errorMsg))

	http.Redirect(w, r, appURL, http.StatusTemporaryRedirect)
}

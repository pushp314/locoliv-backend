package auth

import (
	"context"
	"errors"

	"google.golang.org/api/idtoken"
)

var (
	ErrInvalidGoogleToken = errors.New("invalid Google ID token")
	ErrGoogleEmailMissing = errors.New("email not found in Google token")
)

// GoogleUser represents the user info from Google
type GoogleUser struct {
	GoogleID      string
	Email         string
	EmailVerified bool
	Name          string
	Picture       string
}

// GoogleAuthVerifier handles Google ID token verification
type GoogleAuthVerifier struct {
	clientID string
}

// NewGoogleAuthVerifier creates a new Google auth verifier
func NewGoogleAuthVerifier(clientID string) *GoogleAuthVerifier {
	return &GoogleAuthVerifier{
		clientID: clientID,
	}
}

// VerifyIDToken verifies a Google ID token and returns the user info
func (v *GoogleAuthVerifier) VerifyIDToken(ctx context.Context, idToken string) (*GoogleUser, error) {
	payload, err := idtoken.Validate(ctx, idToken, v.clientID)
	if err != nil {
		return nil, ErrInvalidGoogleToken
	}

	// Extract user info from claims
	googleUser := &GoogleUser{}

	// Google ID (sub claim)
	if sub, ok := payload.Claims["sub"].(string); ok {
		googleUser.GoogleID = sub
	} else {
		return nil, ErrInvalidGoogleToken
	}

	// Email
	if email, ok := payload.Claims["email"].(string); ok {
		googleUser.Email = email
	} else {
		return nil, ErrGoogleEmailMissing
	}

	// Email verified
	if verified, ok := payload.Claims["email_verified"].(bool); ok {
		googleUser.EmailVerified = verified
	}

	// Name
	if name, ok := payload.Claims["name"].(string); ok {
		googleUser.Name = name
	}

	// Picture
	if picture, ok := payload.Claims["picture"].(string); ok {
		googleUser.Picture = picture
	}

	return googleUser, nil
}

// IsConfigured returns true if Google OAuth is configured
func (v *GoogleAuthVerifier) IsConfigured() bool {
	return v.clientID != ""
}

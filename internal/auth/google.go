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
	clientIDs []string
}

// NewGoogleAuthVerifier creates a new Google auth verifier
func NewGoogleAuthVerifier(clientIDs []string) *GoogleAuthVerifier {
	return &GoogleAuthVerifier{
		clientIDs: clientIDs,
	}
}

// VerifyIDToken verifies a Google ID token and returns the user info
func (v *GoogleAuthVerifier) VerifyIDToken(ctx context.Context, idToken string) (*GoogleUser, error) {
	// Try to validate with each client ID
	var payload *idtoken.Payload
	var err error

	for _, clientID := range v.clientIDs {
		payload, err = idtoken.Validate(ctx, idToken, clientID)
		if err == nil {
			break
		}
	}

	if payload == nil {
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
	return len(v.clientIDs) > 0 && v.clientIDs[0] != ""
}

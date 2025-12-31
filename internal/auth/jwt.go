package auth

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	ErrInvalidToken = errors.New("invalid token")
	ErrExpiredToken = errors.New("token has expired")
)

// TokenType distinguishes between access and refresh tokens
type TokenType string

const (
	AccessToken  TokenType = "access"
	RefreshToken TokenType = "refresh"
)

// Claims represents the JWT claims
type Claims struct {
	UserID    uuid.UUID `json:"user_id"`
	Email     string    `json:"email,omitempty"`
	TokenType TokenType `json:"token_type"`
	jwt.RegisteredClaims
}

// JWTManager handles JWT operations
type JWTManager struct {
	secret        []byte
	accessExpiry  time.Duration
	refreshExpiry time.Duration
	issuer        string
}

// NewJWTManager creates a new JWT manager
func NewJWTManager(secret string, accessExpiry, refreshExpiry time.Duration) *JWTManager {
	return &JWTManager{
		secret:        []byte(secret),
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
		issuer:        "locolive",
	}
}

// GenerateAccessToken creates a new access token
func (m *JWTManager) GenerateAccessToken(userID uuid.UUID, email string) (string, error) {
	now := time.Now()
	claims := &Claims{
		UserID:    userID,
		Email:     email,
		TokenType: AccessToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(m.accessExpiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(m.secret)
}

// GenerateRefreshToken creates a new refresh token
func (m *JWTManager) GenerateRefreshToken(userID uuid.UUID) (string, time.Time, error) {
	now := time.Now()
	expiresAt := now.Add(m.refreshExpiry)

	claims := &Claims{
		UserID:    userID,
		TokenType: RefreshToken,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    m.issuer,
			Subject:   userID.String(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(m.secret)
	return signed, expiresAt, err
}

// ValidateToken validates a JWT and returns the claims
func (m *JWTManager) ValidateToken(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, ErrInvalidToken
		}
		return m.secret, nil
	})

	if err != nil {
		if errors.Is(err, jwt.ErrTokenExpired) {
			return nil, ErrExpiredToken
		}
		return nil, ErrInvalidToken
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateAccessToken validates an access token
func (m *JWTManager) ValidateAccessToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != AccessToken {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// ValidateRefreshToken validates a refresh token
func (m *JWTManager) ValidateRefreshToken(tokenString string) (*Claims, error) {
	claims, err := m.ValidateToken(tokenString)
	if err != nil {
		return nil, err
	}

	if claims.TokenType != RefreshToken {
		return nil, ErrInvalidToken
	}

	return claims, nil
}

// GenerateSecureToken generates a cryptographically secure random token
func GenerateSecureToken(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

// TokenPair represents an access/refresh token pair
type TokenPair struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"`
}

// GenerateTokenPair creates both access and refresh tokens
func (m *JWTManager) GenerateTokenPair(userID uuid.UUID, email string) (*TokenPair, error) {
	accessToken, err := m.GenerateAccessToken(userID, email)
	if err != nil {
		return nil, err
	}

	refreshToken, expiresAt, err := m.GenerateRefreshToken(userID)
	if err != nil {
		return nil, err
	}

	return &TokenPair{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		ExpiresAt:    expiresAt,
	}, nil
}

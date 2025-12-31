package auth

import (
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordTooShort = errors.New("password must be at least 8 characters")
	ErrPasswordMismatch = errors.New("incorrect password")
)

const (
	MinPasswordLength = 8
	bcryptCost        = 12
)

// HashPassword creates a bcrypt hash of the password
func HashPassword(password string) (string, error) {
	if len(password) < MinPasswordLength {
		return "", ErrPasswordTooShort
	}

	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcryptCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

// VerifyPassword compares a password with a hash
func VerifyPassword(password, hash string) error {
	err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(password))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return ErrPasswordMismatch
		}
		return err
	}
	return nil
}

// HashToken creates a SHA-256 hash of a token (for storing refresh tokens)
func HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

// CompareTokenHash compares a token with its hash
func CompareTokenHash(token, hash string) bool {
	tokenHash := HashToken(token)
	return subtle.ConstantTimeCompare([]byte(tokenHash), []byte(hash)) == 1
}

// ValidatePasswordStrength checks if password meets requirements
func ValidatePasswordStrength(password string) error {
	if len(password) < MinPasswordLength {
		return ErrPasswordTooShort
	}
	// Add more validation rules as needed:
	// - uppercase letters
	// - lowercase letters
	// - numbers
	// - special characters
	return nil
}

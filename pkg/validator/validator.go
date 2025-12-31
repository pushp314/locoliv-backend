package validator

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode"
)

var (
	phoneRegex = regexp.MustCompile(`^\+?[1-9]\d{9,14}$`)
)

// ValidationError represents a validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

// Error implements the error interface
func (v ValidationErrors) Error() string {
	if len(v) == 0 {
		return ""
	}
	var msgs []string
	for _, e := range v {
		msgs = append(msgs, e.Field+": "+e.Message)
	}
	return strings.Join(msgs, "; ")
}

// HasErrors returns true if there are any errors
func (v ValidationErrors) HasErrors() bool {
	return len(v) > 0
}

// Add adds a validation error
func (v *ValidationErrors) Add(field, message string) {
	*v = append(*v, ValidationError{Field: field, Message: message})
}

// ValidateEmail validates an email address
func ValidateEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// ValidatePhone validates a phone number
func ValidatePhone(phone string) bool {
	// Remove spaces and dashes
	cleaned := strings.ReplaceAll(phone, " ", "")
	cleaned = strings.ReplaceAll(cleaned, "-", "")
	return phoneRegex.MatchString(cleaned)
}

// ValidatePassword validates password strength
func ValidatePassword(password string) ValidationErrors {
	var errors ValidationErrors

	if len(password) < 8 {
		errors.Add("password", "must be at least 8 characters")
		return errors
	}

	var hasUpper, hasLower, hasNumber bool
	for _, c := range password {
		switch {
		case unicode.IsUpper(c):
			hasUpper = true
		case unicode.IsLower(c):
			hasLower = true
		case unicode.IsDigit(c):
			hasNumber = true
		}
	}

	if !hasUpper {
		errors.Add("password", "must contain at least one uppercase letter")
	}
	if !hasLower {
		errors.Add("password", "must contain at least one lowercase letter")
	}
	if !hasNumber {
		errors.Add("password", "must contain at least one number")
	}

	return errors
}

// ValidateName validates a user name
func ValidateName(name string) bool {
	name = strings.TrimSpace(name)
	return len(name) >= 2 && len(name) <= 100
}

// SanitizeString trims whitespace and limits length
func SanitizeString(s string, maxLen int) string {
	s = strings.TrimSpace(s)
	if len(s) > maxLen {
		return s[:maxLen]
	}
	return s
}

// SanitizeEmail normalizes an email address
func SanitizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

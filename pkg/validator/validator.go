package validator

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/locolive/backend/pkg/response"
)

// Common validation errors
var (
	ErrRequired       = errors.New("field is required")
	ErrTooShort       = errors.New("field is too short")
	ErrTooLong        = errors.New("field is too long")
	ErrInvalidFormat  = errors.New("invalid format")
	ErrInvalidEmail   = errors.New("invalid email format")
	ErrInvalidPhone   = errors.New("invalid phone number")
	ErrContainsHTML   = errors.New("HTML content not allowed")
	ErrContainsScript = errors.New("script content not allowed")
)

// Validation patterns
var (
	// Indian phone number pattern
	phonePattern = regexp.MustCompile(`^(\+91)?[6-9]\d{9}$`)
	// Alphanumeric with spaces and common punctuation
	safeTextPattern = regexp.MustCompile(`^[\p{L}\p{N}\s.,!?'-]*$`)
	// Username pattern (alphanumeric, underscore, dot)
	usernamePattern = regexp.MustCompile(`^[a-zA-Z0-9_.]{3,30}$`)
	// HTML/Script detection
	htmlPattern   = regexp.MustCompile(`<[^>]*>`)
	scriptPattern = regexp.MustCompile(`(?i)<script|javascript:|on\w+\s*=`)
	// SQL injection patterns
	sqlInjectionPattern = regexp.MustCompile(`(?i)(union|select|insert|update|delete|drop|alter|exec|execute|xp_|sp_|0x)`)
)

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationErrors is a collection of validation errors
type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	if len(ve) == 0 {
		return "validation failed"
	}
	msgs := make([]string, len(ve))
	for i, e := range ve {
		msgs[i] = fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return strings.Join(msgs, "; ")
}

// HasErrors returns true if there are errors
func (ve ValidationErrors) HasErrors() bool {
	return len(ve) > 0
}

// Validator provides validation methods
type Validator struct {
	errors ValidationErrors
}

// New creates a new validator
func New() *Validator {
	return &Validator{errors: make(ValidationErrors, 0)}
}

// HasErrors returns true if there are validation errors
func (v *Validator) HasErrors() bool {
	return len(v.errors) > 0
}

// Errors returns all validation errors
func (v *Validator) Errors() ValidationErrors {
	return v.errors
}

// AddError adds a validation error
func (v *Validator) AddError(field, message string) {
	v.errors = append(v.errors, ValidationError{Field: field, Message: message})
}

// Required checks if a string is not empty
func (v *Validator) Required(field, value string) bool {
	if strings.TrimSpace(value) == "" {
		v.AddError(field, "is required")
		return false
	}
	return true
}

// MinLength checks minimum string length
func (v *Validator) MinLength(field, value string, min int) bool {
	if utf8.RuneCountInString(value) < min {
		v.AddError(field, fmt.Sprintf("must be at least %d characters", min))
		return false
	}
	return true
}

// MaxLength checks maximum string length
func (v *Validator) MaxLength(field, value string, max int) bool {
	if utf8.RuneCountInString(value) > max {
		v.AddError(field, fmt.Sprintf("must be at most %d characters", max))
		return false
	}
	return true
}

// Email validates email format
func (v *Validator) Email(field, value string) bool {
	if value == "" {
		return true // Empty is valid, use Required for non-empty
	}
	_, err := mail.ParseAddress(value)
	if err != nil {
		v.AddError(field, "invalid email format")
		return false
	}
	return true
}

// Phone validates Indian phone number
func (v *Validator) Phone(field, value string) bool {
	if value == "" {
		return true
	}
	if !phonePattern.MatchString(value) {
		v.AddError(field, "invalid phone number (use +91XXXXXXXXXX format)")
		return false
	}
	return true
}

// Username validates username format
func (v *Validator) Username(field, value string) bool {
	if !usernamePattern.MatchString(value) {
		v.AddError(field, "must be 3-30 characters, alphanumeric with _ or .")
		return false
	}
	return true
}

// NoHTML checks for HTML content
func (v *Validator) NoHTML(field, value string) bool {
	if htmlPattern.MatchString(value) {
		v.AddError(field, "HTML content not allowed")
		return false
	}
	return true
}

// NoScript checks for script content
func (v *Validator) NoScript(field, value string) bool {
	if scriptPattern.MatchString(value) {
		v.AddError(field, "script content not allowed")
		return false
	}
	return true
}

// NoSQLInjection checks for SQL injection patterns
func (v *Validator) NoSQLInjection(field, value string) bool {
	if sqlInjectionPattern.MatchString(value) {
		v.AddError(field, "invalid characters detected")
		return false
	}
	return true
}

// SafeText validates text is safe for storage
func (v *Validator) SafeText(field, value string) bool {
	v.NoHTML(field, value)
	v.NoScript(field, value)
	return !v.HasErrors()
}

// IsSafeText checks if text matches safe text pattern (alphanumeric, spaces, punctuation)
func IsSafeText(s string) bool {
	return safeTextPattern.MatchString(s)
}

// Range checks if an integer is within range
func (v *Validator) Range(field string, value, min, max int) bool {
	if value < min || value > max {
		v.AddError(field, fmt.Sprintf("must be between %d and %d", min, max))
		return false
	}
	return true
}

// OneOf checks if value is one of allowed values
func (v *Validator) OneOf(field, value string, allowed []string) bool {
	for _, a := range allowed {
		if value == a {
			return true
		}
	}
	v.AddError(field, fmt.Sprintf("must be one of: %s", strings.Join(allowed, ", ")))
	return false
}

// Sanitize removes potentially dangerous content from a string
func Sanitize(s string) string {
	// Remove HTML tags
	s = htmlPattern.ReplaceAllString(s, "")
	// Trim whitespace
	s = strings.TrimSpace(s)
	// Limit length
	if len(s) > 10000 {
		s = s[:10000]
	}
	return s
}

// SanitizeMap sanitizes all string values in a map
func SanitizeMap(m map[string]interface{}) {
	for k, v := range m {
		if str, ok := v.(string); ok {
			m[k] = Sanitize(str)
		}
	}
}

// ValidateJSON validates and decodes JSON request body
func ValidateJSON(r *http.Request, dest interface{}) error {
	if r.Body == nil {
		return errors.New("request body is empty")
	}

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Strict mode

	if err := decoder.Decode(dest); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	return nil
}

// ValidationMiddleware creates a middleware for request validation
func ValidationMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check Content-Type for POST/PUT/PATCH
		if r.Method == http.MethodPost || r.Method == http.MethodPut || r.Method == http.MethodPatch {
			contentType := r.Header.Get("Content-Type")
			if contentType != "" &&
				!strings.HasPrefix(contentType, "application/json") &&
				!strings.HasPrefix(contentType, "multipart/form-data") {
				response.BadRequest(w, "unsupported content type")
				return
			}
		}

		// Limit request body size (10MB)
		r.Body = http.MaxBytesReader(w, r.Body, 10<<20)

		next.ServeHTTP(w, r)
	})
}

// RateLimitKey generates a rate limit key from request
func RateLimitKey(r *http.Request, prefix string) string {
	// Use X-Forwarded-For if behind proxy, otherwise RemoteAddr
	ip := r.Header.Get("X-Forwarded-For")
	if ip == "" {
		ip = r.Header.Get("X-Real-IP")
	}
	if ip == "" {
		ip = r.RemoteAddr
	}
	// Take first IP if multiple
	if idx := strings.Index(ip, ","); idx > 0 {
		ip = ip[:idx]
	}
	return fmt.Sprintf("%s:%s", prefix, strings.TrimSpace(ip))
}

// ========== Standalone validation functions ==========

// SanitizeEmail cleans and normalizes an email address
func SanitizeEmail(email string) string {
	email = strings.TrimSpace(email)
	email = strings.ToLower(email)
	return email
}

// ValidateEmail checks if email is valid
func ValidateEmail(email string) bool {
	if email == "" {
		return false
	}
	_, err := mail.ParseAddress(email)
	return err == nil
}

// SanitizeString cleans a string and limits its length
func SanitizeString(s string, maxLen int) string {
	// Trim whitespace
	s = strings.TrimSpace(s)
	// Remove HTML tags
	s = htmlPattern.ReplaceAllString(s, "")
	// Limit length
	if len(s) > maxLen {
		s = s[:maxLen]
	}
	return s
}

// ValidateName checks if a name is valid (2-100 chars)
func ValidateName(name string) bool {
	length := utf8.RuneCountInString(name)
	return length >= 2 && length <= 100
}

// ValidatePassword checks password strength and returns errors
func ValidatePassword(password string) ValidationErrors {
	var errs ValidationErrors

	if len(password) < 8 {
		errs = append(errs, ValidationError{Field: "password", Message: "must be at least 8 characters"})
	}
	if len(password) > 128 {
		errs = append(errs, ValidationError{Field: "password", Message: "must be at most 128 characters"})
	}

	var hasUpper, hasLower, hasNumber bool
	for _, c := range password {
		switch {
		case c >= 'A' && c <= 'Z':
			hasUpper = true
		case c >= 'a' && c <= 'z':
			hasLower = true
		case c >= '0' && c <= '9':
			hasNumber = true
		}
	}

	if !hasUpper {
		errs = append(errs, ValidationError{Field: "password", Message: "must contain at least one uppercase letter"})
	}
	if !hasLower {
		errs = append(errs, ValidationError{Field: "password", Message: "must contain at least one lowercase letter"})
	}
	if !hasNumber {
		errs = append(errs, ValidationError{Field: "password", Message: "must contain at least one number"})
	}

	return errs
}

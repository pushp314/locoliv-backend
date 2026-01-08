package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/locolive/backend/pkg/response"
	"github.com/locolive/backend/pkg/validator"
)

// RateLimiter implements a simple in-memory rate limiter
// For production, use Redis-based rate limiting
type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.RWMutex
	limit    int           // Max requests
	window   time.Duration // Time window
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	rl := &RateLimiter{
		requests: make(map[string][]time.Time),
		limit:    limit,
		window:   window,
	}

	// Start cleanup goroutine
	go rl.cleanup()

	return rl
}

// Allow checks if a request should be allowed
func (rl *RateLimiter) Allow(key string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	windowStart := now.Add(-rl.window)

	// Get existing requests for this key
	times := rl.requests[key]

	// Filter to only requests within window
	var validTimes []time.Time
	for _, t := range times {
		if t.After(windowStart) {
			validTimes = append(validTimes, t)
		}
	}

	// Check if over limit
	if len(validTimes) >= rl.limit {
		rl.requests[key] = validTimes
		return false
	}

	// Add new request
	validTimes = append(validTimes, now)
	rl.requests[key] = validTimes

	return true
}

// cleanup removes old entries periodically
func (rl *RateLimiter) cleanup() {
	ticker := time.NewTicker(rl.window)
	defer ticker.Stop()

	for range ticker.C {
		rl.mu.Lock()
		now := time.Now()
		windowStart := now.Add(-rl.window)

		for key, times := range rl.requests {
			var validTimes []time.Time
			for _, t := range times {
				if t.After(windowStart) {
					validTimes = append(validTimes, t)
				}
			}
			if len(validTimes) == 0 {
				delete(rl.requests, key)
			} else {
				rl.requests[key] = validTimes
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimitMiddleware creates a rate limiting middleware
func RateLimitMiddleware(limiter *RateLimiter, prefix string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			key := validator.RateLimitKey(r, prefix)

			if !limiter.Allow(key) {
				w.Header().Set("Retry-After", "60")
				response.Error(w, http.StatusTooManyRequests, "RATE_LIMIT_EXCEEDED", "rate limit exceeded, please try again later")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}

// Common rate limiters
var (
	// GlobalLimiter: 100 requests per minute
	GlobalLimiter = NewRateLimiter(100, time.Minute)

	// AuthLimiter: 10 login attempts per minute
	AuthLimiter = NewRateLimiter(10, time.Minute)

	// UploadLimiter: 20 uploads per minute
	UploadLimiter = NewRateLimiter(20, time.Minute)

	// APILimiter: 60 API calls per minute
	APILimiter = NewRateLimiter(60, time.Minute)
)

// RateLimitAuth rate limits authentication endpoints
func RateLimitAuth(next http.Handler) http.Handler {
	return RateLimitMiddleware(AuthLimiter, "auth")(next)
}

// RateLimitUpload rate limits upload endpoints
func RateLimitUpload(next http.Handler) http.Handler {
	return RateLimitMiddleware(UploadLimiter, "upload")(next)
}

// RateLimitAPI rate limits general API endpoints
func RateLimitAPI(next http.Handler) http.Handler {
	return RateLimitMiddleware(APILimiter, "api")(next)
}

// SecurityHeaders adds security headers to responses
func SecurityHeaders(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Prevent MIME type sniffing
		w.Header().Set("X-Content-Type-Options", "nosniff")

		// Prevent clickjacking
		w.Header().Set("X-Frame-Options", "DENY")

		// Enable XSS filter
		w.Header().Set("X-XSS-Protection", "1; mode=block")

		// Strict transport security (HTTPS only)
		w.Header().Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		// Content Security Policy
		w.Header().Set("Content-Security-Policy", "default-src 'self'")

		// Referrer policy
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")

		next.ServeHTTP(w, r)
	})
}

// RequestSizeLimit limits the request body size
func RequestSizeLimit(maxBytes int64) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			r.Body = http.MaxBytesReader(w, r.Body, maxBytes)
			next.ServeHTTP(w, r)
		})
	}
}

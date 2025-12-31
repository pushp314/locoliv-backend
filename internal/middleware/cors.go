package middleware

import (
	"net/http"

	"github.com/go-chi/cors"
)

// CORSMiddleware returns CORS configuration for mobile clients
func CORSMiddleware() func(next http.Handler) http.Handler {
	return cors.Handler(cors.Options{
		// Allow all origins for mobile apps
		// In production, you may want to restrict this
		AllowedOrigins: []string{"*"},

		// Allow common HTTP methods
		AllowedMethods: []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},

		// Allow common headers
		AllowedHeaders: []string{
			"Accept",
			"Authorization",
			"Content-Type",
			"X-CSRF-Token",
			"X-Requested-With",
		},

		// Expose headers to the client
		ExposedHeaders: []string{
			"Link",
			"X-Request-Id",
		},

		// Allow credentials (cookies, authorization headers)
		AllowCredentials: true,

		// Cache preflight requests for 5 minutes
		MaxAge: 300,
	})
}

package api

import (
	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/locolive/backend/internal/auth"
	"github.com/locolive/backend/internal/middleware"
	"go.uber.org/zap"
)

// Router holds all handlers and creates the chi router
type Router struct {
	authHandler   *AuthHandler
	healthHandler *HealthHandler
	jwtManager    *auth.JWTManager
	logger        *zap.Logger
}

// NewRouter creates a new router
func NewRouter(
	authHandler *AuthHandler,
	healthHandler *HealthHandler,
	jwtManager *auth.JWTManager,
	logger *zap.Logger,
) *Router {
	return &Router{
		authHandler:   authHandler,
		healthHandler: healthHandler,
		jwtManager:    jwtManager,
		logger:        logger,
	}
}

// Setup configures and returns the chi router
func (rt *Router) Setup() *chi.Mux {
	r := chi.NewRouter()

	// Global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(middleware.RecoveryMiddleware(rt.logger))
	r.Use(middleware.LoggingMiddleware(rt.logger))
	r.Use(middleware.CORSMiddleware())
	r.Use(chimiddleware.Compress(5))

	// Health endpoints (no auth required)
	r.Route("/health", func(r chi.Router) {
		r.Get("/", rt.healthHandler.Health)
		r.Get("/ready", rt.healthHandler.Ready)
		r.Get("/live", rt.healthHandler.Live)
	})

	// API v1
	r.Route("/api/v1", func(r chi.Router) {
		// Auth routes (no auth required)
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", rt.authHandler.Register)
			r.Post("/login", rt.authHandler.Login)
			r.Post("/refresh", rt.authHandler.Refresh)
			r.Post("/logout", rt.authHandler.Logout)
			r.Post("/google", rt.authHandler.GoogleLogin)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(rt.jwtManager))

			// User routes
			r.Get("/me", rt.authHandler.Me)
			r.Post("/auth/logout-all", rt.authHandler.LogoutAll)
		})
	})

	// Also expose auth at root level for compatibility
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", rt.authHandler.Register)
		r.Post("/login", rt.authHandler.Login)
		r.Post("/refresh", rt.authHandler.Refresh)
		r.Post("/logout", rt.authHandler.Logout)
		r.Post("/google", rt.authHandler.GoogleLogin)
	})

	return r
}

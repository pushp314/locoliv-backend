package api

import (
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	"github.com/locolive/backend/internal/auth"
	"github.com/locolive/backend/internal/middleware"
	"go.uber.org/zap"
)

type Router struct {
	authHandler         *AuthHandler
	googleOAuthHandler  *GoogleOAuthHandler
	storyHandler        *StoryHandler
	chatHandler         *ChatHandler
	connectionHandler   *ConnectionHandler
	notificationHandler *NotificationHandler
	healthHandler       *HealthHandler
	jwtManager          *auth.JWTManager
	logger              *zap.Logger
}

// NewRouter creates a new router
func NewRouter(
	authHandler *AuthHandler,
	googleOAuthHandler *GoogleOAuthHandler,
	storyHandler *StoryHandler,
	chatHandler *ChatHandler,
	connectionHandler *ConnectionHandler,
	notificationHandler *NotificationHandler,
	healthHandler *HealthHandler,
	jwtManager *auth.JWTManager,
	logger *zap.Logger,
) *Router {
	return &Router{
		authHandler:         authHandler,
		googleOAuthHandler:  googleOAuthHandler,
		storyHandler:        storyHandler,
		chatHandler:         chatHandler,
		connectionHandler:   connectionHandler,
		notificationHandler: notificationHandler,
		healthHandler:       healthHandler,
		jwtManager:          jwtManager,
		logger:              logger,
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

	// Serve static files from uploads directory
	workDir, _ := os.Getwd()
	filesDir := http.Dir(filepath.Join(workDir, "uploads"))
	FileServer(r, "/uploads", filesDir)

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
			r.Post("/forgot-password", rt.authHandler.ForgotPassword)
			r.Post("/reset-password", rt.authHandler.ResetPassword)
		})

		// Protected routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.AuthMiddleware(rt.jwtManager))

			// User routes
			r.Get("/me", rt.authHandler.Me)
			r.Get("/users/{userId}", rt.authHandler.GetProfile)
			r.Post("/auth/logout-all", rt.authHandler.LogoutAll)
			r.Put("/auth/password", rt.authHandler.UpdatePassword)
			r.Put("/auth/email", rt.authHandler.UpdateEmail)
			r.Put("/auth/profile", rt.authHandler.UpdateProfile)

			// Story routes
			r.Route("/stories", func(r chi.Router) {
				r.Post("/", rt.storyHandler.CreateStory)
				r.Get("/feed", rt.storyHandler.GetFeed)
			})

			// Chat routes
			r.Route("/chats", func(r chi.Router) {
				r.Post("/", rt.chatHandler.CreateChat)
				r.Get("/", rt.chatHandler.GetChats)
				r.Get("/{chatId}/messages", rt.chatHandler.GetMessages)
				r.Post("/{chatId}/messages", rt.chatHandler.SendMessage)
			})

			// Connection routes
			r.Route("/connections", func(r chi.Router) {
				r.Post("/request", rt.connectionHandler.SendRequest)
				r.Post("/respond", rt.connectionHandler.RespondRequest)
				r.Get("/", rt.connectionHandler.GetConnections)
				r.Get("/requests", rt.connectionHandler.GetRequests)
			})

			// Notification routes
			r.Route("/notifications", func(r chi.Router) {
				r.Get("/", rt.notificationHandler.GetNotifications)
				r.Put("/{id}/read", rt.notificationHandler.MarkRead)
				r.Post("/fcm-token", rt.notificationHandler.UpdateFCMToken)
			})
		})
	})

	// Auth routes at root level for compatibility
	r.Route("/auth", func(r chi.Router) {
		r.Post("/register", rt.authHandler.Register)
		r.Post("/login", rt.authHandler.Login)
		r.Post("/refresh", rt.authHandler.Refresh)
		r.Post("/logout", rt.authHandler.Logout)
		r.Post("/google", rt.authHandler.GoogleLogin)

		// Browser-based Google OAuth (for mobile in-app browser)
		r.Get("/google/login", rt.googleOAuthHandler.GoogleOAuthLogin)
		r.Get("/google/callback", rt.googleOAuthHandler.GoogleOAuthCallback)
	})

	// WebSocket routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.AuthMiddleware(rt.jwtManager))
		r.Get("/ws/chat", rt.chatHandler.HandleWebSocket)
	})

	return r
}

// FileServer conveniently sets up a http.FileServer handler at the given path
func FileServer(r chi.Router, path string, root http.FileSystem) {
	if strings.ContainsAny(path, "{}*") {
		panic("FileServer does not permit any URL parameters.")
	}

	if path != "/" && path[len(path)-1] != '/' {
		r.Get(path, http.RedirectHandler(path+"/", 301).ServeHTTP)
		path += "/"
	}
	path += "*"

	r.Get(path, func(w http.ResponseWriter, r *http.Request) {
		rctx := chi.RouteContext(r.Context())
		pathPrefix := strings.TrimSuffix(rctx.RoutePattern(), "/*")
		fs := http.StripPrefix(pathPrefix, http.FileServer(root))
		fs.ServeHTTP(w, r)
	})
}

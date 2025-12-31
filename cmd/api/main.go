package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"go.uber.org/zap"

	"github.com/locolive/backend/internal/api"
	"github.com/locolive/backend/internal/auth"
	"github.com/locolive/backend/internal/config"
	"github.com/locolive/backend/internal/domain"
	"github.com/locolive/backend/internal/fcm"
	"github.com/locolive/backend/internal/repository"
	"github.com/locolive/backend/internal/storage"
)

func main() {
	// Load .env file if exists
	_ = godotenv.Load()

	// Initialize logger
	logger, err := initLogger()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer logger.Sync()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		logger.Fatal("Failed to load config", zap.Error(err))
	}

	logger.Info("Starting LocoLive API",
		zap.String("env", cfg.Server.Env),
		zap.String("port", cfg.Server.Port),
	)

	// Initialize database
	ctx := context.Background()
	db, err := initDatabase(ctx, cfg.Database.URL)
	if err != nil {
		logger.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer db.Close()

	logger.Info("Connected to database")

	// Initialize dependencies
	repo := repository.NewPostgresRepository(db)
	jwtManager := auth.NewJWTManager(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
	googleAuth := auth.NewGoogleAuthVerifier(cfg.Google.ClientID)

	// Log Google OAuth status
	if googleAuth.IsConfigured() {
		logger.Info("Google OAuth is configured")
	} else {
		logger.Warn("Google OAuth is NOT configured - set GOOGLE_CLIENT_ID to enable")
	}

	// Initialize Firebase
	fcmClient, err := fcm.NewClient(ctx, logger, os.Getenv("GOOGLE_APPLICATION_CREDENTIALS"))
	if err != nil {
		logger.Warn("Failed to initialize Firebase client - push notifications will be disabled", zap.Error(err))
	} else {
		logger.Info("Firebase client initialized")
	}

	// Initialize storage
	// Ensure upload directory exists
	uploadDir := "./uploads"
	baseURL := fmt.Sprintf("http://localhost:%s/uploads", cfg.Server.Port)
	if cfg.Server.Env == "production" {
		// In production, might be different or use S3, but for now local
		baseURL = "https://api.locolive.com/uploads" // Adjust as needed
	}

	fileStorage, err := storage.NewLocalFileStorage(uploadDir, baseURL)
	if err != nil {
		logger.Fatal("Failed to initialize file storage", zap.Error(err))
	}

	// Initialize services
	authService := domain.NewAuthService(repo, jwtManager, googleAuth)
	storyService := domain.NewStoryService(repo, fileStorage)
	chatService := domain.NewChatService(repo)
	connectionService := domain.NewConnectionService(repo)
	notificationService := domain.NewNotificationService(repo, fcmClient)

	// Initialize WebSocket manager
	wsManager := api.NewWebSocketManager(logger)
	go wsManager.Run()

	// Initialize handlers
	authHandler := api.NewAuthHandler(authService, repo, logger)
	googleOAuthHandler := api.NewGoogleOAuthHandler(cfg, authService, googleAuth, logger)
	storyHandler := api.NewStoryHandler(storyService, logger)
	chatHandler := api.NewChatHandler(chatService, wsManager, logger)
	connectionHandler := api.NewConnectionHandler(connectionService, logger)
	notificationHandler := api.NewNotificationHandler(notificationService, logger)
	healthHandler := api.NewHealthHandler()

	// Initialize router
	router := api.NewRouter(authHandler, googleOAuthHandler, storyHandler, chatHandler, connectionHandler, notificationHandler, healthHandler, jwtManager, logger)
	r := router.Setup()

	// Start cleanup worker
	cleanupCtx, cleanupCancel := context.WithCancel(ctx)
	repo.StartCleanupWorker(cleanupCtx, 1*time.Hour)

	// Create server
	srv := &http.Server{
		Addr:         ":" + cfg.Server.Port,
		Handler:      r,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Server listening", zap.String("addr", srv.Addr))
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Server failed", zap.Error(err))
		}
	}()

	// Wait for interrupt signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// Cancel cleanup worker
	cleanupCancel()

	// Graceful shutdown with timeout
	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 30*time.Second)
	defer shutdownCancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		logger.Error("Server shutdown error", zap.Error(err))
	}

	logger.Info("Server stopped")
}

func initLogger() (*zap.Logger, error) {
	env := os.Getenv("ENV")
	if env == "production" {
		return zap.NewProduction()
	}
	return zap.NewDevelopment()
}

func initDatabase(ctx context.Context, databaseURL string) (*pgxpool.Pool, error) {
	config, err := pgxpool.ParseConfig(databaseURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	// Connection pool settings
	config.MaxConns = 25
	config.MinConns = 5
	config.MaxConnLifetime = 1 * time.Hour
	config.MaxConnIdleTime = 30 * time.Minute
	config.HealthCheckPeriod = 1 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Test connection
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return pool, nil
}

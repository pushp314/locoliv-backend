package config

import (
	"os"
	"time"
)

// Config holds all application configuration
type Config struct {
	Server   ServerConfig
	Database DatabaseConfig
	Redis    RedisConfig
	JWT      JWTConfig
	Google   GoogleConfig
	Log      LogConfig
}

type ServerConfig struct {
	Port string
	Env  string
}

type DatabaseConfig struct {
	URL      string
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type RedisConfig struct {
	URL string
}

type JWTConfig struct {
	Secret        string
	AccessExpiry  time.Duration
	RefreshExpiry time.Duration
}

type GoogleConfig struct {
	ClientID     string
	ClientSecret string
}

type LogConfig struct {
	Level string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	accessExpiry, err := time.ParseDuration(getEnv("JWT_ACCESS_EXPIRY", "15m"))
	if err != nil {
		accessExpiry = 15 * time.Minute
	}

	refreshExpiry, err := time.ParseDuration(getEnv("JWT_REFRESH_EXPIRY", "168h"))
	if err != nil {
		refreshExpiry = 7 * 24 * time.Hour
	}

	return &Config{
		Server: ServerConfig{
			Port: getEnv("PORT", "8080"),
			Env:  getEnv("ENV", "development"),
		},
		Database: DatabaseConfig{
			URL:      getEnv("DATABASE_URL", "postgres://locolive:locolive@localhost:5432/locolive?sslmode=disable"),
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5432"),
			User:     getEnv("DB_USER", "locolive"),
			Password: getEnv("DB_PASSWORD", "locolive"),
			Name:     getEnv("DB_NAME", "locolive"),
		},
		Redis: RedisConfig{
			URL: getEnv("REDIS_URL", "redis://localhost:6379"),
		},
		JWT: JWTConfig{
			Secret:        getEnv("JWT_SECRET", "change-me-in-production"),
			AccessExpiry:  accessExpiry,
			RefreshExpiry: refreshExpiry,
		},
		Google: GoogleConfig{
			ClientID:     getEnv("GOOGLE_CLIENT_ID", ""),
			ClientSecret: getEnv("GOOGLE_CLIENT_SECRET", ""),
		},
		Log: LogConfig{
			Level: getEnv("LOG_LEVEL", "debug"),
		},
	}, nil
}

// getEnv gets an environment variable with a fallback default
func getEnv(key, fallback string) string {
	if value, exists := os.LookupEnv(key); exists {
		return value
	}
	return fallback
}

// IsProduction returns true if running in production
func (c *Config) IsProduction() bool {
	return c.Server.Env == "production"
}

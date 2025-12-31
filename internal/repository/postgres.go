package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/locolive/backend/internal/auth"
	"github.com/locolive/backend/internal/domain"
)

// PostgresRepository implements domain.AuthRepository using PostgreSQL
type PostgresRepository struct {
	db *pgxpool.Pool
}

// NewPostgresRepository creates a new PostgreSQL repository
func NewPostgresRepository(db *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{db: db}
}

// CreateUser creates a new user
func (r *PostgresRepository) CreateUser(ctx context.Context, params domain.CreateUserParams) (*domain.User, error) {
	query := `
		INSERT INTO users (email, phone, password_hash, name, google_id, email_verified)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, email, phone, name, avatar_url, google_id, email_verified, phone_verified, is_active, created_at, updated_at
	`

	row := r.db.QueryRow(ctx, query,
		params.Email,
		params.Phone,
		params.PasswordHash,
		params.Name,
		params.GoogleID,
		params.EmailVerified,
	)

	return scanUser(row)
}

// GetUserByID retrieves a user by ID
func (r *PostgresRepository) GetUserByID(ctx context.Context, id uuid.UUID) (*domain.User, error) {
	query := `
		SELECT id, email, phone, name, avatar_url, google_id, email_verified, phone_verified, is_active, created_at, updated_at
		FROM users WHERE id = $1 AND is_active = TRUE
	`
	row := r.db.QueryRow(ctx, query, id)
	return scanUser(row)
}

// GetUserByEmail retrieves a user by email
func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, phone, name, avatar_url, google_id, email_verified, phone_verified, is_active, created_at, updated_at
		FROM users WHERE email = $1 AND is_active = TRUE
	`
	row := r.db.QueryRow(ctx, query, email)
	return scanUser(row)
}

// GetUserByPhone retrieves a user by phone
func (r *PostgresRepository) GetUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `
		SELECT id, email, phone, name, avatar_url, google_id, email_verified, phone_verified, is_active, created_at, updated_at
		FROM users WHERE phone = $1 AND is_active = TRUE
	`
	row := r.db.QueryRow(ctx, query, phone)
	return scanUser(row)
}

// GetUserByGoogleID retrieves a user by Google ID
func (r *PostgresRepository) GetUserByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	query := `
		SELECT id, email, phone, name, avatar_url, google_id, email_verified, phone_verified, is_active, created_at, updated_at
		FROM users WHERE google_id = $1 AND is_active = TRUE
	`
	row := r.db.QueryRow(ctx, query, googleID)
	return scanUser(row)
}

// GetUserWithPassword retrieves a user with password hash for verification
func (r *PostgresRepository) GetUserWithPassword(ctx context.Context, email string) (*domain.User, string, error) {
	query := `
		SELECT id, email, phone, name, avatar_url, google_id, email_verified, phone_verified, is_active, created_at, updated_at, password_hash
		FROM users WHERE email = $1 AND is_active = TRUE
	`
	row := r.db.QueryRow(ctx, query, email)

	var user domain.User
	var passwordHash *string
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.Name,
		&user.AvatarURL,
		&user.GoogleID,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
		&passwordHash,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", domain.ErrUserNotFound
		}
		return nil, "", err
	}

	hash := ""
	if passwordHash != nil {
		hash = *passwordHash
	}

	return &user, hash, nil
}

// VerifyUserPassword verifies a user's password
func (r *PostgresRepository) VerifyUserPassword(ctx context.Context, email, password string) (*domain.User, error) {
	user, passwordHash, err := r.GetUserWithPassword(ctx, email)
	if err != nil {
		return nil, err
	}

	if passwordHash == "" {
		return nil, domain.ErrInvalidCredentials
	}

	if err := auth.VerifyPassword(password, passwordHash); err != nil {
		return nil, domain.ErrInvalidCredentials
	}

	return user, nil
}

// UpdateUserPassword updates a user's password
func (r *PostgresRepository) UpdateUserPassword(ctx context.Context, userID uuid.UUID, passwordHash string) error {
	query := `UPDATE users SET password_hash = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID, passwordHash)
	return err
}

// LinkGoogleAccount links a Google account to an existing user
func (r *PostgresRepository) LinkGoogleAccount(ctx context.Context, userID uuid.UUID, googleID string) (*domain.User, error) {
	query := `
		UPDATE users SET google_id = $2
		WHERE id = $1
		RETURNING id, email, phone, name, avatar_url, google_id, email_verified, phone_verified, is_active, created_at, updated_at
	`
	row := r.db.QueryRow(ctx, query, userID, googleID)
	return scanUser(row)
}

// UserExistsByEmail checks if a user exists by email
func (r *PostgresRepository) UserExistsByEmail(ctx context.Context, email string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE email = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, email).Scan(&exists)
	return exists, err
}

// UserExistsByPhone checks if a user exists by phone
func (r *PostgresRepository) UserExistsByPhone(ctx context.Context, phone string) (bool, error) {
	query := `SELECT EXISTS(SELECT 1 FROM users WHERE phone = $1)`
	var exists bool
	err := r.db.QueryRow(ctx, query, phone).Scan(&exists)
	return exists, err
}

// CreateSession creates a new session
func (r *PostgresRepository) CreateSession(ctx context.Context, params domain.CreateSessionParams) (*domain.Session, error) {
	query := `
		INSERT INTO sessions (user_id, device_info, ip_address, user_agent, expires_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, user_id, device_info, ip_address, user_agent, is_active, created_at, expires_at, last_activity_at
	`
	row := r.db.QueryRow(ctx, query,
		params.UserID,
		params.DeviceInfo,
		params.IPAddress,
		params.UserAgent,
		params.ExpiresAt,
	)
	return scanSession(row)
}

// GetSessionByID retrieves a session by ID
func (r *PostgresRepository) GetSessionByID(ctx context.Context, id uuid.UUID) (*domain.Session, error) {
	query := `
		SELECT id, user_id, device_info, ip_address, user_agent, is_active, created_at, expires_at, last_activity_at
		FROM sessions WHERE id = $1 AND is_active = TRUE
	`
	row := r.db.QueryRow(ctx, query, id)
	return scanSession(row)
}

// DeactivateSession deactivates a session
func (r *PostgresRepository) DeactivateSession(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE sessions SET is_active = FALSE WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// DeactivateUserSessions deactivates all sessions for a user
func (r *PostgresRepository) DeactivateUserSessions(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE sessions SET is_active = FALSE WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// CreateRefreshToken creates a new refresh token
func (r *PostgresRepository) CreateRefreshToken(ctx context.Context, params domain.CreateRefreshTokenParams) (*domain.RefreshToken, error) {
	query := `
		INSERT INTO refresh_tokens (user_id, session_id, token_hash, expires_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id, user_id, session_id, token_hash, expires_at, revoked, revoked_at, created_at
	`
	row := r.db.QueryRow(ctx, query,
		params.UserID,
		params.SessionID,
		params.TokenHash,
		params.ExpiresAt,
	)
	return scanRefreshToken(row)
}

// GetRefreshTokenByHash retrieves a refresh token by hash
func (r *PostgresRepository) GetRefreshTokenByHash(ctx context.Context, hash string) (*domain.RefreshToken, error) {
	query := `
		SELECT id, user_id, session_id, token_hash, expires_at, revoked, revoked_at, created_at
		FROM refresh_tokens
		WHERE token_hash = $1 AND revoked = FALSE AND expires_at > NOW()
	`
	row := r.db.QueryRow(ctx, query, hash)
	return scanRefreshToken(row)
}

// RevokeRefreshToken revokes a refresh token by ID
func (r *PostgresRepository) RevokeRefreshToken(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE, revoked_at = NOW() WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// RevokeRefreshTokenByHash revokes a refresh token by hash
func (r *PostgresRepository) RevokeRefreshTokenByHash(ctx context.Context, hash string) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE, revoked_at = NOW() WHERE token_hash = $1`
	_, err := r.db.Exec(ctx, query, hash)
	return err
}

// RevokeUserRefreshTokens revokes all refresh tokens for a user
func (r *PostgresRepository) RevokeUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	query := `UPDATE refresh_tokens SET revoked = TRUE, revoked_at = NOW() WHERE user_id = $1`
	_, err := r.db.Exec(ctx, query, userID)
	return err
}

// Helper functions for scanning rows

func scanUser(row pgx.Row) (*domain.User, error) {
	var user domain.User
	err := row.Scan(
		&user.ID,
		&user.Email,
		&user.Phone,
		&user.Name,
		&user.AvatarURL,
		&user.GoogleID,
		&user.EmailVerified,
		&user.PhoneVerified,
		&user.IsActive,
		&user.CreatedAt,
		&user.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return &user, nil
}

func scanSession(row pgx.Row) (*domain.Session, error) {
	var session domain.Session
	err := row.Scan(
		&session.ID,
		&session.UserID,
		&session.DeviceInfo,
		&session.IPAddress,
		&session.UserAgent,
		&session.IsActive,
		&session.CreatedAt,
		&session.ExpiresAt,
		&session.LastActivityAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}
	return &session, nil
}

func scanRefreshToken(row pgx.Row) (*domain.RefreshToken, error) {
	var token domain.RefreshToken
	err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.SessionID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.Revoked,
		&token.RevokedAt,
		&token.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrTokenRevoked
		}
		return nil, err
	}
	return &token, nil
}

// CleanupExpiredTokens removes expired and revoked tokens
func (r *PostgresRepository) CleanupExpiredTokens(ctx context.Context) error {
	queries := []string{
		`DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked = TRUE AND revoked_at < NOW() - INTERVAL '7 days'`,
		`UPDATE sessions SET is_active = FALSE WHERE expires_at < NOW()`,
		`DELETE FROM password_reset_tokens WHERE expires_at < NOW() OR used = TRUE`,
	}

	for _, query := range queries {
		if _, err := r.db.Exec(ctx, query); err != nil {
			return err
		}
	}
	return nil
}

// StartCleanupWorker starts a background worker to clean up expired tokens
func (r *PostgresRepository) StartCleanupWorker(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = r.CleanupExpiredTokens(ctx)
			}
		}
	}()
}

package repository

import (
	"context"
	"encoding/json"
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
		RETURNING id, email, phone, name, avatar_url, bio, gender, date_of_birth, visibility, google_id, email_verified, phone_verified, is_active, created_at, updated_at
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
		SELECT id, email, phone, name, avatar_url, bio, gender, date_of_birth, visibility, google_id, email_verified, phone_verified, is_active, created_at, updated_at
		FROM users WHERE id = $1 AND is_active = TRUE
	`
	row := r.db.QueryRow(ctx, query, id)
	return scanUser(row)
}

// GetUserByEmail retrieves a user by email
func (r *PostgresRepository) GetUserByEmail(ctx context.Context, email string) (*domain.User, error) {
	query := `
		SELECT id, email, phone, name, avatar_url, bio, gender, date_of_birth, visibility, google_id, email_verified, phone_verified, is_active, created_at, updated_at
		FROM users WHERE email = $1 AND is_active = TRUE
	`
	row := r.db.QueryRow(ctx, query, email)
	return scanUser(row)
}

// GetUserByPhone retrieves a user by phone
func (r *PostgresRepository) GetUserByPhone(ctx context.Context, phone string) (*domain.User, error) {
	query := `
		SELECT id, email, phone, name, avatar_url, bio, gender, date_of_birth, visibility, google_id, email_verified, phone_verified, is_active, created_at, updated_at
		FROM users WHERE phone = $1 AND is_active = TRUE
	`
	row := r.db.QueryRow(ctx, query, phone)
	return scanUser(row)
}

// GetUserByGoogleID retrieves a user by Google ID
func (r *PostgresRepository) GetUserByGoogleID(ctx context.Context, googleID string) (*domain.User, error) {
	query := `
		SELECT id, email, phone, name, avatar_url, bio, gender, date_of_birth, visibility, google_id, email_verified, phone_verified, is_active, created_at, updated_at
		FROM users WHERE google_id = $1 AND is_active = TRUE
	`
	row := r.db.QueryRow(ctx, query, googleID)
	return scanUser(row)
}

// GetUserWithPassword retrieves a user with password hash for verification
func (r *PostgresRepository) GetUserWithPassword(ctx context.Context, email string) (*domain.User, string, error) {
	query := `
		SELECT id, email, phone, name, avatar_url, bio, gender, date_of_birth, visibility, google_id, email_verified, phone_verified, is_active, created_at, updated_at, password_hash
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
		&user.Bio,
		&user.Gender,
		&user.DateOfBirth,
		&user.Visibility,
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
		RETURNING id, email, phone, name, avatar_url, bio, gender, date_of_birth, visibility, google_id, email_verified, phone_verified, is_active, created_at, updated_at
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

// UpdateUser updates a user profile
func (r *PostgresRepository) UpdateUser(ctx context.Context, userID uuid.UUID, params domain.UpdateUserParams) (*domain.User, error) {
	query := `
		UPDATE users 
		SET name = COALESCE($2, name),
			bio = COALESCE($3, bio),
			gender = COALESCE($4, gender),
			date_of_birth = COALESCE($5, date_of_birth),
			visibility = COALESCE($6, visibility),
			avatar_url = COALESCE($7, avatar_url)
		WHERE id = $1
		RETURNING id, email, phone, name, avatar_url, bio, gender, date_of_birth, visibility, google_id, email_verified, phone_verified, is_active, created_at, updated_at
	`
	row := r.db.QueryRow(ctx, query,
		userID,
		params.Name,
		params.Bio,
		params.Gender,
		params.DateOfBirth,
		params.Visibility,
		params.AvatarURL,
	)
	return scanUser(row)
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
		&user.Bio,
		&user.Gender,
		&user.DateOfBirth,
		&user.Visibility,
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

// CreatePasswordResetToken creates a new password reset token
func (r *PostgresRepository) CreatePasswordResetToken(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	query := `
		INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`
	_, err := r.db.Exec(ctx, query, userID, tokenHash, expiresAt)
	return err
}

// GetPasswordResetToken retrieves a password reset token by hash
func (r *PostgresRepository) GetPasswordResetToken(ctx context.Context, tokenHash string) (*domain.PasswordResetToken, error) {
	query := `
		SELECT id, user_id, token_hash, expires_at, used, created_at
		FROM password_reset_tokens
		WHERE token_hash = $1
	`
	row := r.db.QueryRow(ctx, query, tokenHash)

	var token domain.PasswordResetToken
	err := row.Scan(
		&token.ID,
		&token.UserID,
		&token.TokenHash,
		&token.ExpiresAt,
		&token.Used,
		&token.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, domain.ErrInvalidToken
		}
		return nil, err
	}
	return &token, nil
}

// MarkPasswordResetTokenUsed marks a password reset token as used
func (r *PostgresRepository) MarkPasswordResetTokenUsed(ctx context.Context, id uuid.UUID) error {
	query := `UPDATE password_reset_tokens SET used = TRUE WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

// UpdateUserEmail updates a user's email
func (r *PostgresRepository) UpdateUserEmail(ctx context.Context, userID uuid.UUID, email string) error {
	query := `UPDATE users SET email = $2, email_verified = FALSE WHERE id = $1`
	_, err := r.db.Exec(ctx, query, userID, email)
	return err
}

// UpdateSessionFCMToken updates a session's FCM token
func (r *PostgresRepository) UpdateSessionFCMToken(ctx context.Context, sessionID uuid.UUID, fcmToken string) error {
	query := `UPDATE sessions SET fcm_token = $2 WHERE id = $1`
	_, err := r.db.Exec(ctx, query, sessionID, fcmToken)
	return err
}

// Helper to scan story with user
func scanStoryWithUser(row pgx.Row) (*domain.Story, error) {
	var s domain.Story
	var u domain.User
	err := row.Scan(
		&s.ID, &s.UserID, &s.MediaURL, &s.MediaType, &s.Caption, &s.LocationLat, &s.LocationLng, &s.ExpiresAt, &s.CreatedAt,
		&u.ID, &u.Email, &u.Phone, &u.Name, &u.AvatarURL, &u.Bio, &u.Gender, &u.DateOfBirth, &u.Visibility, &u.GoogleID, &u.EmailVerified, &u.PhoneVerified, &u.IsActive, &u.CreatedAt, &u.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	s.User = u.ToResponse()
	return &s, nil
}

func (r *PostgresRepository) CreateStory(ctx context.Context, params domain.CreateStoryParams) (*domain.Story, error) {
	query := `
		WITH inserted_story AS (
			INSERT INTO stories (user_id, media_url, media_type, caption, location_lat, location_lng, expires_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
			RETURNING id, user_id, media_url, media_type, caption, location_lat, location_lng, expires_at, created_at
		)
		SELECT s.id, s.user_id, s.media_url, s.media_type, s.caption, s.location_lat, s.location_lng, s.expires_at, s.created_at,
		       u.id, u.email, u.phone, u.name, u.avatar_url, u.bio, u.gender, u.date_of_birth, u.visibility, u.google_id, u.email_verified, u.phone_verified, u.is_active, u.created_at, u.updated_at
		FROM inserted_story s
		JOIN users u ON s.user_id = u.id
	`
	row := r.db.QueryRow(ctx, query,
		params.UserID,
		params.MediaURL,
		params.MediaType,
		params.Caption,
		params.LocationLat,
		params.LocationLng,
		params.ExpiresAt,
	)
	return scanStoryWithUser(row)
}

func (r *PostgresRepository) GetActiveStories(ctx context.Context, limit, offset int) ([]*domain.Story, error) {
	query := `
		SELECT s.id, s.user_id, s.media_url, s.media_type, s.caption, s.location_lat, s.location_lng, s.expires_at, s.created_at,
		       u.id, u.email, u.phone, u.name, u.avatar_url, u.bio, u.gender, u.date_of_birth, u.visibility, u.google_id, u.email_verified, u.phone_verified, u.is_active, u.created_at, u.updated_at
		FROM stories s
		JOIN users u ON s.user_id = u.id
		WHERE s.expires_at > NOW()
		ORDER BY s.created_at DESC
		LIMIT $1 OFFSET $2
	`
	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stories []*domain.Story
	for rows.Next() {
		story, err := scanStoryWithUser(rows)
		if err != nil {
			return nil, err
		}
		stories = append(stories, story)
	}
	return stories, nil
}

func (r *PostgresRepository) GetStoriesByLocation(ctx context.Context, lat, lng, radius float64, limit, offset int) ([]*domain.Story, error) {
	// Radius logic: we use earth_distance extension if available.
	// Since migration 004 adds it, we use it.
	// earth_box(ll_to_earth(lat, lng), radius) creates a bounding box.
	// radius is in meters.
	query := `
		SELECT s.id, s.user_id, s.media_url, s.media_type, s.caption, s.location_lat, s.location_lng, s.expires_at, s.created_at,
		       u.id, u.email, u.phone, u.name, u.avatar_url, u.bio, u.gender, u.date_of_birth, u.visibility, u.google_id, u.email_verified, u.phone_verified, u.is_active, u.created_at, u.updated_at
		FROM stories s
		JOIN users u ON s.user_id = u.id
		WHERE s.expires_at > NOW()
		AND s.location_lat IS NOT NULL AND s.location_lng IS NOT NULL
		AND earth_box(ll_to_earth($1, $2), $3) @> ll_to_earth(s.location_lat, s.location_lng)
		AND earth_distance(ll_to_earth($1, $2), ll_to_earth(s.location_lat, s.location_lng)) < $3
		ORDER BY s.created_at DESC
		LIMIT $4 OFFSET $5
	`
	rows, err := r.db.Query(ctx, query, lat, lng, radius, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var stories []*domain.Story
	for rows.Next() {
		story, err := scanStoryWithUser(rows)
		if err != nil {
			return nil, err
		}
		stories = append(stories, story)
	}
	return stories, nil
}

func (r *PostgresRepository) DeleteExpiredStories(ctx context.Context) (int64, error) {
	query := `DELETE FROM stories WHERE expires_at < NOW()`
	tag, err := r.db.Exec(ctx, query)
	if err != nil {
		return 0, err
	}
	return tag.RowsAffected(), nil
}

// Chat methods

func (r *PostgresRepository) CreateChat(ctx context.Context, user1ID, user2ID uuid.UUID) (*domain.Chat, error) {
	// Check if chat exists
	// This query finds a chat where both users are participants and there are exactly 2 participants
	queryCheck := `
		SELECT cp1.chat_id
		FROM chat_participants cp1
		JOIN chat_participants cp2 ON cp1.chat_id = cp2.chat_id
		WHERE cp1.user_id = $1 AND cp2.user_id = $2
		GROUP BY cp1.chat_id
	`
	var existingChatID uuid.UUID
	err := r.db.QueryRow(ctx, queryCheck, user1ID, user2ID).Scan(&existingChatID)
	if err == nil {
		return r.GetChatByID(ctx, existingChatID)
	}

	// Create new chat
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var chatID uuid.UUID
	var createdAt, updatedAt time.Time
	err = tx.QueryRow(ctx, "INSERT INTO chats DEFAULT VALUES RETURNING id, created_at, updated_at").Scan(&chatID, &createdAt, &updatedAt)
	if err != nil {
		return nil, err
	}

	// Add participants
	_, err = tx.Exec(ctx, "INSERT INTO chat_participants (chat_id, user_id) VALUES ($1, $2), ($1, $3)", chatID, user1ID, user2ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return r.GetChatByID(ctx, chatID)
}

func (r *PostgresRepository) GetChatByID(ctx context.Context, chatID uuid.UUID) (*domain.Chat, error) {
	queryChat := `SELECT id, created_at, updated_at FROM chats WHERE id = $1`
	var chat domain.Chat
	err := r.db.QueryRow(ctx, queryChat, chatID).Scan(&chat.ID, &chat.CreatedAt, &chat.UpdatedAt)
	if err != nil {
		return nil, err
	}

	// Get participants
	queryParticipants := `
		SELECT u.id, u.email, u.phone, u.name, u.avatar_url
		FROM chat_participants cp
		JOIN users u ON cp.user_id = u.id
		WHERE cp.chat_id = $1
	`
	rows, err := r.db.Query(ctx, queryParticipants, chatID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var u domain.UserResponse
		if err := rows.Scan(&u.ID, &u.Email, &u.Phone, &u.Name, &u.AvatarURL); err != nil {
			return nil, err
		}
		chat.Users = append(chat.Users, &u)
	}

	return &chat, nil
}

func (r *PostgresRepository) GetChatsByUserID(ctx context.Context, userID uuid.UUID) ([]*domain.Chat, error) {
	query := `
		SELECT c.id, c.created_at, c.updated_at
		FROM chats c
		JOIN chat_participants cp ON c.id = cp.chat_id
		WHERE cp.user_id = $1
		ORDER BY c.updated_at DESC
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var chats []*domain.Chat
	for rows.Next() {
		var chat domain.Chat
		if err := rows.Scan(&chat.ID, &chat.CreatedAt, &chat.UpdatedAt); err != nil {
			return nil, err
		}
		chats = append(chats, &chat)
	}

	// For each chat, get participants (Optimization: could use array_agg but this is simpler for now)
	for _, chat := range chats {
		// Re-use logic or fetch query
		queryParticipants := `
			SELECT u.id, u.email, u.phone, u.name, u.avatar_url
			FROM chat_participants cp
			JOIN users u ON cp.user_id = u.id
			WHERE cp.chat_id = $1
		`
		pRows, err := r.db.Query(ctx, queryParticipants, chat.ID)
		if err != nil {
			continue // skip error for fetch list
		}
		for pRows.Next() {
			var u domain.UserResponse
			_ = pRows.Scan(&u.ID, &u.Email, &u.Phone, &u.Name, &u.AvatarURL)
			chat.Users = append(chat.Users, &u)
		}
		pRows.Close()

		// Get last message
		queryMsg := `SELECT id, chat_id, sender_id, content, read_at, created_at FROM messages WHERE chat_id = $1 ORDER BY created_at DESC LIMIT 1`
		var msg domain.Message
		if err := r.db.QueryRow(ctx, queryMsg, chat.ID).Scan(&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Content, &msg.ReadAt, &msg.CreatedAt); err == nil {
			chat.LastMessage = &msg
		}
	}

	return chats, nil
}

func (r *PostgresRepository) CreateMessage(ctx context.Context, chatID, senderID uuid.UUID, content string) (*domain.Message, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	query := `
		INSERT INTO messages (chat_id, sender_id, content)
		VALUES ($1, $2, $3)
		RETURNING id, created_at
	`
	var msg domain.Message
	msg.ChatID = chatID
	msg.SenderID = senderID
	msg.Content = content

	err = tx.QueryRow(ctx, query, chatID, senderID, content).Scan(&msg.ID, &msg.CreatedAt)
	if err != nil {
		return nil, err
	}

	// Update chat updated_at
	_, err = tx.Exec(ctx, "UPDATE chats SET updated_at = NOW() WHERE id = $1", chatID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &msg, nil
}

func (r *PostgresRepository) GetMessages(ctx context.Context, chatID uuid.UUID, limit, offset int) ([]*domain.Message, error) {
	query := `
		SELECT id, chat_id, sender_id, content, read_at, created_at
		FROM messages
		WHERE chat_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, chatID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var messages []*domain.Message
	for rows.Next() {
		var msg domain.Message
		if err := rows.Scan(&msg.ID, &msg.ChatID, &msg.SenderID, &msg.Content, &msg.ReadAt, &msg.CreatedAt); err != nil {
			return nil, err
		}
		messages = append(messages, &msg)
	}
	return messages, nil
}

// Connection methods

func (r *PostgresRepository) CreateConnectionRequest(ctx context.Context, requesterID, receiverID uuid.UUID) (*domain.Connection, error) {
	// Check if reverse connection exists
	queryCheck := `SELECT id, status FROM connections WHERE requester_id = $1 AND receiver_id = $2`
	var existingID uuid.UUID
	var status domain.ConnectionStatus
	err := r.db.QueryRow(ctx, queryCheck, receiverID, requesterID).Scan(&existingID, &status)
	if err == nil {
		// If reverse exists and is pending, we could auto-accept.
		// For now simple implementation: just error or let unique constraint fail if direct dupe.
		// If explicit logic needed:
		if status == domain.ConnectionStatusPending {
			// Auto accept logic could go here, but let's stick to standard flow:
			// User B requested User A. User A requesting User B should probably just accept User B's request.
			// Implementing auto-accept:
			return r.UpdateConnectionStatus(ctx, existingID, domain.ConnectionStatusAccepted)
		}
	}

	query := `
		INSERT INTO connections (requester_id, receiver_id, status)
		VALUES ($1, $2, 'pending')
		ON CONFLICT (requester_id, receiver_id) DO UPDATE SET updated_at = NOW() -- prevent duplicate error, maybe return existing
		RETURNING id, requester_id, receiver_id, status, created_at, updated_at
	`
	// Note: On conflict we might want to check status. If rejected, maybe allow re-request?
	// For MVP, just return the inserted/updated row.

	var conn domain.Connection
	err = r.db.QueryRow(ctx, query, requesterID, receiverID).Scan(
		&conn.ID, &conn.RequesterID, &conn.ReceiverID, &conn.Status, &conn.CreatedAt, &conn.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func (r *PostgresRepository) UpdateConnectionStatus(ctx context.Context, connectionID uuid.UUID, status domain.ConnectionStatus) (*domain.Connection, error) {
	query := `
		UPDATE connections
		SET status = $2, updated_at = NOW()
		WHERE id = $1
		RETURNING id, requester_id, receiver_id, status, created_at, updated_at
	`
	var conn domain.Connection
	err := r.db.QueryRow(ctx, query, connectionID, status).Scan(
		&conn.ID, &conn.RequesterID, &conn.ReceiverID, &conn.Status, &conn.CreatedAt, &conn.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func (r *PostgresRepository) GetConnectionByID(ctx context.Context, connectionID uuid.UUID) (*domain.Connection, error) {
	query := `SELECT id, requester_id, receiver_id, status, created_at, updated_at FROM connections WHERE id = $1`
	var conn domain.Connection
	err := r.db.QueryRow(ctx, query, connectionID).Scan(
		&conn.ID, &conn.RequesterID, &conn.ReceiverID, &conn.Status, &conn.CreatedAt, &conn.UpdatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &conn, nil
}

func (r *PostgresRepository) GetConnections(ctx context.Context, userID uuid.UUID, status domain.ConnectionStatus, limit, offset int) ([]*domain.Connection, error) {
	// If status is accepted, we want connections where user is EITHER requester OR receiver
	// If status is pending, usually we want requests RECEIVED by user (to accept/reject)
	// or requests SENT by user (to see what they sent).
	// Let's implement generic filter.

	var query string
	var rows pgx.Rows
	var err error

	switch status {
	case domain.ConnectionStatusAccepted:
		query = `
			SELECT c.id, c.requester_id, c.receiver_id, c.status, c.created_at, c.updated_at,
			       u.id, u.email, u.phone, u.name, u.avatar_url
			FROM connections c
			JOIN users u ON (CASE WHEN c.requester_id = $1 THEN c.receiver_id ELSE c.requester_id END) = u.id
			WHERE (c.requester_id = $1 OR c.receiver_id = $1)
			AND c.status = 'accepted'
			ORDER BY c.updated_at DESC
			LIMIT $2 OFFSET $3
		`
		rows, err = r.db.Query(ctx, query, userID, limit, offset)
	case domain.ConnectionStatusPending:
		// Default to requests RECEIVED by user (to accept)
		query = `
			SELECT c.id, c.requester_id, c.receiver_id, c.status, c.created_at, c.updated_at,
			       u.id, u.email, u.phone, u.name, u.avatar_url
			FROM connections c
			JOIN users u ON c.requester_id = u.id
			WHERE c.receiver_id = $1
			AND c.status = 'pending'
			ORDER BY c.created_at DESC
			LIMIT $2 OFFSET $3
		`
		rows, err = r.db.Query(ctx, query, userID, limit, offset)
	default:
		return nil, errors.New("unsupported status filter")
	}

	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var connections []*domain.Connection
	for rows.Next() {
		var conn domain.Connection
		var u domain.UserResponse
		// We join to get the "other" user details
		err := rows.Scan(
			&conn.ID, &conn.RequesterID, &conn.ReceiverID, &conn.Status, &conn.CreatedAt, &conn.UpdatedAt,
			&u.ID, &u.Email, &u.Phone, &u.Name, &u.AvatarURL,
		)
		if err != nil {
			return nil, err
		}
		conn.User = &u
		connections = append(connections, &conn)
	}
	return connections, nil
}

func (r *PostgresRepository) DeleteConnection(ctx context.Context, connectionID uuid.UUID) error {
	_, err := r.db.Exec(ctx, "DELETE FROM connections WHERE id = $1", connectionID)
	return err
}

// Notification methods

func (r *PostgresRepository) CreateNotification(ctx context.Context, userID uuid.UUID, typeStr, title, body string, data map[string]interface{}) error {
	dataJSON, err := json.Marshal(data)
	if err != nil {
		return err
	}

	query := `
		INSERT INTO notifications (user_id, type, title, body, data)
		VALUES ($1, $2, $3, $4, $5)
	`
	_, err = r.db.Exec(ctx, query, userID, typeStr, title, body, dataJSON)
	return err
}

func (r *PostgresRepository) GetNotifications(ctx context.Context, userID uuid.UUID, limit, offset int) ([]*domain.Notification, error) {
	query := `
		SELECT id, user_id, type, title, body, data, is_read, created_at
		FROM notifications
		WHERE user_id = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	rows, err := r.db.Query(ctx, query, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var notifications []*domain.Notification
	for rows.Next() {
		var n domain.Notification
		var dataJSON []byte
		if err := rows.Scan(&n.ID, &n.UserID, &n.Type, &n.Title, &n.Body, &dataJSON, &n.IsRead, &n.CreatedAt); err != nil {
			return nil, err
		}
		if len(dataJSON) > 0 {
			_ = json.Unmarshal(dataJSON, &n.Data)
		}
		notifications = append(notifications, &n)
	}
	return notifications, nil
}

func (r *PostgresRepository) MarkNotificationRead(ctx context.Context, notificationID uuid.UUID) error {
	query := `UPDATE notifications SET is_read = TRUE WHERE id = $1`
	_, err := r.db.Exec(ctx, query, notificationID)
	return err
}

func (r *PostgresRepository) GetFCMTokens(ctx context.Context, userID uuid.UUID) ([]string, error) {
	query := `
		SELECT DISTINCT fcm_token
		FROM sessions
		WHERE user_id = $1 AND is_active = TRUE AND fcm_token IS NOT NULL AND fcm_token != ''
	`
	rows, err := r.db.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tokens []string
	for rows.Next() {
		var token string
		if err := rows.Scan(&token); err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	}
	return tokens, nil
}

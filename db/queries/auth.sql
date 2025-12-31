-- name: CreateSession :one
INSERT INTO sessions (
    user_id, device_info, ip_address, user_agent, expires_at
) VALUES (
    $1, $2, $3, $4, $5
) RETURNING *;

-- name: GetSessionByID :one
SELECT * FROM sessions WHERE id = $1 AND is_active = TRUE;

-- name: GetUserSessions :many
SELECT * FROM sessions 
WHERE user_id = $1 AND is_active = TRUE 
ORDER BY last_activity_at DESC;

-- name: UpdateSessionActivity :exec
UPDATE sessions SET last_activity_at = NOW() WHERE id = $1;

-- name: DeactivateSession :exec
UPDATE sessions SET is_active = FALSE WHERE id = $1;

-- name: DeactivateUserSessions :exec
UPDATE sessions SET is_active = FALSE WHERE user_id = $1;

-- name: CleanupExpiredSessions :exec
UPDATE sessions SET is_active = FALSE WHERE expires_at < NOW();

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (
    user_id, session_id, token_hash, expires_at
) VALUES (
    $1, $2, $3, $4
) RETURNING *;

-- name: GetRefreshTokenByHash :one
SELECT * FROM refresh_tokens 
WHERE token_hash = $1 AND revoked = FALSE AND expires_at > NOW();

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked = TRUE, revoked_at = NOW() WHERE id = $1;

-- name: RevokeRefreshTokenByHash :exec
UPDATE refresh_tokens SET revoked = TRUE, revoked_at = NOW() WHERE token_hash = $1;

-- name: RevokeUserRefreshTokens :exec
UPDATE refresh_tokens SET revoked = TRUE, revoked_at = NOW() WHERE user_id = $1;

-- name: CleanupExpiredTokens :exec
DELETE FROM refresh_tokens WHERE expires_at < NOW() OR revoked = TRUE;

-- name: CreatePasswordResetToken :one
INSERT INTO password_reset_tokens (user_id, token_hash, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetPasswordResetToken :one
SELECT * FROM password_reset_tokens
WHERE token_hash = $1 AND used = FALSE AND expires_at > NOW();

-- name: MarkPasswordResetTokenUsed :exec
UPDATE password_reset_tokens SET used = TRUE WHERE id = $1;

-- name: CleanupExpiredPasswordResetTokens :exec
DELETE FROM password_reset_tokens WHERE expires_at < NOW() OR used = TRUE;

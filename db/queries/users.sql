-- name: CreateUser :one
INSERT INTO users (
    email, phone, password_hash, name, google_id, email_verified
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: GetUserByID :one
SELECT * FROM users WHERE id = $1 AND is_active = TRUE;

-- name: GetUserByEmail :one
SELECT * FROM users WHERE email = $1 AND is_active = TRUE;

-- name: GetUserByPhone :one
SELECT * FROM users WHERE phone = $1 AND is_active = TRUE;

-- name: GetUserByGoogleID :one
SELECT * FROM users WHERE google_id = $1 AND is_active = TRUE;

-- name: UpdateUser :one
UPDATE users SET
    name = COALESCE($2, name),
    avatar_url = COALESCE($3, avatar_url),
    email_verified = COALESCE($4, email_verified),
    phone_verified = COALESCE($5, phone_verified)
WHERE id = $1
RETURNING *;

-- name: UpdateUserPassword :exec
UPDATE users SET password_hash = $2 WHERE id = $1;

-- name: LinkGoogleAccount :one
UPDATE users SET google_id = $2 WHERE id = $1 RETURNING *;

-- name: DeactivateUser :exec
UPDATE users SET is_active = FALSE WHERE id = $1;

-- name: UserExistsByEmail :one
SELECT EXISTS(SELECT 1 FROM users WHERE email = $1);

-- name: UserExistsByPhone :one
SELECT EXISTS(SELECT 1 FROM users WHERE phone = $1);

-- Rollback migration 001

DROP TRIGGER IF EXISTS update_users_updated_at ON users;
DROP FUNCTION IF EXISTS update_updated_at_column();

DROP INDEX IF EXISTS idx_refresh_tokens_expires_at;
DROP INDEX IF EXISTS idx_refresh_tokens_token_hash;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP INDEX IF EXISTS idx_sessions_expires_at;
DROP INDEX IF EXISTS idx_sessions_user_id;
DROP INDEX IF EXISTS idx_users_google_id;
DROP INDEX IF EXISTS idx_users_phone;
DROP INDEX IF EXISTS idx_users_email;

DROP TABLE IF EXISTS password_reset_tokens;
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS sessions;
DROP TABLE IF EXISTS users;

DROP EXTENSION IF EXISTS "uuid-ossp";

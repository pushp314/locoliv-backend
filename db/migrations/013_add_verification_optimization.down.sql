-- Rollback migration 013

-- Drop verification indexes
DROP INDEX IF EXISTS idx_users_verified;
DROP INDEX IF EXISTS idx_verification_requests_status;
DROP INDEX IF EXISTS idx_verification_requests_user;

-- Drop optimization indexes
DROP INDEX IF EXISTS idx_stories_feed;
DROP INDEX IF EXISTS idx_stories_location;
DROP INDEX IF EXISTS idx_stories_expires;
DROP INDEX IF EXISTS idx_messages_chat_time;
DROP INDEX IF EXISTS idx_chats_user_updated;
DROP INDEX IF EXISTS idx_connections_user_status;
DROP INDEX IF EXISTS idx_connections_recipient_status;
DROP INDEX IF EXISTS idx_connections_accepted;
DROP INDEX IF EXISTS idx_karma_user_created;
DROP INDEX IF EXISTS idx_user_karma_tier;
DROP INDEX IF EXISTS idx_profiles_city_karma;
DROP INDEX IF EXISTS idx_user_karma_leaderboard;
DROP INDEX IF EXISTS idx_events_city_date;
DROP INDEX IF EXISTS idx_events_creator;
DROP INDEX IF EXISTS idx_neighborhood_city_created;
DROP INDEX IF EXISTS idx_neighborhood_resolved;
DROP INDEX IF EXISTS idx_matches_user_date;
DROP INDEX IF EXISTS idx_matches_pending;
DROP INDEX IF EXISTS idx_notifications_user_read;
DROP INDEX IF EXISTS idx_subscriptions_active;
DROP INDEX IF EXISTS idx_sessions_user_active;

-- Drop verification table
DROP TABLE IF EXISTS verification_requests;

-- Remove verification columns from users
ALTER TABLE users DROP COLUMN IF EXISTS is_verified;
ALTER TABLE users DROP COLUMN IF EXISTS verification_type;
ALTER TABLE users DROP COLUMN IF EXISTS verified_at;
ALTER TABLE users DROP COLUMN IF EXISTS verified_by;

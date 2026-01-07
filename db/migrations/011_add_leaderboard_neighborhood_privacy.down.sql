-- Rollback Phase 14
DROP TABLE IF EXISTS neighborhood_upvotes;
DROP TABLE IF EXISTS neighborhood_answers;
DROP TABLE IF EXISTS neighborhood_posts;
DROP TABLE IF EXISTS blocked_users;
DROP TABLE IF EXISTS close_friends;
DROP TABLE IF EXISTS user_privacy_settings;

ALTER TABLE stories DROP COLUMN IF EXISTS visibility;
ALTER TABLE profiles DROP COLUMN IF EXISTS city;
ALTER TABLE profiles DROP COLUMN IF EXISTS state;

-- Profile Verification Badges & Database Optimization
-- Migration 013: Add verification badges and optimize indexes

-- Add verification fields to users
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_verified BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS verification_type VARCHAR(20); -- 'blue', 'celebrity', 'official', 'business'
ALTER TABLE users ADD COLUMN IF NOT EXISTS verified_at TIMESTAMP;
ALTER TABLE users ADD COLUMN IF NOT EXISTS verified_by UUID REFERENCES users(id);

-- Create verification requests table
CREATE TABLE IF NOT EXISTS verification_requests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    request_type VARCHAR(20) NOT NULL, -- 'identity', 'celebrity', 'official', 'business'
    document_url TEXT,
    document_type VARCHAR(50), -- 'aadhaar', 'pan', 'passport', 'govt_id', 'business_reg'
    status VARCHAR(20) DEFAULT 'pending', -- 'pending', 'approved', 'rejected'
    rejection_reason TEXT,
    reviewed_by UUID REFERENCES users(id),
    reviewed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes for verification
CREATE INDEX IF NOT EXISTS idx_users_verified ON users(is_verified) WHERE is_verified = TRUE;
CREATE INDEX IF NOT EXISTS idx_verification_requests_status ON verification_requests(status, created_at);
CREATE INDEX IF NOT EXISTS idx_verification_requests_user ON verification_requests(user_id);

-- Database Query Optimization: Add missing indexes

-- Story feed optimization
CREATE INDEX IF NOT EXISTS idx_stories_feed ON stories(created_at DESC, user_id) WHERE expires_at > NOW();
CREATE INDEX IF NOT EXISTS idx_stories_location ON stories(location_lat, location_lng) WHERE expires_at > NOW();
CREATE INDEX IF NOT EXISTS idx_stories_expires ON stories(expires_at) WHERE expires_at > NOW();

-- Chat optimization
CREATE INDEX IF NOT EXISTS idx_messages_chat_time ON messages(chat_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_chats_user_updated ON chats(updated_at DESC);

-- Connection optimization
CREATE INDEX IF NOT EXISTS idx_connections_user_status ON connections(requester_id, status);
CREATE INDEX IF NOT EXISTS idx_connections_recipient_status ON connections(recipient_id, status);
CREATE INDEX IF NOT EXISTS idx_connections_accepted ON connections(requester_id, recipient_id) WHERE status = 'accepted';

-- Karma optimization
CREATE INDEX IF NOT EXISTS idx_karma_user_created ON karma_transactions(user_id, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_user_karma_tier ON user_karma(current_tier, total_karma DESC);

-- Leaderboard optimization
CREATE INDEX IF NOT EXISTS idx_profiles_city_karma ON profiles(city) INCLUDE (user_id);
CREATE INDEX IF NOT EXISTS idx_user_karma_leaderboard ON user_karma(total_karma DESC) INCLUDE (user_id, current_tier);

-- Events optimization
CREATE INDEX IF NOT EXISTS idx_events_city_date ON local_events(city, event_date) WHERE event_date >= NOW();
CREATE INDEX IF NOT EXISTS idx_events_creator ON local_events(creator_id, event_date DESC);

-- Neighborhood posts optimization
CREATE INDEX IF NOT EXISTS idx_neighborhood_city_created ON neighborhood_posts(city, created_at DESC);
CREATE INDEX IF NOT EXISTS idx_neighborhood_resolved ON neighborhood_posts(is_resolved, created_at DESC);

-- Daily matches optimization
CREATE INDEX IF NOT EXISTS idx_matches_user_date ON daily_matches(user_id, match_date);
CREATE INDEX IF NOT EXISTS idx_matches_pending ON daily_matches(user_id) WHERE is_accepted IS NULL;

-- Notifications optimization
CREATE INDEX IF NOT EXISTS idx_notifications_user_read ON notifications(user_id, is_read, created_at DESC);

-- Subscription optimization
CREATE INDEX IF NOT EXISTS idx_subscriptions_active ON subscriptions(user_id, status, expires_at) WHERE status = 'active';

-- Session optimization
CREATE INDEX IF NOT EXISTS idx_sessions_user_active ON sessions(user_id) WHERE expires_at > NOW();

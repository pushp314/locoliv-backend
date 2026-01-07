-- Phase 14: Leaderboard, Ask Neighborhood & Privacy

-- Add city/location to user profiles for leaderboard
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS city VARCHAR(100);
ALTER TABLE profiles ADD COLUMN IF NOT EXISTS state VARCHAR(100);

-- Create index for leaderboard queries
CREATE INDEX IF NOT EXISTS idx_profiles_city ON profiles(city);

-- Privacy settings for users
CREATE TABLE user_privacy_settings (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    default_story_visibility VARCHAR(20) DEFAULT 'public', -- public, connections, family
    show_in_discovery BOOLEAN DEFAULT TRUE,
    show_location BOOLEAN DEFAULT TRUE,
    who_can_message VARCHAR(20) DEFAULT 'everyone', -- everyone, connections, nobody
    incognito_mode BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Story visibility enhancement
ALTER TABLE stories ADD COLUMN IF NOT EXISTS visibility VARCHAR(20) DEFAULT 'public';

-- Close friends/family list
CREATE TABLE close_friends (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    friend_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    list_type VARCHAR(20) DEFAULT 'family', -- family, close_friends
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, friend_id)
);

-- Block list
CREATE TABLE blocked_users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    blocked_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    reason TEXT,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, blocked_user_id)
);

-- Ask Neighborhood posts
CREATE TABLE neighborhood_posts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_type VARCHAR(20) NOT NULL DEFAULT 'question', -- question, recommendation, help, event
    title VARCHAR(200) NOT NULL,
    content TEXT,
    location_lat DECIMAL(10,8),
    location_lng DECIMAL(11,8),
    city VARCHAR(100),
    radius_km INT DEFAULT 5,
    is_resolved BOOLEAN DEFAULT FALSE,
    upvotes INT DEFAULT 0,
    answer_count INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Answers to neighborhood posts
CREATE TABLE neighborhood_answers (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    post_id UUID NOT NULL REFERENCES neighborhood_posts(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    is_accepted BOOLEAN DEFAULT FALSE,
    upvotes INT DEFAULT 0,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Upvotes for posts and answers
CREATE TABLE neighborhood_upvotes (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    post_id UUID REFERENCES neighborhood_posts(id) ON DELETE CASCADE,
    answer_id UUID REFERENCES neighborhood_answers(id) ON DELETE CASCADE,
    created_at TIMESTAMP DEFAULT NOW(),
    CONSTRAINT upvote_target CHECK (
        (post_id IS NOT NULL AND answer_id IS NULL) OR
        (post_id IS NULL AND answer_id IS NOT NULL)
    )
);

-- Indexes
CREATE INDEX idx_neighborhood_posts_location ON neighborhood_posts(city, created_at DESC);
CREATE INDEX idx_neighborhood_posts_type ON neighborhood_posts(post_type, is_resolved);
CREATE INDEX idx_neighborhood_answers_post ON neighborhood_answers(post_id);
CREATE INDEX idx_blocked_users_user ON blocked_users(user_id);
CREATE INDEX idx_close_friends_user ON close_friends(user_id);

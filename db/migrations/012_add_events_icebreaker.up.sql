-- Phase 15: Events and Daily Icebreaker Match

-- Local Events
CREATE TABLE local_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    creator_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    title_hindi VARCHAR(200),
    description TEXT,
    event_type VARCHAR(50) NOT NULL DEFAULT 'general', -- general, sports, religious, social, party
    location_lat DECIMAL(10,8),
    location_lng DECIMAL(11,8),
    venue VARCHAR(200),
    city VARCHAR(100),
    event_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP,
    max_attendees INT,
    is_public BOOLEAN DEFAULT TRUE,
    cover_image_url TEXT,
    karma_reward INT DEFAULT 10,
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Event RSVPs
CREATE TABLE event_rsvps (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    event_id UUID NOT NULL REFERENCES local_events(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    status VARCHAR(20) DEFAULT 'going', -- going, interested, not_going
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(event_id, user_id)
);

-- Daily Icebreaker Match
CREATE TABLE daily_matches (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    matched_user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    match_date DATE DEFAULT CURRENT_DATE,
    match_reason TEXT, -- Why they were matched (similar interests, location, etc.)
    is_accepted BOOLEAN,
    accepted_at TIMESTAMP,
    karma_awarded BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, match_date)
);

-- User interests for matching
CREATE TABLE user_interests (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    interest VARCHAR(100) NOT NULL,
    category VARCHAR(50), -- sports, music, food, travel, etc.
    created_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, interest)
);

-- Indexes
CREATE INDEX idx_local_events_city ON local_events(city, event_date);
CREATE INDEX idx_local_events_date ON local_events(event_date);
CREATE INDEX idx_event_rsvps_event ON event_rsvps(event_id);
CREATE INDEX idx_event_rsvps_user ON event_rsvps(user_id);
CREATE INDEX idx_daily_matches_user ON daily_matches(user_id, match_date);
CREATE INDEX idx_user_interests_user ON user_interests(user_id);

-- Seed some interest categories
INSERT INTO user_interests (user_id, interest, category)
SELECT gen_random_uuid(), interest, category
FROM (VALUES 
    ('Cricket', 'sports'),
    ('Football', 'sports'),
    ('Badminton', 'sports'),
    ('Gym', 'fitness'),
    ('Yoga', 'fitness'),
    ('Photography', 'hobbies'),
    ('Cooking', 'hobbies'),
    ('Reading', 'hobbies'),
    ('Bollywood', 'entertainment'),
    ('Travel', 'lifestyle'),
    ('Food', 'lifestyle'),
    ('Music', 'entertainment'),
    ('Gaming', 'entertainment'),
    ('Startups', 'business'),
    ('Technology', 'business')
) AS interests(interest, category)
WHERE FALSE; -- Don't actually insert, just create template

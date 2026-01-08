-- Karma & Engagement System
-- User karma tracking
CREATE TABLE user_karma (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    total_karma INT DEFAULT 0,
    current_tier VARCHAR(50) DEFAULT 'navasadhak',
    prabhav_score DECIMAL(10,2) DEFAULT 0,
    login_streak INT DEFAULT 0,
    last_login_date DATE,
    last_calculated_at TIMESTAMP DEFAULT NOW(),
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Karma transaction log
CREATE TABLE karma_transactions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    action VARCHAR(50) NOT NULL,
    points INT NOT NULL,
    reference_type VARCHAR(50), -- 'story', 'connection', 'reaction', 'login', etc.
    reference_id UUID, -- ID of the related entity
    description TEXT,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Badge definitions
CREATE TABLE badges (
    code VARCHAR(50) PRIMARY KEY,
    name_sanskrit VARCHAR(100) NOT NULL,
    name_hindi VARCHAR(100) NOT NULL,
    description TEXT,
    min_karma INT NOT NULL,
    icon_url TEXT,
    sort_order INT DEFAULT 0
);

-- User badges earned
CREATE TABLE user_badges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    badge_code VARCHAR(50) NOT NULL REFERENCES badges(code),
    earned_at TIMESTAMP DEFAULT NOW(),
    UNIQUE(user_id, badge_code)
);

-- Premium subscriptions
CREATE TABLE subscriptions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    plan VARCHAR(20) NOT NULL, -- 'plus_monthly', 'plus_annual', 'karma_free'
    status VARCHAR(20) DEFAULT 'active', -- 'active', 'cancelled', 'expired'
    started_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL,
    payment_provider VARCHAR(20), -- 'razorpay', 'karma', 'promotional'
    payment_id VARCHAR(100),
    amount_paid INT, -- in paise
    created_at TIMESTAMP DEFAULT NOW(),
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Story boosts
CREATE TABLE story_boosts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    story_id UUID NOT NULL REFERENCES stories(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    boost_type VARCHAR(20) NOT NULL, -- 'paid', 'premium_free', 'karma'
    boost_multiplier DECIMAL(3,1) DEFAULT 2.0,
    boosted_at TIMESTAMP DEFAULT NOW(),
    expires_at TIMESTAMP NOT NULL
);

-- Indexes for performance
CREATE INDEX idx_karma_transactions_user ON karma_transactions(user_id);
CREATE INDEX idx_karma_transactions_created ON karma_transactions(created_at);
CREATE INDEX idx_user_badges_user ON user_badges(user_id);
CREATE INDEX idx_subscriptions_user ON subscriptions(user_id);
CREATE INDEX idx_subscriptions_status ON subscriptions(status, expires_at);
CREATE INDEX idx_story_boosts_story ON story_boosts(story_id);
CREATE INDEX idx_story_boosts_active ON story_boosts(expires_at);

-- Seed default badges
INSERT INTO badges (code, name_sanskrit, name_hindi, description, min_karma, sort_order) VALUES
('navasadhak', 'नवसाधक', 'नया साधक', 'Just getting started on your journey', 0, 1),
('prayatnasheel', 'प्रयत्नशील', 'प्रयास करने वाला', 'The one who tries with dedication', 100, 2),
('saathi', 'साथी', 'साथी', 'A trusted companion in the community', 500, 3),
('margdarshak', 'मार्गदर्शक', 'मार्गदर्शक', 'A guide who helps others find their way', 1000, 4),
('prabhavshali', 'प्रभावशाली', 'प्रभावशाली', 'One who makes an impact', 2500, 5),
('lokpriya', 'लोकप्रिय', 'लोकप्रिय', 'Beloved by the people', 5000, 6),
('yogi', 'योगी', 'योगी', 'Enlightened master of the community', 10000, 7);

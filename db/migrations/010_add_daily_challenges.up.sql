-- Daily Challenges System
CREATE TABLE daily_challenges (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    challenge_type VARCHAR(50) NOT NULL,
    title VARCHAR(100) NOT NULL,
    title_hindi VARCHAR(100),
    description TEXT,
    reward_karma INT NOT NULL,
    is_completed BOOLEAN DEFAULT FALSE,
    progress INT DEFAULT 0,
    target INT DEFAULT 1,
    assigned_date DATE DEFAULT CURRENT_DATE,
    completed_at TIMESTAMP,
    expires_at TIMESTAMP NOT NULL,
    created_at TIMESTAMP DEFAULT NOW()
);

-- Challenge templates (for random assignment)
CREATE TABLE challenge_templates (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,
    title VARCHAR(100) NOT NULL,
    title_hindi VARCHAR(100),
    description TEXT,
    description_hindi TEXT,
    reward_karma INT NOT NULL,
    target INT DEFAULT 1,
    difficulty VARCHAR(20) DEFAULT 'easy', -- easy, medium, hard
    is_active BOOLEAN DEFAULT TRUE
);

-- Festival events for bonus karma
CREATE TABLE festival_events (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    name_hindi VARCHAR(100),
    description TEXT,
    karma_multiplier DECIMAL(3,2) DEFAULT 2.0,
    bonus_karma INT DEFAULT 0,
    start_date TIMESTAMP NOT NULL,
    end_date TIMESTAMP NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT NOW()
);

-- User streaks and achievements
CREATE TABLE user_streaks (
    user_id UUID PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    current_streak INT DEFAULT 0,
    longest_streak INT DEFAULT 0,
    last_challenge_date DATE,
    total_challenges_completed INT DEFAULT 0,
    updated_at TIMESTAMP DEFAULT NOW()
);

-- Indexes
CREATE INDEX idx_daily_challenges_user ON daily_challenges(user_id, assigned_date);
CREATE INDEX idx_daily_challenges_active ON daily_challenges(user_id, is_completed, expires_at);
CREATE INDEX idx_festival_events_active ON festival_events(start_date, end_date, is_active);

-- Seed challenge templates
INSERT INTO challenge_templates (type, title, title_hindi, description, description_hindi, reward_karma, target, difficulty) VALUES
-- Content Creation
('post_story', 'Share Your Day', 'अपना दिन शेयर करो', 'Post a story today', 'आज एक स्टोरी पोस्ट करो', 15, 1, 'easy'),
('post_stories', 'Story Teller', 'कहानीकार', 'Post 3 stories today', 'आज 3 स्टोरी पोस्ट करो', 40, 3, 'medium'),
('morning_story', 'Good Morning!', 'सुप्रभात!', 'Post a story before 9 AM', 'सुबह 9 बजे से पहले स्टोरी पोस्ट करो', 25, 1, 'medium'),
('evening_story', 'Evening Vibes', 'शाम की खुशबू', 'Post a story after 6 PM', 'शाम 6 बजे के बाद स्टोरी पोस्ट करो', 20, 1, 'easy'),

-- Social Interactions  
('react_stories', 'Spread Love', 'प्यार बांटो', 'React to 5 stories', '5 स्टोरीज पर रिएक्ट करो', 20, 5, 'easy'),
('react_stories_10', 'Super Supporter', 'सुपर सपोर्टर', 'React to 10 stories', '10 स्टोरीज पर रिएक्ट करो', 35, 10, 'medium'),
('reply_story', 'Start a Conversation', 'बातचीत शुरू करो', 'Reply to someone''s story', 'किसी की स्टोरी पर रिप्लाई करो', 25, 1, 'easy'),
('send_thanks', 'Say Thanks', 'धन्यवाद बोलो', 'Send a Thanks to someone helpful', 'किसी मददगार को धन्यवाद भेजो', 30, 1, 'medium'),

-- Connections
('new_connection', 'Make a Friend', 'दोस्त बनाओ', 'Connect with someone new', 'किसी नए से जुड़ो', 40, 1, 'medium'),
('chat_new', 'Break the Ice', 'बर्फ तोड़ो', 'Start a chat with a new connection', 'नए connection से चैट शुरू करो', 30, 1, 'medium'),

-- Engagement
('get_reactions', 'Get Noticed', 'नोटिस हो जाओ', 'Get 3 reactions on your stories', 'अपनी स्टोरीज पर 3 reactions पाओ', 35, 3, 'medium'),
('get_views', 'Viral Moment', 'वायरल पल', 'Get 20 views on a single story', 'एक स्टोरी पर 20 views पाओ', 50, 20, 'hard'),

-- Special (Indian context)
('local_explorer', 'Local Explorer', 'लोकल एक्स्प्लोरर', 'View 5 stories from your area', 'अपने इलाके की 5 स्टोरीज देखो', 25, 5, 'easy'),
('weekend_warrior', 'Weekend Warrior', 'वीकेंड वॉरियर', 'Complete all 3 challenges today', 'आज तीनों challenges पूरे करो', 100, 3, 'hard');

-- Seed upcoming Indian festivals (2024-2025)
INSERT INTO festival_events (name, name_hindi, description, karma_multiplier, bonus_karma, start_date, end_date) VALUES
('Makar Sankranti', 'मकर संक्रांति', 'Celebrate the harvest festival! 2x Karma all day', 2.0, 50, '2025-01-14 00:00:00', '2025-01-14 23:59:59'),
('Republic Day', 'गणतंत्र दिवस', 'Happy Republic Day! Share patriotic stories', 1.5, 75, '2025-01-26 00:00:00', '2025-01-26 23:59:59'),
('Holi', 'होली', 'Festival of Colors! 2x Karma and bonus', 2.0, 100, '2025-03-14 00:00:00', '2025-03-14 23:59:59'),
('Eid', 'ईद', 'Eid Mubarak! Share joy with everyone', 1.5, 50, '2025-03-31 00:00:00', '2025-03-31 23:59:59'),
('Independence Day', 'स्वतंत्रता दिवस', 'Jai Hind! Celebrate freedom', 1.5, 75, '2025-08-15 00:00:00', '2025-08-15 23:59:59'),
('Raksha Bandhan', 'रक्षा बंधन', 'Celebrate sibling love!', 1.5, 50, '2025-08-09 00:00:00', '2025-08-09 23:59:59'),
('Janmashtami', 'जन्माष्टमी', 'Happy Janmashtami!', 1.5, 50, '2025-08-16 00:00:00', '2025-08-16 23:59:59'),
('Ganesh Chaturthi', 'गणेश चतुर्थी', 'Ganpati Bappa Morya!', 2.0, 75, '2025-08-27 00:00:00', '2025-09-06 23:59:59'),
('Navratri', 'नवरात्रि', 'Nine nights of devotion', 1.5, 50, '2025-09-22 00:00:00', '2025-10-01 23:59:59'),
('Dussehra', 'दशहरा', 'Victory of good over evil!', 1.5, 50, '2025-10-02 00:00:00', '2025-10-02 23:59:59'),
('Karwa Chauth', 'करवा चौथ', 'Celebrate love and devotion', 1.5, 40, '2025-10-10 00:00:00', '2025-10-10 23:59:59'),
('Diwali', 'दीवाली', 'Festival of Lights! 3x Karma!', 3.0, 200, '2025-10-20 00:00:00', '2025-10-21 23:59:59'),
('New Year', 'नया साल', 'Happy New Year! Fresh start bonus', 2.0, 100, '2025-12-31 18:00:00', '2026-01-01 06:00:00');

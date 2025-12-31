-- Add profile fields to users table
ALTER TABLE users 
ADD COLUMN bio TEXT,
ADD COLUMN gender VARCHAR(20),
ADD COLUMN date_of_birth DATE,
ADD COLUMN visibility VARCHAR(20) DEFAULT 'public' NOT NULL;

-- Index for visibility if needed for filtering later
CREATE INDEX idx_users_visibility ON users(visibility);

-- Remove profile fields from users table
ALTER TABLE users 
DROP COLUMN IF EXISTS visibility,
DROP COLUMN IF EXISTS date_of_birth,
DROP COLUMN IF EXISTS gender,
DROP COLUMN IF EXISTS bio;

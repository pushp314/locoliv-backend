ALTER TABLE sessions ADD COLUMN fcm_token TEXT;
CREATE INDEX idx_sessions_fcm_token ON sessions(fcm_token);

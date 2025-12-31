CREATE EXTENSION IF NOT EXISTS "cube";
CREATE EXTENSION IF NOT EXISTS "earthdistance";

CREATE INDEX idx_stories_location ON stories USING gist (ll_to_earth(location_lat, location_lng));

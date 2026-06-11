CREATE TABLE user_locations (
    user_id     UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    location_id UUID NOT NULL REFERENCES locations(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, location_id)
);

-- Migrate existing user→location relationships before dropping the column.
INSERT INTO user_locations (user_id, location_id)
SELECT user_id, id FROM locations;

ALTER TABLE locations DROP COLUMN user_id;

-- Prevent duplicate location records for the same tidal station.
ALTER TABLE locations ADD CONSTRAINT locations_tide_station_id_key UNIQUE (tide_station_id);

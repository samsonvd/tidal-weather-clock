ALTER TABLE locations ADD COLUMN tide_station_id TEXT NOT NULL DEFAULT '';
UPDATE locations SET tide_station_id = '0158';
ALTER TABLE locations ALTER COLUMN tide_station_id DROP DEFAULT;

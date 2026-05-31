-- name: GetLocationByUser :one
SELECT * FROM locations WHERE user_id = $1 LIMIT 1;

-- name: GetAllLocations :many
SELECT * FROM locations;

-- name: CreateLocation :one
INSERT INTO locations (user_id, name, lat, lon, tide_station_id)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (user_id) DO UPDATE SET
    name           = EXCLUDED.name,
    lat            = EXCLUDED.lat,
    lon            = EXCLUDED.lon,
    tide_station_id = EXCLUDED.tide_station_id
RETURNING *;

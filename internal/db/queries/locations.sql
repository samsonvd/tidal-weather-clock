-- name: GetLocationByUser :one
SELECT l.* FROM locations l
JOIN user_locations ul ON ul.location_id = l.id
WHERE ul.user_id = $1
LIMIT 1;

-- name: GetLocationByStationID :one
SELECT * FROM locations WHERE tide_station_id = $1;

-- name: GetAllLocations :many
SELECT * FROM locations;

-- name: CreateLocation :one
INSERT INTO locations (name, lat, lon, tide_station_id)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: LinkUserLocation :exec
INSERT INTO user_locations (user_id, location_id) VALUES ($1, $2)
ON CONFLICT DO NOTHING;

-- name: DeleteUserLocations :exec
DELETE FROM user_locations WHERE user_id = $1;

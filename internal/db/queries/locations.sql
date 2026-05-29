-- name: EnsureLocation :exec
INSERT INTO locations (user_id, name, lat, lon)
VALUES ($1, 'Burnham Overy Staithe', 52.963583, 0.74417)
ON CONFLICT (user_id) DO NOTHING;

-- name: GetLocationByUser :one
SELECT * FROM locations WHERE user_id = $1 LIMIT 1;

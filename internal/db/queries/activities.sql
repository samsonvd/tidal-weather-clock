-- name: ListActivities :many
SELECT * FROM activities WHERE user_id = $1 ORDER BY created_at;

-- name: GetActivity :one
SELECT * FROM activities WHERE id = $1 AND user_id = $2;

-- name: CreateActivity :one
INSERT INTO activities (user_id, name, duration_hrs, window_start, window_end, constraints)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateActivity :one
UPDATE activities
SET name = $3,
    duration_hrs = $4,
    window_start = $5,
    window_end   = $6,
    constraints  = $7,
    updated_at   = now()
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: DeleteActivity :exec
DELETE FROM activities WHERE id = $1 AND user_id = $2;

-- name: GetOrCreateUser :one
INSERT INTO users (email)
VALUES ($1)
ON CONFLICT (email) DO UPDATE SET email = EXCLUDED.email
RETURNING *;

-- name: CreateMagicLink :one
INSERT INTO magic_links (token, user_id, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetMagicLink :one
SELECT * FROM magic_links
WHERE token = $1
  AND used_at IS NULL
  AND expires_at > now();

-- name: MarkMagicLinkUsed :exec
UPDATE magic_links
SET used_at = now()
WHERE token = $1;

-- name: CreateSession :one
INSERT INTO sessions (token, user_id, expires_at)
VALUES ($1, $2, $3)
RETURNING *;

-- name: GetSessionWithUser :one
SELECT
    s.token        AS session_token,
    s.expires_at   AS session_expires_at,
    s.created_at   AS session_created_at,
    u.id           AS user_id,
    u.email        AS user_email,
    u.created_at   AS user_created_at
FROM sessions s
JOIN users u ON u.id = s.user_id
WHERE s.token = $1
  AND s.expires_at > now();

-- name: DeleteSession :exec
DELETE FROM sessions WHERE token = $1;

-- name: DeleteExpiredSessions :exec
DELETE FROM sessions WHERE expires_at < now();

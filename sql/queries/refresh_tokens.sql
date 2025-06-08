-- name: GetRefreshToken :one
SELECT * FROM refresh_tokens
WHERE token = $1
LIMIT 1;

-- name: CreateRefreshToken :one
INSERT INTO refresh_tokens (token, user_id, expires_at, revoked_at)
VALUES ($1, $2, $3, NULL)
RETURNING *;

-- name: RevokeRefreshToken :exec
UPDATE refresh_tokens SET revoked_at = NOW(), updated_at = NOW() WHERE token = $1;
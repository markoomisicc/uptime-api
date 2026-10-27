-- name: CreateSession :exec
INSERT INTO sessions (user_id, token_hash, user_agent, ip_address, expires_at)
VALUES ($1, $2, $3, $4, $5);

-- name: GetUserBySessionToken :one
SELECT sqlc.embed(users)
FROM sessions
JOIN users ON users.id = sessions.user_id
WHERE sessions.token_hash = $1
  AND sessions.expires_at > now();

-- name: DeleteSession :exec
DELETE FROM sessions
WHERE token_hash = $1;
-- name: CreateUserToken :exec
INSERT INTO user_tokens (user_id, purpose, token_hash, expires_at)
VALUES ($1, $2, $3, $4);

-- name: DeleteUserTokens :exec
DELETE FROM user_tokens
WHERE user_id = $1 AND purpose = $2;

-- name: ConsumeUserToken :one
DELETE FROM user_tokens
WHERE token_hash = $1
  AND purpose = $2
  AND expires_at > now()
RETURNING user_id;
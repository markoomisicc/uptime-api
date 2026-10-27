-- name: CreateUser :one
INSERT INTO users (email, username, password_hash, first_name, last_name, phone, timezone, language)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE lower(email) = lower($1);

-- name: UpdateUserLastLogin :exec
UPDATE users SET last_login_at = now()
WHERE id = $1;

-- name: UpdateUserLanguage :one
UPDATE users SET language = $2, updated_at = now()
WHERE id = $1
RETURNING *;
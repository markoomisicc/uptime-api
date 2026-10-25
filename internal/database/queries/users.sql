-- name: CreateUser :one
INSERT INTO users (email, username, password_hash, first_name, last_name, phone, timezone)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: GetUserByEmail :one
SELECT * FROM users
WHERE lower(email) = lower($1);
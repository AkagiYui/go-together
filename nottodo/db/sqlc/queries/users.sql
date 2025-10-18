-- name: CreateUser :one
INSERT INTO users (username, nickname, password_hash)
VALUES ($1, $2, $3)
RETURNING id, username, nickname, created_at;

-- name: DeleteUser :execrows
DELETE FROM users WHERE id = $1;

-- name: GetUserByUsername :one
SELECT id, username, nickname, password_hash, created_at
FROM users WHERE username = $1;

-- name: UpdateUserPassword :execrows
UPDATE users SET password_hash = $1, updated_at = NOW() WHERE username = $2;

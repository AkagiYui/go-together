-- 本文件用于存放 User 相关的 SQL 语句，按 sqlc 约定的注释格式标注方法名

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

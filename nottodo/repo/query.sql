-- 缓存 ===============================

-- name: GetCache :one
-- 获取一个缓存项，同时返回它的值和过期时间
SELECT value, expires_at FROM app_cache
WHERE key = $1;

-- name: SetCache :exec
-- 使用 UPSERT (INSERT ... ON CONFLICT) 来设置缓存，这是原子操作
INSERT INTO app_cache (key, value, expires_at)
VALUES ($1, $2, $3)
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value,
    expires_at = EXCLUDED.expires_at;

-- name: DeleteCache :exec
-- 删除一个缓存项
DELETE FROM app_cache
WHERE key = $1;

-- name: PurgeExpiredCache :exec
-- 清理所有已经过期的缓存项
DELETE FROM app_cache
WHERE expires_at < NOW();


-- 系统设置 ===============================

-- name: GetSetting :one
SELECT * FROM settings
WHERE key = $1;

-- name: ListSettings :many
SELECT * FROM settings;

-- name: SetSetting :one
INSERT INTO settings (key, value)
VALUES ($1, $2)
ON CONFLICT (key) DO UPDATE
SET value = EXCLUDED.value, updated_at = NOW()
RETURNING *;

-- name: DeleteSetting :exec
DELETE FROM settings
WHERE key = $1;


-- 待办事项 ===============================

-- name: ListTodos :many
SELECT * FROM todos
ORDER BY id;

-- name: CountTodos :one
SELECT COUNT(*) FROM todos;

-- name: GetTodo :one
SELECT * FROM todos
WHERE id = $1;

-- name: DeleteTodo :exec
DELETE FROM todos
WHERE id = $1;

-- name: CreateTodo :one
INSERT INTO todos (title, description, completed)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateTodo :exec
UPDATE todos
SET title = $2, description = $3, completed = $4
WHERE id = $1;


-- 用户 ===============================

-- name: GetUser :one
SELECT * FROM users
WHERE id = $1;

-- name: GetUserByUsername :one
SELECT * FROM users
WHERE username = $1;

-- name: ListUsers :many
SELECT * FROM users;

-- name: CountUsers :one
SELECT COUNT(*) FROM users;

-- name: CreateUser :one
INSERT INTO users (username, password, nickname)
VALUES ($1, $2, $3)
RETURNING *;

-- name: UpdateUserPassword :one
UPDATE users
SET password = $2
WHERE id = $1
RETURNING *;

-- 邮箱 ================================

-- name: ListEmails :many
SELECT * FROM emails;

-- name: ListEmailsByUserId :many
SELECT * FROM emails
WHERE user_id = $1;

-- name: GetEmail :one
SELECT * FROM emails
WHERE id = $1;

-- name: GetEmailByAddress :one
SELECT * FROM emails
WHERE email = $1;

-- name: CreateEmail :one
INSERT INTO emails (user_id, email, is_primary, is_verified)
VALUES ($1, $2, $3, $4)
RETURNING *;

-- name: UpdateEmail :exec
UPDATE emails
SET is_primary = $2, is_verified = $3
WHERE id = $1;

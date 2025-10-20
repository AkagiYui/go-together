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

-- name: CreateTodo :exec
INSERT INTO todos (title, description, completed)
VALUES ($1, $2, $3);


-- ===============================

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

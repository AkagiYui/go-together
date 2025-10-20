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

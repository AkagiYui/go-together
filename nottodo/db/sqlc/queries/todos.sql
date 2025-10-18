-- name: ListTodos :many
SELECT id, title, description, completed, created_at
FROM todos
ORDER BY id;

-- name: GetTodo :one
SELECT id, title, description, completed, created_at
FROM todos
WHERE id = $1;

-- name: CreateTodo :one
INSERT INTO todos (title, description, completed)
VALUES ($1, $2, $3)
RETURNING id, title, description, completed, created_at;

-- name: UpdateTodo :execrows
UPDATE todos
SET title = $1,
    description = $2,
    completed = $3
WHERE id = $4;

-- name: DeleteTodo :execrows
DELETE FROM todos WHERE id = $1;

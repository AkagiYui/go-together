package todo

import (
	"context"
	"database/sql"
)

// Todo 领域模型（对应 todos 表）
// 注意：这里作为 sqlc 的“生成代码”示例，包含了 SQL 与查询方法。
// 业务层不要在其它 Go 文件中书写原生 SQL，而是调用这里的方法。
type Todo struct {
	ID          int    `json:"id"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
	CreatedAt   string `json:"created_at"`
}

// Queries 提供 todos 表的 CRUD 方法
// 通常由 sqlc 生成，这里为便于示例手写等效代码。
// 仅此处允许使用原生 SQL，其它位置请通过本类型对数据库进行访问。
type Queries struct{ db *sql.DB }

func New(db *sql.DB) *Queries { return &Queries{db: db} }

func (q *Queries) ListTodos(ctx context.Context) ([]Todo, error) {
	rows, err := q.db.QueryContext(ctx, `SELECT id, title, description, completed, created_at FROM todos ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	list := make([]Todo, 0)
	for rows.Next() {
		var t Todo
		var createdAt sql.NullTime
		if err := rows.Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &createdAt); err != nil {
			return nil, err
		}
		if createdAt.Valid {
			t.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
		}
		list = append(list, t)
	}
	return list, rows.Err()
}

func (q *Queries) GetTodo(ctx context.Context, id int) (Todo, error) {
	row := q.db.QueryRowContext(ctx, `SELECT id, title, description, completed, created_at FROM todos WHERE id = $1`, id)
	var t Todo
	var createdAt sql.NullTime
	if err := row.Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &createdAt); err != nil {
		return Todo{}, err
	}
	if createdAt.Valid {
		t.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
	}
	return t, nil
}

func (q *Queries) CreateTodo(ctx context.Context, title, description string, completed bool) (Todo, error) {
	row := q.db.QueryRowContext(ctx, `INSERT INTO todos (title, description, completed) VALUES ($1, $2, $3) RETURNING id, title, description, completed, created_at`, title, description, completed)
	var t Todo
	var createdAt sql.NullTime
	if err := row.Scan(&t.ID, &t.Title, &t.Description, &t.Completed, &createdAt); err != nil {
		return Todo{}, err
	}
	if createdAt.Valid {
		t.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
	}
	return t, nil
}

func (q *Queries) UpdateTodo(ctx context.Context, id int, title, description string, completed bool) (int64, error) {
	res, err := q.db.ExecContext(ctx, `UPDATE todos SET title = $1, description = $2, completed = $3 WHERE id = $4`, title, description, completed, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (q *Queries) DeleteTodo(ctx context.Context, id int) (int64, error) {
	res, err := q.db.ExecContext(ctx, `DELETE FROM todos WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

package user

import (
	"context"
	"database/sql"

	"golang.org/x/crypto/bcrypt"
)

// User 领域模型（对应 users 表）
// 作为 sqlc 的“生成代码”示例，仅此处包含 SQL，其它位置不要书写原生 SQL。
type User struct {
	ID           int64  `json:"id"`
	Username     string `json:"username"`
	Nickname     string `json:"nickname"`
	PasswordHash string `json:"-"`
	CreatedAt    string `json:"created_at"`
}

// Queries 提供 users 表的增删改查
// 通常由 sqlc 生成，这里为便于示例手写等效代码。
type Queries struct{ db *sql.DB }

func New(db *sql.DB) *Queries { return &Queries{db: db} }

func (q *Queries) CreateUser(ctx context.Context, username, nickname, password string) (User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return User{}, err
	}
	row := q.db.QueryRowContext(ctx, `INSERT INTO users (username, nickname, password_hash) VALUES ($1, $2, $3) RETURNING id, username, nickname, created_at`, username, nickname, string(hash))
	var u User
	var createdAt sql.NullTime
	if err := row.Scan(&u.ID, &u.Username, &u.Nickname, &createdAt); err != nil {
		return User{}, err
	}
	if createdAt.Valid {
		u.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
	}
	return u, nil
}

func (q *Queries) DeleteUser(ctx context.Context, id int64) (int64, error) {
	res, err := q.db.ExecContext(ctx, `DELETE FROM users WHERE id = $1`, id)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

func (q *Queries) GetUserByUsername(ctx context.Context, username string) (User, error) {
	row := q.db.QueryRowContext(ctx, `SELECT id, username, nickname, password_hash, created_at FROM users WHERE username = $1`, username)
	var u User
	var createdAt sql.NullTime
	if err := row.Scan(&u.ID, &u.Username, &u.Nickname, &u.PasswordHash, &createdAt); err != nil {
		return User{}, err
	}
	if createdAt.Valid {
		u.CreatedAt = createdAt.Time.Format("2006-01-02 15:04:05")
	}
	return u, nil
}

func (q *Queries) UpdateUserPassword(ctx context.Context, username, newPassword string) (int64, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return 0, err
	}
	res, err := q.db.ExecContext(ctx, `UPDATE users SET password_hash = $1, updated_at = NOW() WHERE username = $2`, string(hash), username)
	if err != nil {
		return 0, err
	}
	return res.RowsAffected()
}

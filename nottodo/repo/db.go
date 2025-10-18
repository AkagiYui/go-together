package repo

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "github.com/jackc/pgx/v5/stdlib"
)

var db *sql.DB

// InitDB 初始化数据库连接。优先从参数 dsn 读取，其次从环境变量 DATABASE_URL 读取。
// DSN 形如：postgres://user:pass@host:5432/dbname?sslmode=disable
func InitDB(dsn string) error {
	if strings.TrimSpace(dsn) == "" {
		dsn = os.Getenv("DATABASE_URL")
	}
	if strings.TrimSpace(dsn) == "" {
		// 不强制要求提供 DSN，使用一个占位值，避免在构建/启动阶段因为缺少环境变量而直接报错。
		// 注意：未正确配置数据库时，运行时的数据库操作会失败。
		dsn = "postgres://postgres:postgres@localhost:5432/nottodo?sslmode=disable"
	}

	// 使用 pgx 的 database/sql 驱动
	var err error
	db, err = sql.Open("pgx", dsn)
	if err != nil {
		return err
	}
	// 不在这里调用 Ping，避免 CI 或本地无数据库时直接报错。

	// 尝试创建需要的表（幂等）。
	return migrate()
}

func migrate() error {
	if db == nil {
		return fmt.Errorf("db is not initialized")
	}

	stmts := []string{
		`CREATE TABLE IF NOT EXISTS todos (
			id SERIAL PRIMARY KEY,
			title TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '',
			completed BOOLEAN NOT NULL DEFAULT FALSE,
			created_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
		`CREATE TABLE IF NOT EXISTS users (
			id BIGSERIAL PRIMARY KEY,
			username VARCHAR(255) NOT NULL UNIQUE,
			nickname VARCHAR(255) NOT NULL DEFAULT '',
			password_hash TEXT NOT NULL,
			created_at TIMESTAMP NOT NULL DEFAULT NOW(),
			updated_at TIMESTAMP NOT NULL DEFAULT NOW()
		)`,
	}

	for _, s := range stmts {
		if _, err := db.Exec(s); err != nil {
			return err
		}
	}
	return nil
}

// internal helpers
func scanTodoRow(row *sql.Row) (Todo, error) {
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

func scanTodoRows(rows *sql.Rows) ([]Todo, error) {
	list := make([]Todo, 0)
	defer rows.Close()
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

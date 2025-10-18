package repo

import (
    "database/sql"
    "fmt"
    "os"
    "strings"

    _ "github.com/jackc/pgx/v5/stdlib"

    todogen "github.com/akagiyui/go-together/nottodo/repo/todo"
    usergen "github.com/akagiyui/go-together/nottodo/repo/user"
)

// DB 全局数据库连接，由 InitDB 初始化
var DB *sql.DB

// TodoQueries 和 UserQueries 为“sqlc 生成”的查询实例，供业务层调用
var (
    TodoQueries *todogen.Queries
    UserQueries *usergen.Queries
)

// InitDB 初始化数据库连接。只从环境变量或入参读取 DSN。
// DSN 形如：postgres://user:pass@host:5432/dbname?sslmode=disable
func InitDB(dsn string) error {
    if strings.TrimSpace(dsn) == "" {
        dsn = os.Getenv("DATABASE_URL")
    }
    if strings.TrimSpace(dsn) == "" {
        return fmt.Errorf("缺少必需的数据库配置: DATABASE_URL")
    }

    // 使用 pgx 的 database/sql 驱动
    var err error
    DB, err = sql.Open("pgx", dsn)
    if err != nil {
        return err
    }
    // 不在这里调用 Ping，避免 CI 或本地无数据库时直接报错。

    // 尝试创建需要的表（幂等）。
    if err := migrate(); err != nil {
        return err
    }

    // 初始化查询实例（模拟 sqlc 生成的 New 方法）
    TodoQueries = todogen.New(DB)
    UserQueries = usergen.New(DB)
    return nil
}

func migrate() error {
    if DB == nil {
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
        if _, err := DB.Exec(s); err != nil {
            return err
        }
    }
    return nil
}

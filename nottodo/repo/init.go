package repo

import (
	"context"
	"fmt"
	"strings"

	"github.com/akagiyui/go-together/nottodo/config"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/tracelog"
)

var (
	Db   *Queries
	conn *pgx.Conn
	Ctx  = context.Background()
)

// sqlLogger 实现 pgx 的日志接口
type sqlLogger struct{}

func (l *sqlLogger) Log(ctx context.Context, level tracelog.LogLevel, msg string, data map[string]any) {
	// 只打印 SQL 语句相关的日志
	if sql, ok := data["sql"]; ok {
		args := data["args"]
		fmt.Printf("[SQL] %s | args: %v\n", strings.TrimSpace(sql.(string)), args)
	}
}

func init() {
	var err error

	// 解析 DSN 配置
	connConfig, err := pgx.ParseConfig(config.GlobalConfig.DSN)
	if err != nil {
		panic(err)
	}

	// 如果是开发模式,启用 SQL 日志
	if config.GlobalConfig.Mode == config.ModeDev {
		connConfig.Tracer = &tracelog.TraceLog{
			Logger:   &sqlLogger{},
			LogLevel: tracelog.LogLevelInfo,
		}
	}

	// 使用配置连接数据库
	conn, err = pgx.ConnectConfig(Ctx, connConfig)
	if err != nil {
		panic(err)
	}
	Db = New(conn)
}

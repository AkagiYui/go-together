package repo

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/akagiyui/go-together/common/task"

	"github.com/akagiyui/go-together/nottodo/config"
)

var (
	// DB 数据库实例
	DB *gorm.DB
)

// customLogger 自定义 GORM 日志记录器
type customLogger struct {
	logger.Interface
}

// Trace 实现 logger.Interface 的 Trace 方法
func (l *customLogger) Trace(_ context.Context, begin time.Time, fc func() (string, int64), err error) {
	// 获取 SQL 和影响的行数
	sql, rows := fc()

	task.Run(func() {

	})

	// 清理 SQL 语句
	sqlText := strings.TrimSpace(sql)
	// 去除所有注释
	sqlText = regexp.MustCompile(`--.*`).ReplaceAllString(sqlText, "")
	// 去除所有换行，替换为空格
	sqlText = regexp.MustCompile(`\s+`).ReplaceAllString(sqlText, " ")

	// 打印 SQL 日志
	if err != nil {
		fmt.Printf("[SQL] %s | rows: %d | error: %v | elapsed: %v\n", sqlText, rows, err, time.Since(begin))
	} else {
		fmt.Printf("[SQL] %s | rows: %d | elapsed: %v\n", sqlText, rows, time.Since(begin))
	}
}

func init() {
	var err error

	// 配置 GORM logger
	var gormLogger logger.Interface
	if config.GlobalConfig.Mode == config.ModeDev {
		// 开发模式：使用自定义 logger
		gormLogger = &customLogger{
			Interface: logger.New(
				log.New(os.Stdout, "\r\n", log.LstdFlags),
				logger.Config{
					SlowThreshold:             200 * time.Millisecond,
					LogLevel:                  logger.Info,
					IgnoreRecordNotFoundError: false,
					Colorful:                  true,
				},
			),
		}
	} else {
		// 生产模式：只记录错误
		gormLogger = logger.New(
			log.New(os.Stdout, "\r\n", log.LstdFlags),
			logger.Config{
				SlowThreshold:             200 * time.Millisecond,
				LogLevel:                  logger.Error,
				IgnoreRecordNotFoundError: true,
				Colorful:                  false,
			},
		)
	}

	// 连接数据库
	// 注意：数据库结构由 schema.sql 文件管理，GORM 仅用于数据操作
	// 禁用自动迁移功能，防止 GORM 自动修改数据库结构
	DB, err = gorm.Open(postgres.Open(config.GlobalConfig.DSN), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now()
		},
		// 禁用外键约束检查（由数据库本身管理）
		DisableForeignKeyConstraintWhenMigrating: true,
		// 跳过默认事务（提高性能，因为我们不使用 AutoMigrate）
		SkipDefaultTransaction: false,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}

	// 重要提示：
	// 1. 不要在此处调用 DB.AutoMigrate()
	// 2. 数据库表结构由 nottodo/repo/schema.sql 文件定义和管理
	// 3. 使用数据库迁移工具（如 golang-migrate）来管理 schema 变更
	// 4. GORM 仅用于执行 CRUD 操作，不负责 DDL（数据定义语言）操作
}

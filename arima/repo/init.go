package repo

import (
	"context"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
	"time"

	"github.com/akagiyui/go-together/arima/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
	sql, rows := fc()

	sqlText := strings.TrimSpace(sql)
	sqlText = regexp.MustCompile(`--.*`).ReplaceAllString(sqlText, "")
	sqlText = regexp.MustCompile(`\s+`).ReplaceAllString(sqlText, " ")

	if err != nil {
		fmt.Printf("[SQL] %s | rows: %d | error: %v | elapsed: %v\n", sqlText, rows, err, time.Since(begin))
	} else {
		fmt.Printf("[SQL] %s | rows: %d | elapsed: %v\n", sqlText, rows, time.Since(begin))
	}
}

func init() {
	var err error

	var gormLogger logger.Interface
	if config.GlobalConfig.Mode == config.ModeDev {
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

	DB, err = gorm.Open(postgres.Open(config.GlobalConfig.DSN), &gorm.Config{
		Logger: gormLogger,
		NowFunc: func() time.Time {
			return time.Now()
		},
		DisableForeignKeyConstraintWhenMigrating: true,
		SkipDefaultTransaction:                   false,
	})
	if err != nil {
		panic(fmt.Sprintf("failed to connect database: %v", err))
	}
}


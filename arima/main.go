// Package main 是 Arima (Music Database Backend) 应用的入口
package main

import (
	"fmt"
	"log/slog"

	"github.com/akagiyui/go-together/arima/config"

	_ "github.com/akagiyui/go-together/arima/pkg/s3" // 初始化 S3 客户端
	_ "github.com/akagiyui/go-together/arima/repo"   // 初始化数据库
)

const banner = `
    _         _
   / \   _ __(_)_ __ ___   __ _
  / _ \ | '__| | '_ ` + "`" + ` _ \ / _` + "`" + ` |
 / ___ \| |  | | | | | | | (_| |
/_/   \_\_|  |_|_| |_| |_|\__,_|

  Music Database Backend (Go)
`

func main() {
	// 显示启动banner
	println(banner)

	// 读取配置
	cfg := config.GlobalConfig

	// 设置日志级别
	level := slog.LevelInfo
	if cfg.Mode == config.ModeDev {
		level = slog.LevelDebug
	}
	slog.SetLogLoggerLevel(level)

	// 启动服务器
	if err := s.Run(fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)); err != nil {
		panic(err)
	}
}

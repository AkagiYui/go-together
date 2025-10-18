package main

import (
	"github.com/akagiyui/go-together/nottodo/config"
)

func main() {
	// 读取配置（仅用于环境与模式控制，当前无需数据库）
	cfg := config.GlobalConfig

	// 开启交互式终端（仅开发模式）
	runInteractiveShell(cfg.Mode)

	if err := s.Run(":8080"); err != nil {
		panic(err)
	}
}

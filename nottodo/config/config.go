package config

import (
	"bufio"
	"os"
	"strings"

	"github.com/akagiyui/go-together/common/validation"
)

// Mode 应用运行模式的枚举
// 仅在开发模式下开启终端命令交互功能
// 默认生产模式
//
// 使用方式：
//
//	if cfg.Mode == ModeDev { ... }
type Mode string

const (
	ModeProd Mode = "prod"
	ModeDev  Mode = "dev"
)

func (m Mode) String() string {
	switch m {
	case ModeDev:
		return "dev"
	default:
		return "prod"
	}
}

// ParseMode 将字符串解析为枚举
func ParseMode(s string) Mode {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "dev" || s == "development" {
		return ModeDev
	}
	return ModeProd
}

// Config 应用配置，仅从环境变量读取
// 目前 DSN 可为空（使用内存数据库），MODE 为可选，默认 prod
type Config struct {
	DSN  string `validate:"required"`
	Mode Mode   `validate:"oneof=prod dev"`
}

// Load 尝试先加载 .env 文件，再从环境变量读取配置
func Load() (Config, error) {
	_ = loadDotenv(".env")

	dsn := strings.TrimSpace(os.Getenv("DATABASE_URL"))
	mode := ParseMode(os.Getenv("MODE"))
	cfg := Config{DSN: dsn, Mode: mode}
	return cfg, validation.ValidateStruct(cfg)
}

// 轻量级 .env 加载器：仅支持 KEY=VALUE 形式，# 开头为注释
// 若环境变量已存在，不覆盖
func loadDotenv(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()

	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		// 支持简单的 KEY=VALUE，忽略带引号情况
		if i := strings.IndexByte(line, '='); i > 0 {
			key := strings.TrimSpace(line[:i])
			val := strings.TrimSpace(line[i+1:])
			if _, exists := os.LookupEnv(key); !exists {
				_ = os.Setenv(key, val)
			}
		}
	}
	return nil
}

// Package config 提供应用配置管理功能
package config

import (
	"bufio"
	"os"
	"strings"

	"github.com/akagiyui/go-together/common/validation"
)

// Config 应用配置，仅从环境变量读取
// 目前 DSN 可为空（使用内存数据库），MODE 为可选，默认 prod
type Config struct {
	DSN         string `validate:"required"`
	Mode        Mode   `validate:"oneof=prod dev"`
	Port        string `validate:"required"`
	Host        string `validate:"required"`
	AllowOrigin string
}

// Load 尝试先加载 .env 文件，再从环境变量读取配置
func Load() (Config, error) {
	_ = loadDotenv(".env")

	cfg := Config{
		DSN:         getEnv("DSN", "", true),
		Mode:        ParseMode(getEnv("MODE", "prod", true)),
		Port:        getEnv("PORT", "8082", true),
		Host:        getEnv("HOST", "0.0.0.0", true),
		AllowOrigin: getEnv("ALLOW_ORIGIN", "", true),
	}
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

func getEnv(key string, defaultValue string, trim bool) string {
	val := os.Getenv(key)
	if val == "" {
		return defaultValue
	}
	if trim {
		return strings.TrimSpace(val)
	}
	return val
}

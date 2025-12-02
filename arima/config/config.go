package config

import (
	"bufio"
	"os"
	"strings"

	"github.com/akagiyui/go-together/common/validation"
)

// Config 应用配置，从环境变量读取
type Config struct {
	// 数据库配置
	DSN string `validate:"required"`

	// 服务器配置
	Mode        Mode   `validate:"oneof=prod dev"`
	Port        string `validate:"required"`
	Host        string `validate:"required"`
	AllowOrigin string

	// S3 配置
	S3Endpoint  string `validate:"required"`
	S3AccessKey string `validate:"required"`
	S3SecretKey string `validate:"required"`
	S3Bucket    string `validate:"required"`
	S3Region    string

	// FFmpeg 配置
	FFmpegExecutable  string
	FFprobeExecutable string

	// 管理 API Key
	ManageAPIKey string `validate:"required"`
}

// Load 尝试先加载 .env 文件，再从环境变量读取配置
func Load() (Config, error) {
	_ = loadDotenv(".env")

	cfg := Config{
		DSN:         getEnv("DSN", "", true),
		Mode:        ParseMode(getEnv("MODE", "prod", true)),
		Port:        getEnv("PORT", "8083", true),
		Host:        getEnv("HOST", "0.0.0.0", true),
		AllowOrigin: getEnv("ALLOW_ORIGIN", "", true),

		S3Endpoint:  getEnv("S3_ENDPOINT", "", true),
		S3AccessKey: getEnv("S3_ACCESS_KEY", "", true),
		S3SecretKey: getEnv("S3_SECRET_KEY", "", true),
		S3Bucket:    getEnv("S3_BUCKET", "", true),
		S3Region:    getEnv("S3_REGION", "", true),

		FFmpegExecutable:  getEnv("FFMPEG_EXECUTABLE", "ffmpeg", true),
		FFprobeExecutable: getEnv("FFPROBE_EXECUTABLE", "ffprobe", true),

		ManageAPIKey: getEnv("MANAGE_API_KEY", "", true),
	}
	return cfg, validation.ValidateStruct(cfg)
}

// loadDotenv 轻量级 .env 加载器：仅支持 KEY=VALUE 形式，# 开头为注释
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


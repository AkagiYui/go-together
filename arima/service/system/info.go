// Package system 提供系统相关服务
package system

import (
	"context"

	"github.com/akagiyui/go-together/arima/config"
	"github.com/akagiyui/go-together/arima/pkg/ffmpeg"
	"github.com/akagiyui/go-together/arima/pkg/s3"
	"github.com/akagiyui/go-together/arima/repo"
)

// GetSystemInfoRequest 获取系统信息请求
type GetSystemInfoRequest struct{}

// Info 系统信息响应
type Info struct {
	AppName           string `json:"app_name"`
	Version           string `json:"version"`
	Description       string `json:"description"`
	FFmpegExecutable  string `json:"ffmpeg_executable"`
	FFprobeExecutable string `json:"ffprobe_executable"`
	FFmpegVersion     string `json:"ffmpeg_version"`
	FFprobeVersion    string `json:"ffprobe_version"`
	S3Health          bool   `json:"s3_health"`
	DBHealth          bool   `json:"db_health"`
}

// Do 处理获取系统信息请求
func (r GetSystemInfoRequest) Do() (any, error) {
	cfg := config.GlobalConfig
	ff := ffmpeg.NewFFmpeg(cfg.FFmpegExecutable, cfg.FFprobeExecutable)

	ffmpegVersion, _ := ff.FFmpegVersion(context.Background())
	ffprobeVersion, _ := ff.FFprobeVersion(context.Background())

	// 检查 S3 健康状态
	s3Health := s3.S3Client.IsHealthy(context.Background())

	// 检查数据库健康状态
	dbHealth := false
	if sqlDB, err := repo.DB.DB(); err == nil {
		dbHealth = sqlDB.Ping() == nil
	}

	info := Info{
		AppName:           "arima",
		Version:           "0.1.0",
		Description:       "Music Database Backend (Go)",
		FFmpegExecutable:  cfg.FFmpegExecutable,
		FFprobeExecutable: cfg.FFprobeExecutable,
		FFmpegVersion:     ffmpegVersion,
		FFprobeVersion:    ffprobeVersion,
		S3Health:          s3Health,
		DBHealth:          dbHealth,
	}

	return info, nil
}

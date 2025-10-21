package main

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"log/slog"

	"github.com/akagiyui/go-together/nottodo/config"
	"resty.dev/v3"
)

func main() {
	// 读取配置（仅用于环境与模式控制，当前无需数据库）
	cfg := config.GlobalConfig

	// 设置日志级别
	level := slog.LevelInfo
	if cfg.Mode == config.ModeDev {
		level = slog.LevelDebug
	}
	slog.SetLogLoggerLevel(level)

	err := checkTimeDiff(60)
	if err != nil {
		panic(err)
	}

	// 开启交互式终端（仅开发模式）
	runInteractiveShell(cfg.Mode)

	if err := s.Run(":8082"); err != nil {
		panic(err)
	}
}

// 对比互联网时间
func checkTimeDiff(tollerance int64) error {
	client := resty.New()
	defer client.Close()
	res, err := client.R().Get("https://vv.video.qq.com/checktime?otype=json")
	if err != nil {
		slog.Error("无法从网络时间API获取时间，为了确保安全，停止运行。", slog.Any("Detail", err))
		return err
	}
	content := res.String()
	content = strings.ReplaceAll(content, "QZOutputJson=", "")
	content = strings.ReplaceAll(content, ";", "")

	type result struct {
		Timestamp int    `json:"t"`
		Ip        string `json:"ip"`
	}

	var res1 result
	err = json.Unmarshal([]byte(content), &res1)
	if err != nil {
		slog.Error("无法从网络时间API获取时间，为了确保安全，停止运行。", slog.Any("Detail", err))

		return err
	}

	timeDiff := time.Now().Unix() - int64(res1.Timestamp)
	if timeDiff < 0 {
		timeDiff = -timeDiff
	}

	// 时间差必须小于一分钟
	if timeDiff > tollerance {
		slog.Error("本地服务器时间与网络时间相差大于1分钟，为了确保安全，停止运行。", slog.Any("TimeDiff", timeDiff))
		return fmt.Errorf("time diff too large: %d", timeDiff)
	}
	slog.Debug("timediff check pass", slog.Any("TimeDiff", timeDiff))
	return nil
}

package config

import "strings"

// Mode 应用运行模式的枚举
// 仅在开发模式下开启终端命令交互功能
// 默认生产模式
//
// 使用方式：
//
//	if cfg.Mode == ModeDev { ... }
type Mode string

const (
	// ModeProd 生产模式
	ModeProd Mode = "prod"
	// ModeDev 开发模式
	ModeDev Mode = "dev"
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

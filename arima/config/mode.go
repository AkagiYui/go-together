// Package config 提供应用配置管理功能
package config

// Mode 表示应用运行模式
type Mode string

const (
	// ModeProd 生产模式
	ModeProd Mode = "prod"
	// ModeDev 开发模式
	ModeDev Mode = "dev"
)

// ParseMode 将字符串转换为运行模式
func ParseMode(s string) Mode {
	switch s {
	case "dev", "DEV", "development":
		return ModeDev
	default:
		return ModeProd
	}
}


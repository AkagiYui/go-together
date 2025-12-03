package main

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
)

// Config 主配置结构
type Config struct {
	Recording RecordingConfig `toml:"recording"`
	S3        S3Config        `toml:"s3"`
	Proxy     ProxyConfig     `toml:"proxy"`
	Cameras   []CameraConfig  `toml:"cameras"`
}

// RecordingConfig 录制配置
type RecordingConfig struct {
	SegmentDuration int    `toml:"segment_duration"`  // 分段时长(秒)
	StoragePath     string `toml:"storage_path"`      // 本地存储路径
	MaxDiskUsageGB  int    `toml:"max_disk_usage_gb"` // 最大磁盘使用量(GB)
}

// S3Config S3对象存储配置
type S3Config struct {
	Enabled   bool   `toml:"enabled"`
	Endpoint  string `toml:"endpoint"`
	Bucket    string `toml:"bucket"`
	AccessKey string `toml:"access_key"`
	SecretKey string `toml:"secret_key"`
	Region    string `toml:"region"`
}

// ProxyConfig HTTP代理配置
type ProxyConfig struct {
	Enabled  bool   `toml:"enabled"`
	Address  string `toml:"address"`
	Username string `toml:"username"`
	Password string `toml:"password"`
}

// CameraConfig 摄像头配置
type CameraConfig struct {
	Name    string `toml:"name"`
	RtspURL string `toml:"rtsp_url"`
	Enabled bool   `toml:"enabled"`
}

// LoadConfig 从TOML文件加载配置
func LoadConfig(path string) (*Config, error) {
	var config Config

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("读取配置文件失败: %w", err)
	}

	if _, err := toml.Decode(string(data), &config); err != nil {
		return nil, fmt.Errorf("解析配置文件失败: %w", err)
	}

	// 设置默认值
	if config.Recording.SegmentDuration <= 0 {
		config.Recording.SegmentDuration = 600 // 默认10分钟
	}
	if config.Recording.StoragePath == "" {
		config.Recording.StoragePath = "./recordings"
	}
	if config.Recording.MaxDiskUsageGB <= 0 {
		config.Recording.MaxDiskUsageGB = 50
	}
	if config.S3.Region == "" {
		config.S3.Region = "us-east-1"
	}

	// 验证配置
	if err := config.Validate(); err != nil {
		return nil, err
	}

	return &config, nil
}

// Validate 验证配置
func (c *Config) Validate() error {
	if len(c.Cameras) == 0 {
		return fmt.Errorf("至少需要配置一个摄像头")
	}

	for i, cam := range c.Cameras {
		if cam.Name == "" {
			return fmt.Errorf("摄像头 #%d 名称不能为空", i+1)
		}
		if cam.RtspURL == "" {
			return fmt.Errorf("摄像头 %s 的RTSP地址不能为空", cam.Name)
		}
	}

	if c.S3.Enabled {
		if c.S3.Endpoint == "" {
			return fmt.Errorf("S3端点不能为空")
		}
		if c.S3.Bucket == "" {
			return fmt.Errorf("S3存储桶不能为空")
		}
		if c.S3.AccessKey == "" || c.S3.SecretKey == "" {
			return fmt.Errorf("S3访问密钥不能为空")
		}
	}

	if c.Proxy.Enabled {
		if c.Proxy.Address == "" {
			return fmt.Errorf("代理地址不能为空")
		}
	}

	return nil
}

// GetProxyURL 获取完整的代理URL(包含认证信息)
func (c *ProxyConfig) GetProxyURL() string {
	if !c.Enabled || c.Address == "" {
		return ""
	}

	// 如果需要认证
	if c.Username != "" && c.Password != "" {
		// 解析地址并添加认证信息
		// 假设地址格式为 http://host:port
		if len(c.Address) > 7 && c.Address[:7] == "http://" {
			return fmt.Sprintf("http://%s:%s@%s", c.Username, c.Password, c.Address[7:])
		}
		if len(c.Address) > 8 && c.Address[:8] == "https://" {
			return fmt.Sprintf("https://%s:%s@%s", c.Username, c.Password, c.Address[8:])
		}
	}

	return c.Address
}

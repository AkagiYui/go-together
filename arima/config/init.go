package config

// GlobalConfig 全局配置实例
var GlobalConfig Config

func init() {
	var err error
	GlobalConfig, err = Load()
	if err != nil {
		panic(err)
	}
}


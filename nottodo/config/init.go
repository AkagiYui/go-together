package config

var GlobalConfig Config

func init() {
	var err error
	GlobalConfig, err = Load()
	if err != nil {
		panic(err)
	}
}

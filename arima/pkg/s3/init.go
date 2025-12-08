package s3

import (
	"github.com/akagiyui/go-together/arima/config"
)

// S3Client 全局 S3 客户端实例
var S3Client *Client

func init() {
	var err error
	S3Client, err = NewClient(config.GlobalConfig)
	if err != nil {
		panic(err)
	}
}

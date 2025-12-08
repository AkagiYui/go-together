package audio

import (
	"context"
	"time"

	"github.com/akagiyui/go-together/common/validation"

	"github.com/akagiyui/go-together/arima/pkg/s3"
	"github.com/akagiyui/go-together/arima/repo"
)

// GetOriginAudioDownloadURLRequest 获取原始音频下载URL请求
type GetOriginAudioDownloadURLRequest struct {
	ID int64 `path:"id"`
}

// Validate 校验请求参数
func (r GetOriginAudioDownloadURLRequest) Validate() error {
	return validation.PositiveInt64(r.ID, "ID")
}

// Do 处理获取原始音频下载URL请求
func (r GetOriginAudioDownloadURLRequest) Do() (any, error) {
	audio, err := repo.GetOriginAudioByID(r.ID)
	if err != nil {
		return "", err
	}

	url, err := s3.S3Client.GenerateDownloadURL(context.Background(), audio.FileKey, time.Hour, audio.FileName)
	if err != nil {
		return "", err
	}

	return url, nil
}

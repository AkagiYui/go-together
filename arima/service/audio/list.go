// Package audio 提供音频相关服务
package audio

import (
	"github.com/akagiyui/go-together/arima/repo"
	"github.com/akagiyui/go-together/common/model"
)

// ListAudioRequest 获取音频列表请求
type ListAudioRequest struct {
	PageIndex int `query:"page_index"`
	PageSize  int `query:"page_size"`
}

// Do 处理获取音频列表请求
func (r ListAudioRequest) Do() (any, error) {
	if r.PageIndex < 1 {
		r.PageIndex = 1
	}
	if r.PageSize < 1 {
		r.PageSize = 20
	}

	list, total, err := repo.GetAudioList(r.PageIndex, r.PageSize)
	if err != nil {
		return nil, err
	}

	return model.Page(total, list), nil
}

// ListOriginAudioRequest 获取原始音频列表请求
type ListOriginAudioRequest struct {
	PageIndex int `query:"page_index"`
	PageSize  int `query:"page_size"`
}

// Do 处理获取原始音频列表请求
func (r ListOriginAudioRequest) Do() (any, error) {
	if r.PageIndex < 1 {
		r.PageIndex = 1
	}
	if r.PageSize < 1 {
		r.PageSize = 20
	}

	list, total, err := repo.GetOriginAudioList(r.PageIndex, r.PageSize)
	if err != nil {
		return nil, err
	}

	return model.Page(total, list), nil
}

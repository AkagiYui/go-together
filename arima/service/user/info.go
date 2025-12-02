// Package user 提供用户相关服务
package user

import (
	"github.com/akagiyui/go-together/arima/repo"
)

// GetUserMeRequest 获取当前用户信息请求
type GetUserMeRequest struct {
	User repo.User `context:"user"`
}

// Do 处理获取当前用户信息请求
func (r GetUserMeRequest) Do() (any, error) {
	return r.User, nil
}

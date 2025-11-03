package user

import (
	"fmt"
	"time"

	"github.com/akagiyui/go-together/common/task"
	"github.com/akagiyui/go-together/nottodo/cache"
	"github.com/akagiyui/go-together/nottodo/repo"
)

// GetUserInfoRequest 获取用户信息请求
type GetUserInfoRequest struct {
	User repo.User `context:"user"`
}

// Do 执行获取用户信息的业务逻辑
func (r GetUserInfoRequest) Do() (any, error) {
	task.Run(func() {
		cache.Set(fmt.Sprintf("user:%d", r.User.ID), r.User, 24*time.Hour)
	})
	return r.User, nil
}

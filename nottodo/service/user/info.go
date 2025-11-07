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

type GetUserInfoResponse struct {
	ID         int64  `json:"id"`
	Username   string `json:"username"`
	Nickname   string `json:"nickname"`
	RegisterAt int64  `json:"register_at"`
}

// Do 执行获取用户信息的业务逻辑
func (r GetUserInfoRequest) Do() (any, error) {
	task.Run(func() {
		cache.Set(fmt.Sprintf("user:%d", r.User.ID), r.User, 24*time.Hour)
	})

	// 处理可空字段
	nickname := ""
	if r.User.Nickname != nil {
		nickname = *r.User.Nickname
	}

	registerAt := int64(0)
	if r.User.RegisterAt != nil {
		registerAt = r.User.RegisterAt.Unix()
	}

	return GetUserInfoResponse{
		ID:         r.User.ID,
		Username:   r.User.Username,
		Nickname:   nickname,
		RegisterAt: registerAt,
	}, nil
}

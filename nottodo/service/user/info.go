package user

import (
	"fmt"
	"time"

	"github.com/akagiyui/go-together/common/task"
	"github.com/akagiyui/go-together/nottodo/cache"
	"github.com/akagiyui/go-together/nottodo/repo"
)

type GetUserInfoRequest struct {
	User repo.User `context:"user"`
}

func (r GetUserInfoRequest) Do() (any, error) {
	task.Run(func() {
		cache.Set(fmt.Sprintf("user:%d", r.User.ID), r.User, 24*time.Hour)
	})
	return r.User, nil
}

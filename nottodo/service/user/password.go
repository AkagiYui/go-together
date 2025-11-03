package user

import (
	"github.com/akagiyui/go-together/nottodo/pkg"
	"github.com/akagiyui/go-together/nottodo/repo"
)

// ForceChangePassword 强制修改密码
type ForceChangePassword struct {
	UserID      int64
	NewPassword string
}

// Do 执行强制修改密码的业务逻辑
func (r ForceChangePassword) Do() (any, error) {
	password, err := pkg.HashPassword(r.NewPassword)
	if err != nil {
		return nil, err
	}
	return repo.UpdateUserPassword(r.UserID, password)
}

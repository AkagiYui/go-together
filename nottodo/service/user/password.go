package user

import (
	"github.com/akagiyui/go-together/nottodo/pkg"
	"github.com/akagiyui/go-together/nottodo/repo"
)

type ForceChangePassword struct {
	UserId      int64
	NewPassword string
}

func (r ForceChangePassword) Do() (any, error) {
	password, err := pkg.HashPassword(r.NewPassword)
	if err != nil {
		return nil, err
	}
	return repo.UpdateUserPassword(r.UserId, password)
}

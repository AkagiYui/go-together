package user

import (
	"errors"

	"github.com/akagiyui/go-together/arima/repo"
	"github.com/akagiyui/go-together/common/validation"
	"github.com/google/uuid"
)

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Name string `json:"name"`
}

// Validate 校验创建用户请求参数
func (r CreateUserRequest) Validate() error {
	return errors.Join(
		validation.Required(r.Name, "用户名"),
		validation.MaxLength(r.Name, 255, "用户名"),
	)
}

// Do 执行创建用户业务逻辑
func (r CreateUserRequest) Do() (any, error) {
	user := repo.User{
		Name:        r.Name,
		AccessKey:   "sk-" + uuid.New().String(),
		IsActive:    true,
		IsSuperuser: false,
	}
	return repo.CreateUser(user)
}

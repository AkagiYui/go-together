// Package user 提供用户相关的服务
package user

import (
	"errors"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/common/validation"
	"github.com/akagiyui/go-together/nottodo/pkg"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

// CreateUserRequest 创建用户请求
type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

// Validate 校验创建用户的请求参数
func (r CreateUserRequest) Validate() error {
	return errors.Join(
		validation.Required(r.Username, "用户名"),
		validation.Required(r.Password, "密码"),
	)
}

// Response 用户响应
type Response struct {
	ID int64 `json:"id"`
}

// NewUserResponse 创建用户响应
func NewUserResponse(user repo.User) Response {
	return Response{
		ID: user.ID,
	}
}

// Handle 处理创建用户的请求
func (r CreateUserRequest) Handle(ctx *rest.Context) {
	newUser, err := r.Do()
	if err != nil {
		ctx.SetResult(model.InternalError(err))
		return
	}
	ctx.SetResult(model.Success(NewUserResponse(newUser)))
}

// Do 执行创建用户的业务逻辑
func (r CreateUserRequest) Do() (repo.User, error) {
	password, err := pkg.HashPassword(r.Password)
	if err != nil {
		return repo.User{}, err
	}

	r.Password = password

	// 处理可空的 Nickname 字段
	var nickname *string
	if r.Nickname != "" {
		nickname = &r.Nickname
	}

	user, err := repo.CreateUser(repo.User{
		Username: r.Username,
		Password: r.Password,
		Nickname: nickname,
	})

	// RegisterAt 设置为 CreatedAt
	user.RegisterAt = &user.CreatedAt

	return user, err
}

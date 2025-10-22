package user

import (
	"errors"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/common/validation"
	"github.com/akagiyui/go-together/nottodo/pkg"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
	"github.com/jackc/pgx/v5/pgtype"
)

type CreateUserRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
	Nickname string `json:"nickname"`
}

func (r *CreateUserRequest) Validate() error {
	return errors.Join(
		validation.Required(r.Username, "用户名"),
		validation.Required(r.Password, "密码"),
	)
}

type UserResponse struct {
	ID int64 `json:"id"`
}

func NewUserResponse(user repo.User) UserResponse {
	return UserResponse{
		ID: user.ID,
	}
}

func (r CreateUserRequest) Handle(ctx *rest.Context) {
	newUser, err := r.Do()
	if err != nil {
		ctx.SetResult(model.InternalError(err))
		return
	}
	ctx.SetResult(model.Success(NewUserResponse(newUser)))
}

func (r CreateUserRequest) Do() (repo.User, error) {
	password, err := pkg.HashPassword(r.Password)
	if err != nil {
		return repo.User{}, err
	}

	r.Password = password
	return repo.CreateUser(repo.User{
		Username: r.Username,
		Password: r.Password,
		Nickname: pgtype.Text{String: r.Nickname, Valid: r.Nickname != ""},
	})
}

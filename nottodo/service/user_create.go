package service

import (
    "errors"

    "github.com/akagiyui/go-together/common/model"
    "github.com/akagiyui/go-together/common/validation"
    "github.com/akagiyui/go-together/nottodo/repo"
    "github.com/akagiyui/go-together/rest"
)

type CreateUserRequest struct {
    Username string `json:"username"`
    Nickname string `json:"nickname"`
    Password string `json:"password"`
}

func (r *CreateUserRequest) Validate() error {
    return errors.Join(
        validation.Required(r.Username, "用户名"),
        validation.Required(r.Password, "密码"),
        validation.MaxLength(r.Username, 50, "用户名"),
        validation.MaxLength(r.Nickname, 50, "昵称"),
    )
}

func (r *CreateUserRequest) Handle(ctx *rest.Context) {
    u, err := repo.CreateUser(ctx.Request.Context(), r.Username, r.Nickname, r.Password)
    if err != nil {
        ctx.SetResult(model.Error(model.INPUT_ERROR, err.Error()))
        return
    }
    // 不返回密码
    ctx.SetResult(model.Success(struct {
        ID       int64  `json:"id"`
        Username string `json:"username"`
        Nickname string `json:"nickname"`
        Created  string `json:"created_at"`
    }{
        ID:       u.ID,
        Username: u.Username,
        Nickname: u.Nickname,
        Created:  u.CreatedAt,
    }))
}

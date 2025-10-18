package service

import (
    "github.com/akagiyui/go-together/common/model"
    "github.com/akagiyui/go-together/common/validation"
    "github.com/akagiyui/go-together/nottodo/repo"
    "github.com/akagiyui/go-together/rest"
)

type DeleteUserRequest struct {
    ID int64 `path:"id"`
}

func (r *DeleteUserRequest) Validate() error {
    return validation.Positive(int(r.ID), "ID")
}

func (r *DeleteUserRequest) Handle(ctx *rest.Context) {
    ok, err := repo.DeleteUserByID(ctx.Request.Context(), r.ID)
    if err != nil {
        ctx.SetResult(model.InternalError())
        return
    }
    if !ok {
        ctx.SetResult(model.Error(model.NOT_FOUND, "User not found"))
        return
    }
    ctx.SetResult(model.Success(nil))
}

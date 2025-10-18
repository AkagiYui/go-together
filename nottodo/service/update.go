package service

import (
	"errors"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/common/validation"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

type UpdateTodoRequest struct {
	ID int `path:"id"`
	repo.Todo
}

// Validate 实现 Validator 接口，校验更新 Todo 的请求参数
func (r *UpdateTodoRequest) Validate() error {
	errs := make([]error, 0)
	errs = append(errs, validation.Positive(r.ID, "ID"))
	errs = append(errs, validation.MaxLength(r.Description, 500, "描述"))

	// 校验标题（如果提供）
	if r.Title != "" {
		errs = append(errs, validation.MaxLength(r.Title, 100, "标题"))
	}

	return errors.Join(errs...)
}

func (r *UpdateTodoRequest) Handle(ctx *rest.Context) {
	println("UpdateTodoRequest")
	oriTodo, ok, err := repo.GetTodoByID(ctx.Request.Context(), r.ID)
	if err != nil {
		ctx.SetResult(model.InternalError())
		return
	}
	if !ok {
		ctx.SetResult(model.Error(model.NOT_FOUND, "Todo not found"))
		return
	}

	if r.Todo.Title != "" {
		oriTodo.Title = r.Todo.Title
	}
	if r.Todo.Description != "" {
		oriTodo.Description = r.Todo.Description
	}

	if ok, err := repo.UpdateTodo(ctx.Request.Context(), oriTodo); err != nil {
		ctx.SetResult(model.InternalError())
		return
	} else if ok {
		ctx.SetResult(model.Success(oriTodo))
	} else {
		ctx.SetResult(model.InternalError())
	}
}

package service

import (
	"errors"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

type UpdateTodoRequest struct {
	ID int `path:"id"`
	repo.Todo
}

// Validate 实现 Validator 接口，校验更新 Todo 的请求参数
func (r *UpdateTodoRequest) Validate() error {
	if r.ID <= 0 {
		return errors.New("ID 必须大于 0")
	}
	if r.Title != "" && len(r.Title) > 100 {
		return errors.New("标题长度不能超过100个字符")
	}
	if len(r.Description) > 500 {
		return errors.New("描述长度不能超过500个字符")
	}
	return nil
}

func (r *UpdateTodoRequest) Handle(ctx *rest.Context) {
	println("UpdateTodoRequest")
	oriTodo, ok := repo.GetTodoByID(r.ID)
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

	if repo.UpdateTodo(oriTodo) {
		ctx.SetResult(model.Success(oriTodo))
	} else {
		ctx.SetResult(model.InternalError())
	}
}

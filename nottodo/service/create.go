package service

import (
	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

type CreateTodoRequest struct {
	repo.Todo
}

func (r *CreateTodoRequest) Handle(ctx *rest.Context) {
	println("CreateTodoRequest")
	// 验证必填字段
	if r.Title == "" {
		ctx.SetResult(model.Error(model.INPUT_ERROR, "Title is required"))
		return
	}

	// 创建新的 Todo
	newTodo := repo.Todo{
		Title:       r.Title,
		Description: r.Description,
		Completed:   r.Completed,
	}

	if repo.CreateTodo(newTodo) {
		ctx.SetResult(model.Success(newTodo))
	} else {
		ctx.SetResult(model.InternalError())
	}
}

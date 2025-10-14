package service

import (
	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

type UpdateTodoRequest struct {
	ID int `path:"id"`
	repo.Todo
}

func (r *UpdateTodoRequest) Handle(ctx *rest.Context) {
	println("UpdateTodoRequest")
	oriTodo, ok := repo.GetTodoByID(r.ID)
	if !ok {
		ctx.SetResult(model.Error(model.NOT_FOUND, "Todo not found"))
		return
	}

	if r.Todo.Title == "" {
		oriTodo.Title = oriTodo.Title
	}
	if r.Todo.Description == "" {
		oriTodo.Description = oriTodo.Description
	}
	if !r.Todo.Completed {
		oriTodo.Completed = oriTodo.Completed
	}

	if repo.UpdateTodo(oriTodo) {
		ctx.SetResult(model.Success(oriTodo))
	} else {
		ctx.SetResult(model.InternalError())
	}
}

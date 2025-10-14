package main

import (
	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/rest"
)

type UpdateTodoRequest struct {
	ID int `path:"id"`
	Todo
}

func (r *UpdateTodoRequest) Handle(ctx *rest.Context) {
	println("UpdateTodoRequest")
	for i, todo := range todos {
		if todo.ID == r.ID {
			oriTodo := todos[i]

			// 修整请求参数
			if r.Todo.Title == "" {
				r.Todo.Title = oriTodo.Title
			}
			r.Todo.ID = oriTodo.ID
			r.Todo.CreatedAt = oriTodo.CreatedAt

			todos[i] = r.Todo
			ctx.SetResult(model.Success(todos[i]))
			return
		}
	}
	ctx.SetResult(model.Error(model.NOT_FOUND, "Todo not found"))
}

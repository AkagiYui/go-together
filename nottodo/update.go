package main

import (
	"net/http"

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
			ctx.Result(model.Success(todos[i]))
			return
		}
	}
	ctx.Status(http.StatusNotFound)
	ctx.Result(model.Error(model.SUCCESS, "Todo not found"))
}

package main

import (
	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/rest"
)

type DeleteTodoRequest struct {
	ID int `path:"id"`
}

func (r *DeleteTodoRequest) Handle(ctx *rest.Context) {
	println("DeleteTodoRequest")
	for i, todo := range todos {
		if todo.ID == r.ID {
			todos = append(todos[:i], todos[i+1:]...)
			ctx.SetResult(model.Success(todo))
			return
		}
	}
	ctx.SetResult(model.Error(model.NOT_FOUND, "Todo not found"))
}

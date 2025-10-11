package main

import (
	"net/http"

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
			ctx.Result(model.Success(todo))
			return
		}
	}
	ctx.Status(http.StatusNotFound)
	ctx.Result(model.Error(model.SUCCESS, "Todo not found"))
}

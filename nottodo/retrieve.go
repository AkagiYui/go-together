package main

import (
	"net/http"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/rest"
)

type GetTodosRequest struct{}

func (r *GetTodosRequest) Handle(ctx *rest.Context) {
	println("GetTodosRequest")
	ctx.Result(model.Success(model.PageData{
		Total: len(todos),
		List:  todos,
	}))
}

type GetTodoByIDRequest struct {
	ID int `path:"id"`
}

func (r *GetTodoByIDRequest) Handle(ctx *rest.Context) {
	println("GetTodoByIDRequest")
	for _, todo := range todos {
		if todo.ID == r.ID {
			ctx.Result(model.Success(todo))
			return
		}
	}
	ctx.Status(http.StatusNotFound)
	ctx.Result(model.Error(model.NOT_FOUND, "Todo not found"))
}

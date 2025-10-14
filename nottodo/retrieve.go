package main

import (
	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/rest"
)

type GetTodosRequest struct{}

func (r *GetTodosRequest) Handle(ctx *rest.Context) {
	println("GetTodosRequest")
	ctx.SetResult(model.Success(model.PageData{
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
			ctx.SetResult(model.Success(todo))
			return
		}
	}
	ctx.SetResult(model.Error(model.NOT_FOUND, "Todo not found"))
}

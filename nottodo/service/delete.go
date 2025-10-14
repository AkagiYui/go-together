package service

import (
	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

type DeleteTodoRequest struct {
	ID int `path:"id"`
}

func (r *DeleteTodoRequest) Handle(ctx *rest.Context) {
	println("DeleteTodoRequest")
	if repo.DeleteTodo(r.ID) {
		ctx.SetResult(model.Success(nil))
	} else {
		ctx.SetResult(model.Error(model.NOT_FOUND, "Todo not found"))
	}
}

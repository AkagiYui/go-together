package service

import (
	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/common/validation"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

type DeleteTodoRequest struct {
	ID int `path:"id"`
}

// Validate 实现 Validator 接口，校验删除 Todo 的请求参数
func (r *DeleteTodoRequest) Validate() error {
	return validation.Positive(r.ID, "ID")
}

func (r *DeleteTodoRequest) Handle(ctx *rest.Context) {
	println("DeleteTodoRequest")
	if ok, err := repo.DeleteTodo(ctx.Request.Context(), r.ID); err != nil {
		ctx.SetResult(model.InternalError())
		return
	} else if ok {
		ctx.SetResult(model.Success(nil))
	} else {
		ctx.SetResult(model.Error(model.NOT_FOUND, "Todo not found"))
	}
}

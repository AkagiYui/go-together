package todo

import (
	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/common/validation"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

type DeleteTodoRequest struct {
	ID int64 `path:"id"`
}

// Validate 实现 Validator 接口，校验删除 Todo 的请求参数
func (r *DeleteTodoRequest) Validate() error {
	return validation.PositiveInt64(r.ID, "ID")
}

func (r *DeleteTodoRequest) Handle(ctx *rest.Context) {
	if err := repo.DeleteTodo(int64(r.ID)); err != nil {
		ctx.SetResult(model.Error(model.NOT_FOUND, "Todo not found"))
		return
	}
	ctx.SetResult(model.Success(nil))
}

package todo

import (
	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/common/validation"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

// DeleteTodoRequest 删除待办事项
// 根据 ID 删除指定的待办事项
type DeleteTodoRequest struct {
	ID int64 `path:"id"`
}

func (r DeleteTodoRequest) Validate() error {
	return validation.PositiveInt64(r.ID, "ID")
}

func (r DeleteTodoRequest) Handle(ctx *rest.Context) {
	if err := r.Do(); err != nil {
		ctx.SetResult(model.Error(model.NOT_FOUND, "Todo not found"))
		return
	}
	ctx.SetResult(model.Success(nil))
}

func (r DeleteTodoRequest) Do() error {
	return repo.DeleteTodo(r.ID)
}

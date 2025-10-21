package todo

import (
	"errors"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/common/validation"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

// UpdateTodoRequest 更新待办事项
// 根据 ID 更新待办事项的信息，支持部分更新
type UpdateTodoRequest struct {
	ID int64 `path:"id"`
	repo.Todo
}

func (r *UpdateTodoRequest) Validate() error {
	errs := make([]error, 0)
	errs = append(errs, validation.PositiveInt64(r.ID, "ID"))
	errs = append(errs, validation.MaxLength(r.Description.String, 500, "描述"))

	// 校验标题（如果提供）
	if r.Title != "" {
		errs = append(errs, validation.MaxLength(r.Title, 100, "标题"))
	}

	return errors.Join(errs...)
}

func (r *UpdateTodoRequest) Handle(ctx *rest.Context) {
	oriTodo, err := repo.GetTodoByID(r.ID)
	if err != nil {
		ctx.SetResult(model.Error(model.NOT_FOUND, "Todo not found"))
		return
	}

	if r.Todo.Title != "" {
		oriTodo.Title = r.Todo.Title
	}
	if r.Todo.Description.String != "" {
		oriTodo.Description = r.Todo.Description
	}

	if repo.UpdateTodo(oriTodo) {
		ctx.SetResult(model.Success(oriTodo))
	} else {
		ctx.SetResult(model.InternalError())
	}
}

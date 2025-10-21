package todo

import (
	"errors"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/common/validation"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

// CreateTodoRequest 创建待办事项
// 创建一个新的待办事项，需要提供标题和可选的描述
type CreateTodoRequest struct {
	repo.Todo
}

func (r *CreateTodoRequest) Validate() error {
	return errors.Join(
		validation.Required(r.Title, "标题"),
		validation.MaxLength(r.Title, 100, "标题"),
		validation.MaxLength(r.Description.String, 500, "描述"),
	)
}

func (r *CreateTodoRequest) Handle(ctx *rest.Context) {
	newTodo := repo.Todo{
		Title:       r.Title,
		Description: r.Description,
		Completed:   r.Completed,
	}

	newTodo, err := repo.CreateTodo(newTodo)
	if err != nil {
		ctx.SetResult(model.InternalError(err))
		return
	}
	ctx.SetResult(model.Success(newTodo))
}

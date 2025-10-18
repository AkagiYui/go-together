package service

import (
	"errors"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/common/validation"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

type CreateTodoRequest struct {
	repo.Todo
}

// Validate 实现 Validator 接口，校验创建 Todo 的请求参数
func (r *CreateTodoRequest) Validate() error {
	return errors.Join(
		validation.Required(r.Title, "标题"),
		validation.MaxLength(r.Title, 100, "标题"),
		validation.MaxLength(r.Description, 500, "描述"),
	)
}

func (r *CreateTodoRequest) Handle(ctx *rest.Context) {
	println("CreateTodoRequest")
	// 数据已通过校验，直接创建 Todo
	newTodo := repo.Todo{
		Title:       r.Title,
		Description: r.Description,
		Completed:   r.Completed,
	}

	if created, ok, err := repo.CreateTodo(ctx.Request.Context(), newTodo); err != nil {
		ctx.SetResult(model.InternalError())
		return
	} else if ok {
		ctx.SetResult(model.Success(created))
	} else {
		ctx.SetResult(model.InternalError())
	}
}

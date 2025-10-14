package service

import (
	"errors"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

type CreateTodoRequest struct {
	repo.Todo
}

// Validate 实现 Validator 接口，校验创建 Todo 的请求参数
func (r *CreateTodoRequest) Validate() error {
	if r.Title == "" {
		return errors.New("标题不能为空")
	}
	if len(r.Title) > 100 {
		return errors.New("标题长度不能超过100个字符")
	}
	if len(r.Description) > 500 {
		return errors.New("描述长度不能超过500个字符")
	}
	return nil
}

func (r *CreateTodoRequest) Handle(ctx *rest.Context) {
	println("CreateTodoRequest")
	// 数据已通过校验，直接创建 Todo
	newTodo := repo.Todo{
		Title:       r.Title,
		Description: r.Description,
		Completed:   r.Completed,
	}

	if repo.CreateTodo(newTodo) {
		ctx.SetResult(model.Success(newTodo))
	} else {
		ctx.SetResult(model.InternalError())
	}
}

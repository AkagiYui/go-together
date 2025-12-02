// Package todo 提供待办事项相关的服务
package todo

import (
	"errors"

	"github.com/akagiyui/go-together/common/validation"
	"github.com/akagiyui/go-together/nottodo/repo"
)

// CreateTodoRequest 创建待办事项
// 创建一个新的待办事项，需要提供标题和可选的描述
type CreateTodoRequest struct {
	repo.Todo
}

// Validate 校验创建待办事项的请求参数
func (r CreateTodoRequest) Validate() error {
	// 处理可空的 Description 字段
	description := ""
	if r.Description != nil {
		description = *r.Description
	}

	return errors.Join(
		validation.Required(r.Title, "标题"),
		validation.MaxLength(r.Title, 100, "标题"),
		validation.MaxLength(description, 500, "描述"),
	)
}

// Do 执行创建待办事项的业务逻辑
func (r CreateTodoRequest) Do() (any, error) {
	return repo.CreateTodo(repo.Todo{
		Title:       r.Title,
		Description: r.Description,
		Completed:   r.Completed,
	})
}

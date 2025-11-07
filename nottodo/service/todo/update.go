package todo

import (
	"database/sql"
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

// Validate 校验更新待办事项的请求参数
func (r UpdateTodoRequest) Validate() error {
	errs := make([]error, 0)
	errs = append(errs, validation.PositiveInt64(r.ID, "ID"))

	// 处理可空的 Description 字段
	if r.Description != nil {
		errs = append(errs, validation.MaxLength(*r.Description, 500, "描述"))
	}

	// 校验标题（如果提供）
	if r.Title != "" {
		errs = append(errs, validation.MaxLength(r.Title, 100, "标题"))
	}

	return errors.Join(errs...)
}

// Handle 处理更新待办事项的请求
func (r UpdateTodoRequest) Handle(ctx *rest.Context) {
	err := r.Do()
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			ctx.SetResult(model.Error(model.ErrNotFound, "Todo not found"))
			return
		}
		ctx.SetResult(model.InternalError(err))
		return
	}
	ctx.SetResult(model.Success(nil))
}

// Do 执行更新待办事项的业务逻辑
func (r UpdateTodoRequest) Do() error {
	oriTodo, err := repo.GetTodoByID(r.ID)
	if err != nil {
		return err
	}

	if r.Todo.Title != "" {
		oriTodo.Title = r.Todo.Title
	}
	// 处理可空的 Description 字段
	if r.Todo.Description != nil && *r.Todo.Description != "" {
		oriTodo.Description = r.Todo.Description
	}

	return repo.UpdateTodo(oriTodo)
}

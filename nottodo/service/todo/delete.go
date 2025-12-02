package todo

import (
	"github.com/akagiyui/go-together/common/validation"
	"github.com/akagiyui/go-together/nottodo/repo"
)

// DeleteTodoRequest 删除待办事项
// 根据 ID 删除指定的待办事项
type DeleteTodoRequest struct {
	ID int64 `path:"id"`
}

// Validate 校验删除待办事项的请求参数
func (r DeleteTodoRequest) Validate() error {
	return validation.PositiveInt64(r.ID, "ID")
}

// Do 执行删除待办事项的业务逻辑
func (r DeleteTodoRequest) Do() (any, error) {
	return nil, repo.DeleteTodo(r.ID)
}

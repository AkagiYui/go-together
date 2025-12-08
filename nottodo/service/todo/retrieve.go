package todo

import (
	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/common/validation"

	"github.com/akagiyui/go-together/nottodo/repo"
)

// GetTodosRequest 获取所有待办事项
// 返回所有待办事项的列表
type GetTodosRequest struct{}

// Do 执行获取所有待办事项的业务逻辑
func (GetTodosRequest) Do() (any, error) {
	list, total, err := repo.GetTodos()
	if err != nil {
		return nil, err
	}
	return model.Page(total, list), nil
}

// GetTodoByIDRequest 获取指定ID的待办事项
// 根据待办事项的 ID 获取其详细信息
type GetTodoByIDRequest struct {
	ID int64 `path:"id"`
}

// Validate 校验获取待办事项的请求参数
func (r GetTodoByIDRequest) Validate() error {
	return validation.PositiveInt64(r.ID, "ID")
}

// Do 执行获取指定待办事项的业务逻辑
func (r GetTodoByIDRequest) Do() (any, error) {
	return repo.GetTodoByID(r.ID)
}

package todo

import (
	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/common/validation"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

// GetTodosRequest 获取所有待办事项
// 返回所有待办事项的列表
type GetTodosRequest struct{}

func (r GetTodosRequest) Handle(ctx *rest.Context) {
	list, total, err := r.Do()
	if err != nil {
		ctx.SetResult(model.Error(model.INPUT_ERROR, err.Error()))
		return
	}
	ctx.SetResult(model.Success(model.Page(total, list)))
}

func (GetTodosRequest) Do() ([]repo.Todo, int64, error) {
	return repo.GetTodos()
}

// GetTodoByIDRequest 获取指定ID的待办事项
// 根据待办事项的 ID 获取其详细信息
type GetTodoByIDRequest struct {
	ID int64 `path:"id"`
}

func (r GetTodoByIDRequest) Validate() error {
	return validation.PositiveInt64(r.ID, "ID")
}

func (r GetTodoByIDRequest) Handle(ctx *rest.Context) {
	todo, err := r.Do()
	if err != nil {
		ctx.SetResult(model.Error(model.NOT_FOUND, "Todo not found"))
		return
	}
	ctx.SetResult(model.Success(todo))
}

func (r GetTodoByIDRequest) Do() (repo.Todo, error) {
	return repo.GetTodoByID(r.ID)
}

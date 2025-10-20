package todo

import (
	"fmt"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/common/validation"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

type GetTodosRequest struct{}

func (r *GetTodosRequest) Handle(ctx *rest.Context) {
	list, total, err := repo.GetTodos()
	fmt.Printf("list: %v\n", list)
	if err != nil {
		ctx.SetResult(model.Error(model.INPUT_ERROR, err.Error()))
		return
	}
	ctx.SetResult(model.Success(model.Page(total, list)))
}

type GetTodoByIDRequest struct {
	ID int64 `path:"id"`
}

// Validate 实现 Validator 接口，校验获取单个 Todo 的请求参数
func (r *GetTodoByIDRequest) Validate() error {
	return validation.PositiveInt64(r.ID, "ID")
}

func (r *GetTodoByIDRequest) Handle(ctx *rest.Context) {
	todo, err := repo.GetTodoByID(r.ID)
	if err != nil {
		ctx.SetResult(model.Error(model.NOT_FOUND, "Todo not found"))
		return
	}
	ctx.SetResult(model.Success(todo))
}

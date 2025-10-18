package service

import (
	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/common/validation"
	"github.com/akagiyui/go-together/nottodo/repo"
	"github.com/akagiyui/go-together/rest"
)

type GetTodosRequest struct{}

func (r *GetTodosRequest) Handle(ctx *rest.Context) {
	println("GetTodosRequest")
	list, total := repo.GetTodos()
	ctx.SetResult(model.Success(model.PageData{
		Total: total,
		List:  list,
	}))
}

type GetTodoByIDRequest struct {
	ID int `path:"id"`
}

// Validate 实现 Validator 接口，校验获取单个 Todo 的请求参数
func (r *GetTodoByIDRequest) Validate() error {
	return validation.Positive(r.ID, "ID")
}

func (r *GetTodoByIDRequest) Handle(ctx *rest.Context) {
	println("GetTodoByIDRequest")
	todo, ok := repo.GetTodoByID(r.ID)
	if ok {
		ctx.SetResult(model.Success(todo))
		return
	}
	ctx.SetResult(model.Error(model.NOT_FOUND, "Todo not found"))
}

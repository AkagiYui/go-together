package main

import (
	"time"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/rest"
)

type CreateTodoRequest struct {
	Todo
}

func (r *CreateTodoRequest) Handle(ctx *rest.Context) {
	println("CreateTodoRequest")
	// 验证必填字段
	if r.Title == "" {
		ctx.Result(model.Error(model.INPUT_ERROR, "Title is required"))
		return
	}

	// 创建新的 Todo
	newTodo := Todo{
		ID:          nextID,
		Title:       r.Title,
		Description: r.Description,
		Completed:   r.Completed,
		CreatedAt:   time.Now().Format("2006-01-02 15:04:05"),
	}

	nextID++
	todos = append(todos, newTodo)
	ctx.Result(model.Success(newTodo))
}

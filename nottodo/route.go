package main

import (
	"github.com/akagiyui/go-together/nottodo/middleware"
	service "github.com/akagiyui/go-together/nottodo/service/todo"
)

const comment = `🚀 Server starting on http://localhost:8080
📚 API Documentation:
GET    /v1/todos        - 获取所有Todo
GET    /v1/todos/{id}   - 获取指定ID的Todo
POST   /v1/todos        - 创建Todo
PUT    /v1/todos/{id}   - 更新指定ID的Todo
DELETE /v1/todos/{id}   - 删除指定ID的Todo`

func init() {
	v1 := s.Group("/v1")

	todoGroup := v1.Group("/todos", middleware.AuthMiddleware())
	{
		todoGroup.Get("", &service.GetTodosRequest{})
		todoGroup.Get("/{id}", &service.GetTodoByIDRequest{})
		todoGroup.Post("", &service.CreateTodoRequest{})
		todoGroup.Put("/{id}", &service.UpdateTodoRequest{})
		todoGroup.Delete("/{id}", &service.DeleteTodoRequest{})
	}

	println(comment)
}

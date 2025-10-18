package main

import (
	"github.com/akagiyui/go-together/nottodo/middleware"
	"github.com/akagiyui/go-together/nottodo/service"
)

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

	println("🚀 Server starting on http://localhost:8080")
	println("📚 API Documentation:")
	println("  GET    /v1/todos        - 获取所有Todo")
	println("  GET    /v1/todos/{id}   - 获取指定ID的Todo")
	println("  POST   /v1/todos        - 创建Todo")
	println("  PUT    /v1/todos/{id}   - 更新指定ID的Todo")
	println("  DELETE /v1/todos/{id}   - 删除指定ID的Todo")
}

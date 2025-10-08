package main

import (
	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/rest"
)

func main() {
	s := rest.NewServer()

	s.GETFunc("/healthz", func(ctx *rest.Context) {
		ctx.Result(model.Success("Hello, World!"))
	})

	todoGroup := s.Group("/todos")
	{
		todoGroup.GET("", &GetTodosRequest{})
		todoGroup.GET("/{id}", &GetTodoByIDRequest{})
		todoGroup.POST("", &CreateTodoRequest{})
		todoGroup.PUT("/{id}", &UpdateTodoRequest{})
		todoGroup.DELETE("/{id}", &DeleteTodoRequest{})
	}

	println("🚀 Server starting on http://localhost:8080")
	println("📚 API Documentation:")
	println("  GET    /todos        - 获取所有Todo")
	println("  GET    /todos/{id}   - 获取指定ID的Todo")
	println("  POST   /todos        - 创建Todo")
	println("  PUT    /todos/{id}   - 更新指定ID的Todo")
	println("  DELETE /todos/{id}   - 删除指定ID的Todo")

	if err := s.Run(":8080"); err != nil {
		panic(err)
	}
}

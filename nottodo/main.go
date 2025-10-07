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
	s.GET("/todos", &GetTodosRequest{})
	s.GET("/todos/{id}", &GetTodoByIDRequest{})
	s.POST("/todos", &CreateTodoRequest{})
	s.PUT("/todos/{id}", &UpdateTodoRequest{})
	s.DELETE("/todos/{id}", &DeleteTodoRequest{})

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

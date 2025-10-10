package main

import (
	"net/http"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/rest"
)

type TestRequest struct {
	Test string `query:"test"`
}

func (r *TestRequest) Handle(ctx *rest.Context) {
	println(r.Test)
}

func CORSMiddleware() rest.HandlerFunc {
	return func(ctx *rest.Context) {
		ctx.Response.Header("Access-Control-Allow-Origin", "*")
		ctx.Response.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		ctx.Response.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")
		ctx.Response.Header("Access-Control-Allow-Credentials", "true")
		ctx.Response.Header("Access-Control-Max-Age", "86400")

		if ctx.Request.Method == "OPTIONS" {
			ctx.Status(http.StatusNoContent)
			ctx.Result(nil)
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

func main() {
	s := rest.NewServer()

	s.UseFunc(CORSMiddleware())

	s.GETFunc("/healthz", func(ctx *rest.Context) {
		ctx.Set("test", "123\n")
		ctx.Result(model.Success("Hello, World!"))
	}, func(ctx *rest.Context) {
		println(ctx.Get("test"))
	})

	todoGroup := s.Group("/todos", func(ctx *rest.Context) {
		println("todos group")
	})
	todoGroup.Use(&TestRequest{})
	{
		todoGroup.GET("", &TestRequest{}, &GetTodosRequest{})
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

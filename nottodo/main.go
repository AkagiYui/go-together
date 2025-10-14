package main

import (
	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/nottodo/middleware"
	"github.com/akagiyui/go-together/nottodo/service"
	"github.com/akagiyui/go-together/rest"
)

func AuthMiddleware() rest.HandlerFunc {
	return func(ctx *rest.Context) {
		// 验证 token
		token := ctx.Request.Header.Get("Authorization")
		if token != "Bearer 123" {
			ctx.SetResult(model.Error(model.UNAUTHORIZED, "Unauthorized"))
			ctx.Abort()
			return
		}
	}
}

func main() {
	s := rest.NewServer()
	s.Debug = true

	// 设置全局校验错误处理器
	s.SetValidationErrorHandler(func(ctx *rest.Context, err error) {
		ctx.SetResult(model.Error(model.INPUT_ERROR, err.Error()))
	})

	s.UseFunc(middleware.CorsMiddleware(), middleware.TimeConsumeMiddleware())
	s.UseFunc(func(ctx *rest.Context) {
		ctx.Next()
		if obj, ok := ctx.Result.(model.GeneralResponse); ok {
			ctx.Status(model.HttpStatus(obj.Code))
		}
	})

	// 设置 404 处理器
	s.SetNotFoundHandlers(
		func(ctx *rest.Context) {
			ctx.Response.Header("X-Custom", "NotFound")
		},
		func(ctx *rest.Context) {
			ctx.SetResult(model.Error(model.NOT_FOUND, "Route not found"))
		},
	)

	s.GetFunc("/healthz", func(ctx *rest.Context) {})

	todoGroup := s.Group("/todos", AuthMiddleware())
	{
		todoGroup.Get("", &service.GetTodosRequest{})
		todoGroup.Get("/{id}", &service.GetTodoByIDRequest{})
		todoGroup.Post("", &service.CreateTodoRequest{})
		todoGroup.Put("/{id}", &service.UpdateTodoRequest{})
		todoGroup.Delete("/{id}", &service.DeleteTodoRequest{})
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

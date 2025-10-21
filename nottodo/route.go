package main

import (
	"github.com/akagiyui/go-together/nottodo/middleware"
	"github.com/akagiyui/go-together/nottodo/service/system"
	"github.com/akagiyui/go-together/nottodo/service/todo"
)

const comment = `🚀 Server starting on http://localhost:8080
📚 API Documentation:
GET    /v1/todos        - 获取所有Todo
GET    /v1/todos/{id}   - 获取指定ID的Todo
POST   /v1/todos        - 创建Todo
PUT    /v1/todos/{id}   - 更新指定ID的Todo
DELETE /v1/todos/{id}   - 删除指定ID的Todo`

func registerRoute() {
	v1 := s.Group("/v1")
	{
		todoGroup := v1.Group("/todos", middleware.AuthMiddleware())
		{
			todoGroup.Get("", &todo.GetTodosRequest{})
			todoGroup.Get("/{id}", &todo.GetTodoByIDRequest{})
			todoGroup.Post("", &todo.CreateTodoRequest{})
			todoGroup.Put("/{id}", &todo.UpdateTodoRequest{})
			todoGroup.Delete("/{id}", &todo.DeleteTodoRequest{})
		}

		systemGroup := v1.Group("/system")
		{
			settingGroup := systemGroup.Group("/settings")
			{
				settingGroup.Get("/is_allow_registration", &system.GetIsAllowRegistration{})
				settingGroup.Put("/is_allow_registration", &system.SetIsAllowRegistration{})
			}
		}
	}

	println(comment)
}

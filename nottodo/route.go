package main

import (
	"fmt"
	"strings"

	"github.com/akagiyui/go-together/nottodo/config"
	"github.com/akagiyui/go-together/nottodo/middleware"
	"github.com/akagiyui/go-together/nottodo/service/system"
	"github.com/akagiyui/go-together/nottodo/service/todo"
	"github.com/akagiyui/go-together/nottodo/service/user"
	"github.com/akagiyui/go-together/rest"
)

const comment = `🚀 Server starting on http://LISTEN`

func registerRoute() {
	cfg := config.GlobalConfig
	registerV1Route(s.Group("/v1"))
	println(strings.Replace(comment, "LISTEN", fmt.Sprintf("%s:%s", cfg.Host, cfg.Port), 1))
}

func registerV1Route(r *rest.RouteGroup) {
	r.Use(middleware.AuthMiddleware())

	requireAuthGroup := r.Group("", middleware.RequireAuth())
	{
		todoGroup := requireAuthGroup.Group("/todo")
		{
			todoGroup.Get("", rest.Service[todo.GetTodosRequest]())
			todoGroup.Get("/{id}", rest.Service[todo.GetTodoByIDRequest]())
			todoGroup.Post("", rest.Service[todo.CreateTodoRequest]())
			todoGroup.Put("/{id}", rest.Service[todo.UpdateTodoRequest]())
			todoGroup.Delete("/{id}", rest.Service[todo.DeleteTodoRequest]())
		}

		systemGroup := requireAuthGroup.Group("/system")
		{
			settingGroup := systemGroup.Group("/settings")
			{
				settingGroup.Get("/is_allow_registration", rest.Service[system.GetIsAllowRegistration]())
				settingGroup.Put("/is_allow_registration", rest.Service[system.SetIsAllowRegistration]())
			}
		}

		userGroup := requireAuthGroup.Group("/user")
		{
			userGroup.Get("/info", rest.Service[user.GetUserInfoRequest]())
			userGroup.Post("", rest.Service[user.CreateUserRequest]())
		}
	}

	anonymousGroup := r.Group("")
	{
		userGroup := anonymousGroup.Group("/user")
		{
			userGroup.Post("/token", rest.Service[user.GenerateTokenRequest]())
		}
	}

}

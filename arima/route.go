package main

import (
	"fmt"
	"strings"

	"github.com/akagiyui/go-together/arima/config"
	"github.com/akagiyui/go-together/arima/middleware"
	"github.com/akagiyui/go-together/arima/service/audio"
	"github.com/akagiyui/go-together/arima/service/system"
	"github.com/akagiyui/go-together/arima/service/user"
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

	// 需要认证的路由组
	requireAuthGroup := r.Group("", middleware.RequireAuth())
	{
		// 用户路由
		userGroup := requireAuthGroup.Group("/users")
		{
			userGroup.Get("/me", rest.Service[user.GetUserMeRequest]())
		}
	}

	// 需要超级用户权限的路由组
	requireSuperuserGroup := r.Group("", middleware.RequireAuth(), middleware.RequireSuperuser())
	{
		// 用户管理
		userGroup := requireSuperuserGroup.Group("/users")
		{
			userGroup.Post("", rest.Service[user.CreateUserRequest]())
		}

		// 音频路由
		audioGroup := requireSuperuserGroup.Group("/audio")
		{
			audioGroup.Get("", rest.Service[audio.ListAudioRequest]())
			audioGroup.Get("/origin", rest.Service[audio.ListOriginAudioRequest]())
			audioGroup.Get("/origin/{id}/url", rest.Service[audio.GetOriginAudioDownloadURLRequest]())
			audioGroup.Post("/origin", rest.Service[audio.UploadOriginAudioRequest]())
		}

		// 系统路由
		systemGroup := requireSuperuserGroup.Group("/system")
		{
			systemGroup.Get("", rest.Service[system.GetSystemInfoRequest]())
		}
	}
}

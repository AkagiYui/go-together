package main

import (
	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/nottodo/config"
	"github.com/akagiyui/go-together/nottodo/middleware"
	"github.com/akagiyui/go-together/rest"
)

var s *rest.Server = rest.NewServer()

func init() {
	cfg := config.GlobalConfig
	s.Debug = cfg.Mode == config.ModeDev

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

	s.GetFunc("/healthz", func(ctx *rest.Context) {
		ctx.SetResult(model.Success("OK"))
	})

	registerRoute()
}

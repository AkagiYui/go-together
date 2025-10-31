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
		ctx.SetResult(model.Error(model.ErrInputError, err.Error()))
	})

	// 设置全局中间件
	s.UseFunc(middleware.CorsMiddleware(), middleware.TimeConsumeMiddleware())

	// s.UseFunc(func(ctx *rest.Context) {
	// 	ctx.Next()
	// 	if obj, ok := ctx.Result.(model.GeneralResponse); ok {
	// 		ctx.Status(model.HttpStatus(model.BusinessCode(obj.Code)))

	// 		if !slices.Contains([]model.BusinessCode{model.ErrSuccess, model.ErrInternalError}, obj.Code) {
	// 			fmt.Printf("500: %s\n", obj.Message)
	// 		}
	// 	} else {
	// 		if ctx.StatusCode == http.StatusBadRequest {
	// 			ctx.SetResult(model.Error(model.ErrInputError, "Invalid request"))
	// 		}
	// 	}
	// })

	// 设置 404 处理器
	s.SetNotFoundHandlers(func(ctx *rest.Context) {
		ctx.SetResult(model.Error(model.ErrNotFound))
	})

	// 服务健康检查
	s.GetFunc("/healthz", func(ctx *rest.Context) {
		ctx.SetResult(model.Success())
	})

	// 注册业务路由
	registerRoute()
}

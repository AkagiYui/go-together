package main

import (
	"github.com/akagiyui/go-together/arima/config"
	"github.com/akagiyui/go-together/arima/middleware"
	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/common/object"
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
	if object.HasText(cfg.AllowOrigin) {
		s.UseFunc(middleware.CorsMiddleware(cfg.AllowOrigin))
	}
	s.UseFunc(middleware.TimeConsumeMiddleware())

	// 统一封装响应体并设置HTTP状态码
	s.UseFunc(middleware.ResponseWrapperMiddleware())

	// 设置 404 处理器
	s.SetNotFoundHandlers(func(ctx *rest.Context) {
		ctx.SetResult(model.Error(model.ErrNotFound))
	})

	// 服务健康检查
	s.GetFunc("/healthz", func(ctx *rest.Context) {
		ctx.SetResult(model.Success(GetBuildInfo()))
	})

	// 注册业务路由
	registerRoute()
}

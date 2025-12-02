package main

import (
	"fmt"
	"net/http"
	"slices"

	"github.com/akagiyui/go-together/common/model"
	"github.com/akagiyui/go-together/common/object"
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
	if object.HasText(cfg.AllowOrigin) {
		s.Use(middleware.CorsMiddleware(cfg.AllowOrigin))
	}
	s.Use(middleware.TimeConsumeMiddleware())

	// 统一封装响应体并设置HTTP状态码
	s.Use(func(ctx *rest.Context) {
		ctx.Next()

		// 检测 ctx.Status
		if ctx.Status != nil {
			if ctx.Status != model.ErrSuccess {
				businessCodeObj := ctx.Status.(model.BusinessCode)
				httpStatusCode := model.HTTPStatus(businessCodeObj)

				ctx.SetStatusCode(httpStatusCode)
				if httpStatusCode < 500 {
					ctx.SetResult(model.Error(businessCodeObj))
				} else {
					ctx.SetResult(model.InternalError())
					fmt.Printf("500: %s\n", businessCodeObj.Error())
				}
				return
			}
		}

		// 如果已经是 GeneralResponse,则无需再次封装
		if obj, ok := ctx.Result.(model.GeneralResponse); ok {
			businessCodeObj := model.BusinessCodeFromInt(obj.Code)
			ctx.SetStatusCode(model.HTTPStatus(businessCodeObj))

			if !slices.Contains([]model.BusinessCode{model.ErrSuccess, model.ErrInternalError}, businessCodeObj) {
				fmt.Printf("500: %s\n", obj.Message)
			}
		} else {
			if ctx.StatusCode == http.StatusBadRequest {
				ctx.SetResult(model.Error(model.ErrInputError, "Invalid request"))
			}
		}
	})

	// 设置 404 处理器
	s.SetNotFound(func(ctx *rest.Context) {
		ctx.SetResult(model.Error(model.ErrNotFound))
	})

	// 服务健康检查
	s.Get("/healthz", func(ctx *rest.Context) {
		ctx.SetResult(model.Success(GetBuildInfo()))
	})

	// 注册业务路由
	registerRoute()
}

package rest

import (
	"encoding/json"
	"net/http"
	"reflect"
)

// Service 将 ServiceHandlerInterface 类型转换为 HandlerFunc
// T: 处理器结构体类型
// PT: T 的指针类型，必须实现 ServiceHandlerInterface
//
// 使用示例:
//
//	router.Get("/users", rest.Service[GetUsersRequest]())
func Service[T any, PT interface {
	*T
	ServiceHandlerInterface
}]() HandlerFunc {
	var zero T
	t := reflect.TypeOf(zero)

	if t.Kind() != reflect.Struct {
		panic("rest.Service: type parameter must be a struct type")
	}

	return func(ctx *Context) {
		// 创建新实例
		handlerValue := reflect.New(t)
		handlerPtr := handlerValue.Interface().(PT)

		// 解析参数并注入字段
		needParseJSONBody, err := parseParams(ctx, handlerPtr)
		if err != nil {
			ctx.SetStatusCode(http.StatusBadRequest)
			ctx.SetResult("Failed to parse parameters: " + err.Error())
			return
		}

		// 解析 JSON 请求体
		if needParseJSONBody && ctx.BodyType == JSON && ctx.ContentLength > 0 {
			if err := json.Unmarshal(ctx.FillBody(), handlerPtr); err != nil {
				ctx.SetStatusCode(http.StatusBadRequest)
				ctx.SetResult("Invalid JSON format: " + err.Error())
				return
			}
		}

		// 执行校验
		if validator, ok := any(handlerPtr).(Validator); ok {
			if err := validator.Validate(); err != nil {
				if ctx.Server.validationErrorHandler != nil {
					ctx.Server.validationErrorHandler(ctx, err)
				} else {
					ctx.SetStatusCode(http.StatusBadRequest)
					ctx.SetResult("Validation failed: " + err.Error())
				}
				return
			}
		}

		// 调用 Do 方法
		result, err := handlerPtr.Do()
		ctx.SetResult(result)
		ctx.SetStatus(err)
	}
}

// Struct 将 HandlerInterface 类型转换为 HandlerFunc
// T: 处理器结构体类型
// PT: T 的指针类型，必须实现 HandlerInterface
//
// 使用示例:
//
//	router.Post("/users", rest.Struct[CreateUserHandler]())
func Struct[T any, PT interface {
	*T
	HandlerInterface
}]() HandlerFunc {
	var zero T
	t := reflect.TypeOf(zero)

	if t.Kind() != reflect.Struct {
		panic("rest.Struct: type parameter must be a struct type")
	}

	return func(ctx *Context) {
		// 创建新实例
		handlerValue := reflect.New(t)
		handlerPtr := handlerValue.Interface().(PT)

		// 解析参数并注入字段
		needParseJSONBody, err := parseParams(ctx, handlerPtr)
		if err != nil {
			ctx.SetStatusCode(http.StatusBadRequest)
			ctx.SetResult("Failed to parse parameters: " + err.Error())
			return
		}

		// 解析 JSON 请求体
		if needParseJSONBody && ctx.BodyType == JSON && ctx.ContentLength > 0 {
			if err := json.Unmarshal(ctx.FillBody(), handlerPtr); err != nil {
				ctx.SetStatusCode(http.StatusBadRequest)
				ctx.SetResult("Invalid JSON format: " + err.Error())
				return
			}
		}

		// 执行校验
		if validator, ok := any(handlerPtr).(Validator); ok {
			if err := validator.Validate(); err != nil {
				if ctx.Server.validationErrorHandler != nil {
					ctx.Server.validationErrorHandler(ctx, err)
				} else {
					ctx.SetStatusCode(http.StatusBadRequest)
					ctx.SetResult("Validation failed: " + err.Error())
				}
				return
			}
		}

		// 调用 Handle 方法
		handlerPtr.Handle(ctx)
	}
}


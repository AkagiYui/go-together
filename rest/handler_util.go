package rest

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
)

// runnersFromHandlers 将实现 HandlerInterface 的结构体类型转换为每次请求创建新实例并执行的 HandlerFunc 序列
func runnersFromHandlers(handlerTypes ...HandlerInterface) ([]HandlerFunc, error) {
	runners := make([]HandlerFunc, 0, len(handlerTypes))
	it := reflect.TypeOf((*HandlerInterface)(nil)).Elem()

	for _, handlerType := range handlerTypes {
		t := reflect.TypeOf(handlerType)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		// 确保实现了 HandlerInterface
		tt := t // 稳定副本
		if !reflect.PointerTo(tt).Implements(it) {
			return nil, ErrHandlerNotImplementHandlerInterface{}
		}

		runner := func(ctx *Context) {
			// 创建新的 HandlerInterface 实例
			handlerValue := reflect.New(tt)
			handlerInterface := handlerValue.Interface()
			handler, ok := handlerInterface.(HandlerInterface)
			if !ok {
				panic("Handler does not implement HandlerInterface")
			}

			// 解析 query/path/header 参数
			needParseBody, err := parseParams(ctx, handlerInterface)
			if err != nil {
				ctx.Status(http.StatusBadRequest)
				ctx.Result("Failed to parse parameters: " + err.Error())
				return
			}

			// 如果需要解析请求体，尝试解析 JSON 到结构体
			if needParseBody {
				if ctx.Body == nil {
					body, err := io.ReadAll(ctx.OriginalRequest.Body)
					if err != nil {
						ctx.Status(http.StatusBadRequest)
						ctx.Result("Failed to read request body")
						return
					}
					ctx.Body = body
				}

				contentType := strings.ToLower(strings.Trim(ctx.Request.Header.Get("Content-Type"), " "))
				if len(ctx.Body) > 0 {
					if strings.HasPrefix(contentType, "application/json") {
						if err := json.Unmarshal(ctx.Body, handlerInterface); err != nil {
							ctx.Status(http.StatusBadRequest)
							ctx.Result("Invalid JSON format: " + err.Error())
							return
						}
					}
				}
			}

			handler.Handle(ctx) // 调用 handler
		}

		runners = append(runners, runner)
	}

	return runners, nil
}

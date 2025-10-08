package rest

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strings"
)

type HandlerFunc func(*Context)

type HandlerInterface interface {
	Handle(*Context)
}

// HandlerFactory 用于创建 handler 实例的工厂
type HandlerFactory struct {
	Path        string
	Method      string
	RunnerChain []func(*Context)
}

// Handle 接受结构体类型，每次请求时创建新实例
func (g *RouteGroup) Handle(path string, method string, handlerTypes ...HandlerInterface) error {

	factory := HandlerFactory{
		Path:        path,
		Method:      method,
		RunnerChain: make([]func(*Context), len(handlerTypes)),
	}

	for i, handlerType := range handlerTypes {
		t := reflect.TypeOf(handlerType)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		// 确保实现了 HandlerInterface
		it := reflect.TypeOf((*HandlerInterface)(nil)).Elem()
		tt := t // 每次迭代的“稳定类型副本”
		if !reflect.PointerTo(tt).Implements(it) {
			return ErrHandlerNotImplementHandlerInterface{}
		}

		factory.RunnerChain[i] = func(ctx *Context) {
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

				contentType := strings.ToLower(strings.Trim(ctx.Header.Get("Content-Type"), " "))
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
	}

	g.Factories = append(g.Factories, factory)
	return nil
}

func (g *RouteGroup) HandleFunc(path string, method string, handlerFuncs ...HandlerFunc) {
	factory := HandlerFactory{
		Path:        path,
		Method:      method,
		RunnerChain: make([]func(*Context), len(handlerFuncs)),
	}

	for i, handlerFunc := range handlerFuncs {
		factory.RunnerChain[i] = handlerFunc
	}

	g.Factories = append(g.Factories, factory)
}

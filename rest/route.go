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
func (g *RouteGroup) Handle(path string, method string, handlerTypes ...HandlerInterface) {

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
		factory.RunnerChain[i] = func(ctx *Context) {
			// 创建新的 HandlerInterface 实例
			handlerValue := reflect.New(t)
			handlerInterface := handlerValue.Interface()

			// 确保实现了 HandlerInterface
			handler, ok := handlerInterface.(HandlerInterface)
			if !ok {
				ctx.Status(http.StatusInternalServerError)
				ctx.Result("Handler does not implement HandlerInterface")
				return
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
				body, err := io.ReadAll(ctx.OriginalRequest.Body)
				if err != nil {
					ctx.Status(http.StatusBadRequest)
					ctx.Result("Failed to read request body")
					return
				}
				ctx.Body = body

				contentType := strings.ToLower(strings.Trim(ctx.Header.Get("Content-Type"), " "))
				if len(body) > 0 {
					if strings.HasPrefix(contentType, "application/json") {
						if err := json.Unmarshal(body, handlerInterface); err != nil {
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

package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"runtime"
	"strconv"
)

type Server struct {
	RouteGroup       // 用于为用户组织路由
	Debug            bool
	flattenFactories []HandlerFactory // 最终注册的所有路由

	// 404 处理器
	notFoundHandlers []HandlerFunc
	notFoundNames    []string

	// 校验错误处理器
	validationErrorHandler func(*Context, error)
}

func NewServer() *Server {
	server := &Server{
		RouteGroup:       NewRouteGroup(nil, ""),
		Debug:            false,
		flattenFactories: nil,

		notFoundHandlers: nil,
		notFoundNames:    nil,

		validationErrorHandler: nil,
	}
	server.RouteGroup.server = server

	return server
}

// SetValidationErrorHandler 设置全局校验错误处理器
// 当 handler 实现了 Validator 接口且 Validate() 返回错误时，会调用此处理器
// 如果未设置，将使用默认的错误处理（返回 400 状态码和错误信息）
func (s *Server) SetValidationErrorHandler(handler func(*Context, error)) {
	s.validationErrorHandler = handler
}

// flattenFactories 递归地将路由组中的路由展开为一个列表
//
// preBasePath 上一级路由组的路径
// prePreRunnerChain 上一级路由组的前置 handler 链
// prePreRunnerNames 上一级路由组的前置 handler 名称链
func flattenFactories(group *RouteGroup, preBasePath string, prePreRunnerChain []HandlerFunc, prePreRunnerNames []string) []HandlerFactory {
	factories := make([]HandlerFactory, 0)                                   // 这一级路由组的所有路由
	thisBasePath := preBasePath + group.BasePath                             // 当前路由组的路径
	thisPreRunnerChain := append(prePreRunnerChain, group.PreRunnerChain...) // 当前路由组的前置 handler 链
	thisPreRunnerNames := append(prePreRunnerNames, group.PreRunnerNames...) // 当前路由组的前置 handler 名称链
	// 处理当前路由组的路由
	for _, factory := range group.Factories {
		factory.Path = thisBasePath + factory.Path // 上一级路由组的路径 + 当前路由组的路径 + 当前路由的路径

		// 合并当前路由组的前置 handler 链和当前路由的 handler 链
		newRunnerChain := make([]HandlerFunc, 0, len(thisPreRunnerChain)+len(factory.RunnerChain))
		newRunnerChain = append(newRunnerChain, thisPreRunnerChain...)
		newRunnerChain = append(newRunnerChain, factory.RunnerChain...)
		factory.RunnerChain = newRunnerChain

		// 合并当前路由组的前置 handler 名称链和当前路由的 handler 名称链
		newHandlerNames := make([]string, 0, len(thisPreRunnerNames)+len(factory.HandlerNames))
		newHandlerNames = append(newHandlerNames, thisPreRunnerNames...)
		newHandlerNames = append(newHandlerNames, factory.HandlerNames...)
		factory.HandlerNames = newHandlerNames

		factories = append(factories, factory)
	}
	// 递归处理子路由组
	for _, childGroup := range group.ChildGroups {
		factories = append(factories, flattenFactories(childGroup, thisBasePath, thisPreRunnerChain, thisPreRunnerNames)...)
	}
	return factories
}

func (s *Server) Run(addr string) error {
	mux := http.NewServeMux()
	registerRouteGroup(mux, &s.RouteGroup, s) // 处理所有注册的 handler

	if s.Debug {
		// 计算最长路径长度，用于对齐
		endpointMaxLen := 0
		for _, factory := range s.flattenFactories {
			if len(factory.Path) > endpointMaxLen {
				endpointMaxLen = len(factory.Path)
			}
		}

		// 输出所有注册的路由
		for _, factory := range s.flattenFactories {
			handlerCount := len(factory.RunnerChain)
			// 优先使用 HandlerNames 中的最后一个名称，如果没有则使用反射获取
			var lastHandlerName string
			if len(factory.HandlerNames) > 0 && len(factory.HandlerNames) == handlerCount {
				// HandlerNames 与 RunnerChain 长度一致时，使用存储的名称
				lastHandlerName = factory.HandlerNames[handlerCount-1]
			} else {
				// 否则使用反射获取函数名称
				lastHandlerName = runtime.FuncForPC(reflect.ValueOf(factory.RunnerChain[handlerCount-1]).Pointer()).Name()
			}
			// 使用格式化字符串实现左对齐
			fmt.Printf("[%7s] %-*s --> %s (%d handlers)\n", factory.Method, endpointMaxLen, factory.Path, lastHandlerName, handlerCount)
		}
	}

	return http.ListenAndServe(addr, mux)
}

func registerRouteGroup(mux *http.ServeMux, group *RouteGroup, server *Server) {
	factories := flattenFactories(group, "", make([]HandlerFunc, 0), make([]string, 0))
	server.flattenFactories = factories

	// 注册用户路由
	for _, factory := range factories {
		// 构建路由路径
		pattern := factory.Path
		if factory.Method != "" {
			pattern = fmt.Sprintf("%s %s", factory.Method, factory.Path)
		}

		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			ctx := NewContext(r, &w, server, factory.RunnerChain) // 创建上下文

			defer func() {
				if err := recover(); err != nil {
					fmt.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal Server Error"))
				}
			}()

			// dispatch request
			ctx.Next()

			// response
			if !ctx.responseAsStream {
				ctx.writeHeaders()
				server.writeResponse(w, ctx.Result, ctx)
			}
		})
	}

	// 注册 404 处理器（捕获所有未匹配的请求）
	if len(server.notFoundHandlers) > 0 {
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			handlers := make([]HandlerFunc, 0, len(server.PreRunnerChain)+len(server.notFoundHandlers))
			handlers = append(handlers, server.PreRunnerChain...)
			handlers = append(handlers, server.notFoundHandlers...)

			ctx := NewContext(r, &w, server, handlers)
			ctx.Status(http.StatusNotFound) // 默认设置 404 状态码
			defer func() {
				if err := recover(); err != nil {
					fmt.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal Server Error"))
				}
			}()

			// dispatch request
			ctx.Next()

			// response
			if !ctx.disableInternalResponse {
				ctx.writeHeaders()
				server.writeResponse(w, ctx.Result, ctx)
			}
		})
	}
}

// writeResponse 统一处理响应写入
func (s *Server) writeResponse(w http.ResponseWriter, result any, ctx *Context) {
	if result == nil {
		w.WriteHeader(ctx.statusCode)
		return
	}

	// 调用 w.Write 时，如果没有调用 WriteHeader，会自动调用 WriteHeader(200)
	// 在 w.WriteHeader 后，就不能再修改 Header 了

	// 判断类型
	switch result := result.(type) {
	case string:
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(ctx.statusCode)
		w.Write([]byte(result))
	case int:
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(ctx.statusCode)
		w.Write([]byte(strconv.Itoa(result)))
	default:
		b, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(ctx.statusCode)
		w.Write(b)
	}
}

// SetNotFoundHandlers 设置 404 处理器
func (s *Server) SetNotFoundHandlers(handlers ...HandlerFunc) {
	s.notFoundHandlers = handlers
	s.notFoundNames = make([]string, len(handlers))
	for i, f := range handlers {
		s.notFoundNames[i] = funcName(f)
	}
}

// SetNotFound 设置 404 处理器（支持 HandlerInterface）
func (s *Server) SetNotFound(handlers ...HandlerInterface) error {
	runners, names, err := runnersFromHandlers(handlers...)
	if err != nil {
		return err
	}
	s.notFoundHandlers = runners
	s.notFoundNames = names
	return nil
}

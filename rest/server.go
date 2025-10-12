package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
)

type Server struct {
	RouteGroup       // 用于开发者组织路由
	Debug            bool
	flattenFactories []HandlerFactory // 最终注册的所有路由
}

func NewServer() *Server {
	server := &Server{
		RouteGroup:       NewRouteGroup(nil, ""),
		Debug:            false,
		flattenFactories: nil,
	}
	server.RouteGroup.server = server
	return server
}

// flattenFactories 递归地将路由组中的路由展开为一个列表
//
// preBasePath 上一级路由组的路径
// prePreRunnerChain 上一级路由组的前置 handler 链
func flattenFactories(group *RouteGroup, preBasePath string, prePreRunnerChain []HandlerFunc) []HandlerFactory {
	factories := make([]HandlerFactory, 0)                                   // 这一级路由组的所有路由
	thisBasePath := preBasePath + group.BasePath                             // 当前路由组的路径
	thisPreRunnerChain := append(prePreRunnerChain, group.PreRunnerChain...) // 当前路由组的前置 handler 链
	// 处理当前路由组的路由
	for _, factory := range group.Factories {
		factory.Path = thisBasePath + factory.Path // 上一级路由组的路径 + 当前路由组的路径 + 当前路由的路径

		// 合并当前路由组的前置 handler 链和当前路由的 handler 链
		newRunnerChain := make([]HandlerFunc, 0, len(thisPreRunnerChain)+len(factory.RunnerChain))
		newRunnerChain = append(newRunnerChain, thisPreRunnerChain...)
		newRunnerChain = append(newRunnerChain, factory.RunnerChain...)
		factory.RunnerChain = newRunnerChain

		factories = append(factories, factory)
	}
	// 递归处理子路由组
	for _, childGroup := range group.ChildGroups {
		factories = append(factories, flattenFactories(childGroup, thisBasePath, thisPreRunnerChain)...)
	}
	return factories
}

func (s *Server) Run(addr string) error {
	mux := http.NewServeMux()
	registerRouteGroup(mux, &s.RouteGroup, s) // 处理所有注册的 handler

	if s.Debug {
		// 输出所有注册的路由
		for _, factory := range s.flattenFactories {
			fmt.Printf("[%7s] %s\n", factory.Method, factory.Path)
		}
	}

	return http.ListenAndServe(addr, mux)
}

func registerRouteGroup(mux *http.ServeMux, group *RouteGroup, server *Server) {
	factories := flattenFactories(group, "", make([]HandlerFunc, 0))
	server.flattenFactories = factories
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
				server.writeResponse(w, ctx.result, ctx)
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

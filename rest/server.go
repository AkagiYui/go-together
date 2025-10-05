package rest

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strconv"
)

type Server struct {
	RouteGroup
}

func NewServer() *Server {
	return &Server{
		RouteGroup: RouteGroup{
			Factories: make([]HandlerFactory, 0),
			BasePath:  "",
		},
	}
}

func (s *Server) Run(addr string) error {
	mux := http.NewServeMux()

	// 处理所有注册的处理器
	for _, factory := range s.Factories {

		pattern := factory.Path
		if factory.Method != "" {
			pattern = factory.Method + " " + pattern
		}

		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			// 读取请求体
			body, err := io.ReadAll(r.Body)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Failed to read request body"))
				return
			}

			// 创建新的处理器实例
			handlerValue := reflect.New(factory.HandlerType)
			handlerInterface := handlerValue.Interface()

			// 如果请求体不为空，尝试解析 JSON 到结构体
			// todo 除了json tag，还支持form/query/path tag
			if len(body) > 0 {
				if err := json.Unmarshal(body, handlerInterface); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("Invalid JSON format: " + err.Error()))
					return
				}
			}

			// 确保实现了 HandlerInterface
			if handler, ok := handlerInterface.(HandlerInterface); ok {
				ctx := &Context{
					Method: r.Method,
					Path:   r.URL.Path,
					Body:   body,
				}
				result := handler.Handle(ctx)
				s.writeResponse(w, result)
			} else {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Handler does not implement HandlerInterface"))
			}
		})
	}

	return http.ListenAndServe(addr, mux)
}

// writeResponse 统一处理响应写入
func (s *Server) writeResponse(w http.ResponseWriter, result any) {
	if result == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// 设置 Content-Type
	w.Header().Set("Content-Type", "application/json")

	// 判断类型
	switch result := result.(type) {
	case string:
		w.Write([]byte(result))
	case int:
		w.Write([]byte(strconv.Itoa(result)))
	default:
		b, err := json.Marshal(result)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			return
		}
		w.Write(b)
	}
}

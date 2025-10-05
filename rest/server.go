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
		// 构建路由路径
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

			// 确保实现了 HandlerInterface
			handler, ok := handlerInterface.(HandlerInterface)
			if !ok {
				w.WriteHeader(http.StatusInternalServerError)
				w.Write([]byte("Handler does not implement HandlerInterface"))
				return
			}

			// 如果请求体不为空，尝试解析 JSON 到结构体
			if len(body) > 0 {
				if err := json.Unmarshal(body, handlerInterface); err != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("Invalid JSON format: " + err.Error()))
					return
				}
			}

			// 解析 query 和 path 参数
			if err := s.parseParams(r, handlerInterface); err != nil {
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte("Failed to parse parameters: " + err.Error()))
				return
			}

			ctx := &Context{
				Method: r.Method,
				Path:   r.URL.Path,
				Body:   body,
			}
			result := handler.Handle(ctx)
			s.writeResponse(w, result)
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

// parseParams 解析URL查询参数和路径参数到结构体字段
func (s *Server) parseParams(r *http.Request, handlerInterface interface{}) error {
	handlerValue := reflect.ValueOf(handlerInterface)
	if handlerValue.Kind() == reflect.Ptr {
		handlerValue = handlerValue.Elem()
	}
	handlerType := handlerValue.Type()

	queryValues := r.URL.Query()

	for i := 0; i < handlerType.NumField(); i++ {
		field := handlerType.Field(i)
		fieldValue := handlerValue.Field(i)

		// 检查字段是否可设置
		if !fieldValue.CanSet() {
			continue
		}

		// 处理 query tag
		if queryTag := field.Tag.Get("query"); queryTag != "" {
			if queryParam := queryValues.Get(queryTag); queryParam != "" {
				if err := s.setFieldValue(fieldValue, queryParam); err != nil {
					return err
				}
			}
		}

		// 处理 path tag
		if pathTag := field.Tag.Get("path"); pathTag != "" {
			if pathParam := r.PathValue(pathTag); pathParam != "" {
				if err := s.setFieldValue(fieldValue, pathParam); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// setFieldValue 根据字段类型设置值
func (s *Server) setFieldValue(fieldValue reflect.Value, value string) error {
	switch fieldValue.Kind() {
	case reflect.String:
		fieldValue.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetUint(uintVal)
	case reflect.Float32, reflect.Float64:
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		fieldValue.SetFloat(floatVal)
	case reflect.Bool:
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		fieldValue.SetBool(boolVal)
	default:
		// 对于不支持的类型，暂时跳过
		return nil
	}
	return nil
}

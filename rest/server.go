package rest

import (
	"encoding/json"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
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
			ctx := &Context{
				Method: r.Method,

				Endpoint: r.URL.Path,
				URI:      r.RequestURI,

				Body:          nil,
				URL:           *r.URL,
				Host:          r.Host,
				RemoteAddr:    r.RemoteAddr,
				ContentLength: r.ContentLength,

				Params: make(map[string]string),
				Query:  make(map[string]string),
				Header: make(map[string]string),
				Form:   make(map[string]string),

				header:     r.Header,
				statusCode: http.StatusOK,
			}

			println("URL.Path", r.URL.Path)
			println("RequestURI", r.RequestURI)

			var result any

			if factory.IsFunc {
				// 处理函数处理器
				result = factory.HandlerFunc(ctx)
			} else {
				// 处理结构体处理器
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

				// 解析 query 和 path 和 header 参数
				needParseBody, err := s.parseParams(r, handlerInterface)
				if err != nil {
					w.WriteHeader(http.StatusBadRequest)
					w.Write([]byte("Failed to parse parameters: " + err.Error()))
					return
				}

				// 如果请求体不为空，尝试解析 JSON 到结构体
				if needParseBody {
					// 读取请求体
					body, err := io.ReadAll(r.Body)
					if err != nil {
						w.WriteHeader(http.StatusBadRequest)
						w.Write([]byte("Failed to read request body"))
						return
					}
					ctx.Body = body
					if len(body) > 0 {
						if strings.HasPrefix(r.Header.Get("Content-Type"), "application/json") {
							if err := json.Unmarshal(body, handlerInterface); err != nil {
								w.WriteHeader(http.StatusBadRequest)
								w.Write([]byte("Invalid JSON format: " + err.Error()))
								return
							}
						}
					}
				}

				result = handler.Handle(ctx)
			}

			s.writeResponse(w, result, ctx)
		})
	}

	return http.ListenAndServe(addr, mux)
}

// writeResponse 统一处理响应写入
func (s *Server) writeResponse(w http.ResponseWriter, result any, ctx *Context) {
	if result == nil {
		w.WriteHeader(ctx.statusCode)
		return
	}

	contentType := "application/json"

	// 判断类型
	switch result := result.(type) {
	case string:
		contentType = "text/plain"
		w.Write([]byte(result))
	case int:
		contentType = "text/plain"
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

	w.Header().Set("Content-Type", contentType)
}

// parseParams 解析query参数和path参数和header参数到结构体字段
func (s *Server) parseParams(r *http.Request, handlerInterface interface{}) (needParseBody bool, err error) {
	needParseBody = false
	handlerValue := reflect.ValueOf(handlerInterface)
	if handlerValue.Kind() == reflect.Ptr {
		handlerValue = handlerValue.Elem()
	}
	handlerType := handlerValue.Type()

	queryValues := r.URL.Query()
	headers := r.Header

	for i := 0; i < handlerType.NumField(); i++ {
		field := handlerType.Field(i)
		fieldValue := handlerValue.Field(i)

		if !needParseBody {
			needParseBody = field.Tag.Get("json") != "" || field.Tag.Get("form") != "" || field.Tag.Get("body") != ""
		}

		// 检查字段是否可设置
		if !fieldValue.CanSet() {
			continue
		}

		// 处理 query tag
		if queryTag := field.Tag.Get("query"); queryTag != "" {
			if queryParam := queryValues.Get(queryTag); queryParam != "" {
				if err = s.setFieldValue(fieldValue, queryParam); err != nil {
					return
				}
			}
		}

		// 处理 path tag
		if pathTag := field.Tag.Get("path"); pathTag != "" {
			if pathParam := r.PathValue(pathTag); pathParam != "" {
				if err = s.setFieldValue(fieldValue, pathParam); err != nil {
					return
				}
			}
		}

		// 处理 header tag
		if headerTag := field.Tag.Get("header"); headerTag != "" {
			if headerValue := headers.Get(headerTag); headerValue != "" {
				if err = s.setFieldValue(fieldValue, headerValue); err != nil {
					return
				}
			}
		}
	}

	return
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

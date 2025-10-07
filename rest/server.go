package rest

import (
	"encoding/json"
	"fmt"
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
		pattern := s.RouteGroup.BasePath + factory.Path
		if factory.Method != "" {
			pattern = factory.Method + " " + pattern
		}

		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			ctx := NewContext(r, &w) // 创建上下文

			defer func() {
				if err := recover(); err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal Server Error"))
				}
			}()

			if factory.IsFunc {
				factory.HandlerFunc(ctx) // 调用函数 handler
			} else {
				// 处理结构体 handler
				handlerValue := reflect.New(factory.HandlerType) // 创建新的处理器实例
				handlerInterface := handlerValue.Interface()

				handler, ok := handlerInterface.(HandlerInterface) // 确保实现了 HandlerInterface
				if !ok {
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Handler does not implement HandlerInterface"))
					return
				}

				// 解析 query 和 path 和 header 参数
				needParseBody, err := parseParams(ctx, handlerInterface)
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

					contentType := strings.ToLower(strings.Trim(r.Header.Get("Content-Type"), " "))
					if len(body) > 0 {
						if strings.HasPrefix(contentType, "application/json") {
							if err := json.Unmarshal(body, handlerInterface); err != nil {
								w.WriteHeader(http.StatusBadRequest)
								w.Write([]byte("Invalid JSON format: " + err.Error()))
								return
							}
						}
					}
				}

				handler.Handle(ctx)
			}

			s.writeResponse(w, ctx.result, ctx)
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
func parseParams(ctx *Context, handlerInterface interface{}) (needParseBody bool, err error) {
	needParseBody = false
	handlerValue := reflect.ValueOf(handlerInterface)
	if handlerValue.Kind() == reflect.Ptr {
		handlerValue = handlerValue.Elem()
	}
	handlerType := handlerValue.Type()

	pathParams := ctx.PathParams
	queryValues := ctx.Query
	headers := ctx.Header

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
				if err = setFieldValue(fieldValue, queryParam); err != nil {
					return
				}
			}
		}

		// 处理 path tag
		if pathTag := field.Tag.Get("path"); pathTag != "" {
			if pathParam, ok := pathParams[pathTag]; ok && pathParam != "" {
				if err = setFieldValue(fieldValue, pathParam); err != nil {
					return
				}
			}
		}

		// 处理 header tag
		if headerTag := field.Tag.Get("header"); headerTag != "" {
			if headerValue := headers.Get(headerTag); headerValue != "" {
				if err = setFieldValue(fieldValue, headerValue); err != nil {
					return
				}
			}
		}
	}

	return
}

// setFieldValue 根据字段类型设置值
func setFieldValue(fieldValue reflect.Value, value string) error {
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
		fmt.Printf("Unsupported field type: %s for value: %s\n", fieldValue.Kind(), value)
		return nil
	}
	return nil
}

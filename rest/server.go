package rest

import (
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
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

// FlattenFactories 递归地将路由组中的路由展开为一个列表
//
// preBasePath 上一级路由组的路径
// prePreRunnerChain 上一级路由组的前置 handler 链
func FlattenFactories(group *RouteGroup, preBasePath string, prePreRunnerChain []HandlerFunc) []HandlerFactory {
	factories := make([]HandlerFactory, 0)                                   // 这一级路由组的所有路由
	thisBasePath := preBasePath + group.BasePath                             // 当前路由组的路径
	thisPreRunnerChain := append(prePreRunnerChain, group.PreRunnerChain...) // 当前路由组的前置 handler 链
	// 处理当前路由组的路由
	for _, factory := range group.Factories {
		factory.Path = thisBasePath + factory.Path                               // 上一级路由组的路径 + 当前路由组的路径 + 当前路由的路径
		factory.RunnerChain = append(thisPreRunnerChain, factory.RunnerChain...) // 上一级路由组的前置 handler 链 + 当前路由组的前置 handler 链 + 当前路由的 handler 链
		factories = append(factories, factory)
	}
	// 递归处理子路由组
	for _, childGroup := range group.ChildGroups {
		factories = append(factories, FlattenFactories(childGroup, thisBasePath, thisPreRunnerChain)...)
	}
	return factories
}

func (s *Server) Run(addr string) error {
	mux := http.NewServeMux()
	registerRouteGroup(mux, &s.RouteGroup, s) // 处理所有注册的 handler

	if s.Debug {
		// 输出所有注册的路由
		for _, factory := range s.flattenFactories {
			fmt.Printf("[%s] %s\n", factory.Method, factory.Path)
		}
	}

	return http.ListenAndServe(addr, mux)
}

func registerRouteGroup(mux *http.ServeMux, group *RouteGroup, server *Server) {
	// for _, factory := range group.Factories {
	// 	// 构建路由路径
	// 	pattern := group.BasePath + factory.Path
	// 	if factory.Method != "" {
	// 		pattern = factory.Method + " " + pattern
	// 	}

	// 	mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
	// 		ctx := NewContext(r, &w, server) // 创建上下文

	// 		defer func() {
	// 			if err := recover(); err != nil {
	// 				fmt.Println(err)
	// 				w.WriteHeader(http.StatusInternalServerError)
	// 				w.Write([]byte("Internal Server Error"))
	// 			}
	// 		}()

	// 		// dispatch request
	// 		ctx.runnerChain = factory.RunnerChain
	// 		ctx.Next()
	// 		for key, values := range ctx.Response.Headers {
	// 			for _, value := range values {
	// 				w.Header().Add(key, value)
	// 			}
	// 		}

	// 		server.writeResponse(w, ctx.result, ctx)
	// 	})
	// }

	// for _, childGroup := range group.ChildGroups {
	// 	registerRouteGroup(mux, childGroup, server)
	// }

	factories := FlattenFactories(group, "", make([]HandlerFunc, 0))
	server.flattenFactories = factories
	for _, factory := range factories {
		// 构建路由路径
		pattern := factory.Path
		if factory.Method != "" {
			pattern = fmt.Sprintf("%s %s", factory.Method, factory.Path)
		}

		mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			ctx := NewContext(r, &w, server) // 创建上下文

			defer func() {
				if err := recover(); err != nil {
					fmt.Println(err)
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte("Internal Server Error"))
				}
			}()

			// dispatch request
			ctx.runnerChain = factory.RunnerChain
			ctx.Next()
			for key, values := range ctx.Response.Headers {
				for _, value := range values {
					w.Header().Add(key, value)
				}
			}

			server.writeResponse(w, ctx.result, ctx)
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

// parseParams 解析query参数和path参数和header参数到结构体字段
func parseParams(ctx *Context, handlerInterface interface{}) (needParseBody bool, err error) {
	handlerValue := reflect.ValueOf(handlerInterface)
	if handlerValue.Kind() == reflect.Ptr {
		handlerValue = handlerValue.Elem()
	}
	needParseBody, err = parseStructFields(handlerValue, ctx)
	return
}

// parseStructFields 递归解析结构体字段
func parseStructFields(structValue reflect.Value, ctx *Context) (needParseBody bool, err error) {
	pathParams := ctx.PathParams
	queryValues := ctx.Query
	headers := ctx.Request.Header

	if structValue.Kind() == reflect.Ptr {
		if structValue.IsNil() {
			return false, nil
		}
		structValue = structValue.Elem()
	}

	if structValue.Kind() != reflect.Struct {
		return false, nil
	}

	structType := structValue.Type()

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)

		// 检查是否需要解析请求体
		needParseBody = needParseBody || field.Tag.Get("json") != "" || field.Tag.Get("form") != "" || field.Tag.Get("body") != ""

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

		// 递归处理嵌套结构体
		fieldType := field.Type
		if fieldType.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldType.Elem())) // 为nil的指针字段创建新实例
			}
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() == reflect.Struct {
			childNeedParseBody, childErr := parseStructFields(fieldValue, ctx)
			if childErr != nil {
				return needParseBody, childErr
			}
			needParseBody = needParseBody || childNeedParseBody
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

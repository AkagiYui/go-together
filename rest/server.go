package rest

import (
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
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
	factories := FlattenFactories(group, "", make([]HandlerFunc, 0))
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
	return parseStructFields(handlerValue, ctx)
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

		// 检查字段是否可设置
		if !fieldValue.CanSet() {
			continue
		}

		// 处理 query tag
		if queryTag := field.Tag.Get("query"); queryTag != "" {
			if queryParam, ok := queryValues[queryTag]; ok {
				if err = setFieldValue(fieldValue, queryParam...); err != nil {
					return
				}
			}
			continue
		}

		// 处理 path tag
		if pathTag := field.Tag.Get("path"); pathTag != "" {
			if pathParam, ok := pathParams[pathTag]; ok && pathParam != "" {
				if err = setFieldValue(fieldValue, pathParam); err != nil {
					return
				}
			}
			continue
		}

		// 处理 header tag
		if headerTag := field.Tag.Get("header"); headerTag != "" {
			if headerValue, ok := headers[textproto.CanonicalMIMEHeaderKey(headerTag)]; ok {
				if err = setFieldValue(fieldValue, headerValue...); err != nil {
					return
				}
			}
			continue
		}

		// 处理 json tag
		if jsonTag := field.Tag.Get("json"); jsonTag != "" && ctx.BodyType == Json {
			needParseBody = true // 交给 json.Unmarshal 处理
			continue
		}

		// 处理 form tag
		if formTag := field.Tag.Get("form"); formTag != "" {
			switch ctx.BodyType {
			case EncodeUrl:
				// 解析表单
				ctx.OriginalRequest.ParseForm()
				form := ctx.OriginalRequest.PostForm
				if form == nil {
					return false, nil
				}
				if formValue, ok := form[formTag]; ok {
					if err = setFieldValue(fieldValue, formValue...); err != nil {
						return
					}
				}
			case FormData:
				ctx.OriginalRequest.ParseMultipartForm(32 << 20) // 32MB
				form := ctx.OriginalRequest.MultipartForm
				if form == nil {
					return false, nil
				}

				// 处理普通表单字段
				notFileFieldsMap := form.Value
				if notFileValues, ok := notFileFieldsMap[formTag]; ok && len(notFileValues) > 0 {
					if err = setFieldValue(fieldValue, notFileValues...); err != nil {
						return
					}
				}

				// 处理文件
				fileFieldsMap := form.File
				if fileFields, ok := fileFieldsMap[formTag]; ok && len(fileFields) > 0 {
					if err = serFileFieldValue(fieldValue, fileFields...); err != nil {
						return
					}
				}
			}
			continue
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

func serFileFieldValue(fieldValue reflect.Value, fileHeader ...*multipart.FileHeader) error {
	// 判断字段类型
	// 如果是 *multipart.FileHeader，直接设置
	if fieldValue.Type() == reflect.TypeOf(&multipart.FileHeader{}) {
		fieldValue.Set(reflect.ValueOf(fileHeader[0]))
		return nil
	}
	// 如果是 []byte ，读取文件内容并设置
	if fieldValue.Type() == reflect.TypeOf([]byte{}) {
		file, err := fileHeader[0].Open()
		if err != nil {
			return err
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		fieldValue.Set(reflect.ValueOf(content))
		return nil
	}
	// 如果是 string ，使用 utf-8 解码 []byte 并设置
	if fieldValue.Kind() == reflect.String {
		file, err := fileHeader[0].Open()
		if err != nil {
			return err
		}
		defer file.Close()
		content, err := io.ReadAll(file)
		if err != nil {
			return err
		}
		fieldValue.SetString(string(content))
		return nil
	}
	return nil
}

// setFieldValue 根据字段类型设置值
func setFieldValue(fieldValue reflect.Value, values ...string) error {
	if len(values) == 0 {
		return nil
	}

	switch fieldValue.Kind() {
	case reflect.Slice:
		// 创建新的切片
		newSlice := reflect.MakeSlice(fieldValue.Type(), len(values), len(values))

		// 为每个元素设置值
		for i, value := range values {
			elemValue := newSlice.Index(i)
			if err := setScalarValue(elemValue, value); err != nil {
				return err
			}
		}

		// 设置切片
		fieldValue.Set(newSlice)

	default:
		// 对于非切片类型，只使用第一个值
		return setScalarValue(fieldValue, values[0])
	}
	return nil
}

// setScalarValue 设置标量值
func setScalarValue(fieldValue reflect.Value, value string) error {
	switch fieldValue.Kind() {
	case reflect.String: // 字符串
		fieldValue.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64: // 整数
		intVal, err := strconv.ParseInt(value, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetInt(intVal)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64: // 无符号整数
		uintVal, err := strconv.ParseUint(value, 10, 64)
		if err != nil {
			return err
		}
		fieldValue.SetUint(uintVal)
	case reflect.Float32, reflect.Float64: // 浮点数
		floatVal, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return err
		}
		fieldValue.SetFloat(floatVal)
	case reflect.Bool: // 布尔值
		boolVal, err := strconv.ParseBool(value)
		if err != nil {
			return err
		}
		fieldValue.SetBool(boolVal)
	default:
		fmt.Printf("Unsupported field type: %s for value: %s\n", fieldValue.Kind(), value)
		return nil
	}
	return nil
}

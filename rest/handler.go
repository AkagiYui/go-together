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

type BodyType int
type BodyFieldMap map[BodyType]map[string]string

const (
	Nil BodyType = iota
	EncodeUrl
	Json
	FormData
)

// runnersFromHandlers 将实现 HandlerInterface 的结构体类型转换为每次请求创建新实例并执行的 HandlerFunc 序列
func runnersFromHandlers(handlerTypes ...HandlerInterface) ([]HandlerFunc, error) {
	runners := make([]HandlerFunc, 0, len(handlerTypes))
	it := reflect.TypeOf((*HandlerInterface)(nil)).Elem()

	for _, handlerType := range handlerTypes {
		t := reflect.TypeOf(handlerType)
		if t.Kind() == reflect.Ptr {
			t = t.Elem()
		}

		// 确保实现了 HandlerInterface
		if !reflect.PointerTo(t).Implements(it) {
			return nil, ErrHandlerNotImplementHandlerInterface{}
		}

		// handler 仅处理该 handler 需要的数据，所有 handler 共用的数据请在 Context 中处理
		runner := func(ctx *Context) {
			// 创建新的 HandlerInterface 实例
			handlerValue := reflect.New(t)
			handlerInterface := handlerValue.Interface()
			handler, ok := handlerInterface.(HandlerInterface)
			if !ok {
				panic("Handler does not implement HandlerInterface")
			}

			// 解析 query/path/header 参数
			needParseJsonBody, err := parseParams(ctx, handlerInterface)
			if err != nil {
				ctx.Status(http.StatusBadRequest)
				ctx.Result("Failed to parse parameters: " + err.Error())
				return
			}

			// 如果需要解析请求体，尝试解析 JSON 到结构体
			if needParseJsonBody && ctx.BodyType == Json && ctx.ContentLength > 0 {
				if err := json.Unmarshal(ctx.FillBody(), handlerInterface); err != nil {
					ctx.Status(http.StatusBadRequest)
					ctx.Result("Invalid JSON format: " + err.Error())
					return
				}
			}

			handler.Handle(ctx) // 调用 handler
		}

		runners = append(runners, runner)
	}

	return runners, nil
}

// parseParams 解析query参数和path参数和header参数到结构体字段
func parseParams(ctx *Context, handlerInterface interface{}) (needParseJsonBody bool, err error) {
	handlerValue := reflect.ValueOf(handlerInterface)
	if handlerValue.Kind() == reflect.Ptr {
		handlerValue = handlerValue.Elem()
	}
	return parseStructFields(handlerValue, ctx)
}

// parseStructFields 递归解析结构体字段
func parseStructFields(structValue reflect.Value, ctx *Context) (needParseJsonBody bool, err error) {
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
			needParseJsonBody = true // 交给 json.Unmarshal 处理
			continue
		}

		// 处理 form tag
		if formTag := field.Tag.Get("form"); formTag != "" {
			switch ctx.BodyType {
			case EncodeUrl:
				// 解析表单
				ctx.FillBody()
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
				ctx.FillBody()
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
			childNeedParseJsonBody, childErr := parseStructFields(fieldValue, ctx)
			if childErr != nil {
				return needParseJsonBody, childErr
			}
			needParseJsonBody = needParseJsonBody || childNeedParseJsonBody
		}
	}

	return
}

// serFileFieldValue 根据字段类型设置文件值
func serFileFieldValue(fieldValue reflect.Value, fileHeader ...*multipart.FileHeader) error {
	if len(fileHeader) == 0 {
		return nil
	}

	// 如果不是切片，或者是指向 []byte 的切片，只使用第一个文件
	if fieldValue.Kind() != reflect.Slice || fieldValue.Type() == reflect.TypeOf([]byte{}) {
		return setFileScalarValue(fieldValue, fileHeader[0])
	}

	// 创建新的切片
	newSlice := reflect.MakeSlice(fieldValue.Type(), len(fileHeader), len(fileHeader))
	// 为每个元素设置值
	for i, fh := range fileHeader {
		elemValue := newSlice.Index(i)
		if err := setFileScalarValue(elemValue, fh); err != nil {
			return err
		}
	}
	// 设置切片
	fieldValue.Set(newSlice)

	return nil
}

// setFileScalarValue 设置单个文件标量值
func setFileScalarValue(fieldValue reflect.Value, fileHeader *multipart.FileHeader) error {
	// 如果是 *multipart.FileHeader，直接设置
	if fieldValue.Type() == reflect.TypeOf(&multipart.FileHeader{}) {
		fieldValue.Set(reflect.ValueOf(fileHeader))
		return nil
	}
	// 如果是 []byte ，读取文件内容并设置
	if fieldValue.Type() == reflect.TypeOf([]byte{}) {
		file, err := fileHeader.Open()
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
		file, err := fileHeader.Open()
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

	// 对于非切片类型，只使用第一个值
	if fieldValue.Kind() != reflect.Slice {
		return setScalarValue(fieldValue, values[0])
	}

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

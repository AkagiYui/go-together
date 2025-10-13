package rest

import (
	"encoding/json"
	"net/http"
	"net/textproto"
	"reflect"
	"slices"

	"github.com/akagiyui/go-together/common/cache"
)

type BodyType int

const (
	Nil BodyType = iota
	EncodeUrl
	Json
	FormData
)

var structInfoCache = cache.NewCacheMap[reflect.Type, *structInfo]()

type structInfo struct {
	fields []fieldInfo
}

type fieldInfo struct {
	index     int
	name      string
	tagType   string // "query", "path", "header", "json", "form"
	tagValue  string
	fieldType reflect.Type
	isPtr     bool
}

// 获取或创建结构体信息缓存
func getStructInfo(t reflect.Type) *structInfo {
	return structInfoCache.GetOrSet(t, func() *structInfo {
		info := &structInfo{
			fields: make([]fieldInfo, 0),
		}

		// 预处理所有字段信息
		for i := 0; i < t.NumField(); i++ {
			field := t.Field(i)

			var tagType, tagValue string
			if tag := field.Tag.Get("query"); tag != "" {
				tagType, tagValue = "query", tag
			} else if tag := field.Tag.Get("path"); tag != "" {
				tagType, tagValue = "path", tag
			} else if tag := field.Tag.Get("header"); tag != "" {
				tagType, tagValue = "header", tag
			} else if tag := field.Tag.Get("json"); tag != "" {
				tagType, tagValue = "json", tag
			} else if tag := field.Tag.Get("form"); tag != "" {
				tagType, tagValue = "form", tag
			}

			if tagType != "" {
				info.fields = append(info.fields, fieldInfo{
					index:     i,
					name:      field.Name,
					tagType:   tagType,
					tagValue:  tagValue,
					fieldType: field.Type,
					isPtr:     field.Type.Kind() == reflect.Ptr,
				})
			}
		}

		return info
	})
}

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

// 优化后的 parseStructFields
func parseStructFields(structValue reflect.Value, ctx *Context) (needParseJsonBody bool, err error) {
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
	info := getStructInfo(structType)

	pathParams := ctx.PathParams
	queryValues := ctx.Query
	headers := ctx.Request.Header

	// 使用缓存的字段信息，避免重复反射
	for _, fieldInfo := range info.fields {
		fieldValue := structValue.Field(fieldInfo.index)

		if !fieldValue.CanSet() {
			continue
		}

		switch fieldInfo.tagType {
		case "query":
			if queryParam, ok := queryValues[fieldInfo.tagValue]; ok && len(queryParam) > 0 {
				if err = setFieldValue(fieldValue, queryParam...); err != nil {
					return
				}
			}
		case "path":
			if pathParam, ok := pathParams[fieldInfo.tagValue]; ok && pathParam != "" {
				if err = setFieldValue(fieldValue, pathParam); err != nil {
					return
				}
			}
		case "header":
			if headerValue, ok := headers[textproto.CanonicalMIMEHeaderKey(fieldInfo.tagValue)]; ok && len(headerValue) > 0 {
				if err = setFieldValue(fieldValue, headerValue...); err != nil {
					return
				}
			}
		case "json":
			if ctx.BodyType == Json {
				// 交给 json.Unmarshal 处理
				needParseJsonBody = true
			}
		case "form":
			switch ctx.BodyType {
			case EncodeUrl:
				// 解析表单
				ctx.FillBody()
				ctx.OriginalRequest.ParseForm()
				form := ctx.OriginalRequest.PostForm
				if form == nil {
					return false, nil
				}
				if formValue, ok := form[fieldInfo.tagValue]; ok && len(formValue) > 0 {
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
				if notFileValues, ok := notFileFieldsMap[fieldInfo.tagValue]; ok && len(notFileValues) > 0 {
					if err = setFieldValue(fieldValue, notFileValues...); err != nil {
						return
					}
				}

				// 处理文件
				fileFieldsMap := form.File
				if fileFields, ok := fileFieldsMap[fieldInfo.tagValue]; ok && len(fileFields) > 0 {
					if err = setFileFieldValue(fieldValue, fileFields...); err != nil {
						return
					}
				}
			}
			continue
		}
	}

	// 处理嵌套结构体
	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)
		fieldValue := structValue.Field(i)
		fieldType := field.Type

		if fieldType.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldType.Elem()))
			}
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() == reflect.Struct {
			// 检查是否有任何 tag，如果没有才递归
			hasTag := slices.ContainsFunc([]string{"query", "path", "header", "json", "form"}, func(tag string) bool {
				return field.Tag.Get(tag) != ""
			})

			if !hasTag {
				childNeedParseJsonBody, childErr := parseStructFields(fieldValue, ctx)
				if childErr != nil {
					return needParseJsonBody, childErr
				}
				needParseJsonBody = needParseJsonBody || childNeedParseJsonBody
			}
		}
	}

	return
}

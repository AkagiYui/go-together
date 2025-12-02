package rest

import (
	"net/textproto"
	"reflect"
	"runtime"
	"slices"

	"github.com/akagiyui/go-together/common/cache"
)

// BodyType 请求体类型
type BodyType int

const (
	// Nil 空请求体
	Nil BodyType = iota
	// EncodeURL URL编码格式
	EncodeURL
	// JSON JSON格式
	JSON
	// FormData 表单数据格式
	FormData
)

var structInfoCache = cache.NewMap[reflect.Type, *structInfo]()

type structInfo struct {
	fields []fieldInfo
}

type fieldInfo struct {
	index     int
	name      string
	tagType   string // "query", "path", "header", "json", "form", "context"
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
			} else if tag := field.Tag.Get("context"); tag != "" {
				tagType, tagValue = "context", tag
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

// funcName 获取 HandlerFunc 的函数名称
func funcName(f HandlerFunc) string {
	return runtime.FuncForPC(reflect.ValueOf(f).Pointer()).Name()
}

// parseParams 解析query参数和path参数和header参数到结构体字段
func parseParams(ctx *Context, handlerInterface interface{}) (needParseJSONBody bool, err error) {
	handlerValue := reflect.ValueOf(handlerInterface)
	if handlerValue.Kind() == reflect.Ptr {
		handlerValue = handlerValue.Elem()
	}
	return parseStructFields(handlerValue, ctx)
}

// 优化后的 parseStructFields
func parseStructFields(structValue reflect.Value, ctx *Context) (needParseJSONBody bool, err error) {
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
		case "context":
			if contextValue, exists := ctx.Get(fieldInfo.tagValue); exists {
				if err = setAnyValue(fieldValue, contextValue); err != nil {
					return
				}
			}
		case "json":
			if ctx.BodyType == JSON {
				// 交给 json.Unmarshal 处理
				needParseJSONBody = true
			}
		case "form":
			switch ctx.BodyType {
			case EncodeURL:
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

		// 检查是否有任何 tag
		hasTag := slices.ContainsFunc([]string{"query", "path", "header", "json", "form", "context"}, func(tag string) bool {
			return field.Tag.Get(tag) != ""
		})

		// 如果有标签，跳过嵌套结构体处理
		if hasTag {
			continue
		}

		if fieldType.Kind() == reflect.Ptr {
			if fieldValue.IsNil() {
				fieldValue.Set(reflect.New(fieldType.Elem()))
			}
			fieldType = fieldType.Elem()
		}

		if fieldType.Kind() == reflect.Struct {
			childNeedParseJSONBody, childErr := parseStructFields(fieldValue, ctx)
			if childErr != nil {
				return needParseJSONBody, childErr
			}
			needParseJSONBody = needParseJSONBody || childNeedParseJSONBody
		}
	}

	return
}

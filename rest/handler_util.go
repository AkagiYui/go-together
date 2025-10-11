package rest

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
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

		runner := func(ctx *Context) {
			// 创建新的 HandlerInterface 实例
			handlerValue := reflect.New(t)
			handlerInterface := handlerValue.Interface()
			handler, ok := handlerInterface.(HandlerInterface)
			if !ok {
				panic("Handler does not implement HandlerInterface")
			}

			// 判断请求体类型
			bodyType := Nil
			contentType := strings.ToLower(strings.Trim(ctx.Request.Header.Get("Content-Type"), " "))
			if strings.HasPrefix(contentType, "application/x-www-form-urlencoded") {
				bodyType = EncodeUrl
			} else if strings.HasPrefix(contentType, "application/json") {
				bodyType = Json
			} else if strings.HasPrefix(contentType, "multipart/form-data") {
				bodyType = FormData
			}

			// 解析 query/path/header 参数
			needParseBody, err := parseParams(ctx, handlerInterface)
			if err != nil {
				ctx.Status(http.StatusBadRequest)
				ctx.Result("Failed to parse parameters: " + err.Error())
				return
			}

			// 如果需要解析请求体，尝试解析 JSON 到结构体
			if needParseBody {
				if bodyType == Json && ctx.ContentLength > 0 {
					if err := json.Unmarshal(ctx.FillBody(), handlerInterface); err != nil {
						ctx.Status(http.StatusBadRequest)
						ctx.Result("Invalid JSON format: " + err.Error())
						return
					}
				}
			}

			handler.Handle(ctx) // 调用 handler
		}

		runners = append(runners, runner)
	}

	return runners, nil
}

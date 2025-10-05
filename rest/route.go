package rest

import (
	"net/http"
	"reflect"
)

type HandlerInterface interface {
	Handle(*Context) any
}

// HandlerFactory 用于创建处理器实例的工厂
type HandlerFactory struct {
	Path        string
	Method      string
	HandlerType reflect.Type
}

type RouteGroup struct {
	Factories []HandlerFactory
	BasePath  string
}

// Handle 接受结构体类型，每次请求时创建新实例
func (g *RouteGroup) Handle(path string, method string, handlerType HandlerInterface) {
	t := reflect.TypeOf(handlerType)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	factory := HandlerFactory{
		Path:        path,
		Method:      method,
		HandlerType: t,
	}
	g.Factories = append(g.Factories, factory)
}

func (g *RouteGroup) GET(path string, handlerType HandlerInterface) {
	g.Handle(path, http.MethodGet, handlerType)
}

func (g *RouteGroup) POST(path string, handlerType HandlerInterface) {
	g.Handle(path, http.MethodPost, handlerType)
}

func (g *RouteGroup) PUT(path string, handlerType HandlerInterface) {
	g.Handle(path, http.MethodPut, handlerType)
}

func (g *RouteGroup) DELETE(path string, handlerType HandlerInterface) {
	g.Handle(path, http.MethodDelete, handlerType)
}

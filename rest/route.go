package rest

import "reflect"

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

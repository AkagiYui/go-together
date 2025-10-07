package rest

import (
	"net/http"
	"reflect"
)

type HandlerFunc func(*Context)

type HandlerInterface interface {
	Handle(*Context)
}

// HandlerFactory 用于创建处理器实例的工厂
type HandlerFactory struct {
	Path        string
	Method      string
	HandlerType reflect.Type
	HandlerFunc HandlerFunc // 用于存储函数处理器
	IsFunc      bool        // 标识是否为函数处理器
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

// HandlerFuncAdapter 适配器结构体，用于将 HandlerFunc 转换为 HandlerInterface
// 注意：这个结构体现在主要用于测试，实际运行时函数处理器直接调用，不使用适配器
type HandlerFuncAdapter struct {
	handlerFunc HandlerFunc
}

// Handle 实现 HandlerInterface 接口
func (h *HandlerFuncAdapter) Handle(ctx *Context) {
	h.handlerFunc(ctx)
}

func (g *RouteGroup) HandleFunc(path string, method string, handlerFunc HandlerFunc) {
	// 直接存储函数处理器，不使用适配器
	factory := HandlerFactory{
		Path:        path,
		Method:      method,
		HandlerFunc: handlerFunc,
		IsFunc:      true,
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

// 便捷方法：支持 HandlerFunc 的 HTTP 方法
func (g *RouteGroup) GETFunc(path string, handlerFunc HandlerFunc) {
	g.HandleFunc(path, http.MethodGet, handlerFunc)
}

func (g *RouteGroup) POSTFunc(path string, handlerFunc HandlerFunc) {
	g.HandleFunc(path, http.MethodPost, handlerFunc)
}

func (g *RouteGroup) PUTFunc(path string, handlerFunc HandlerFunc) {
	g.HandleFunc(path, http.MethodPut, handlerFunc)
}

func (g *RouteGroup) DELETEFunc(path string, handlerFunc HandlerFunc) {
	g.HandleFunc(path, http.MethodDelete, handlerFunc)
}

package rest

import "net/http"

// TODO 处理err返回值
func (g *RouteGroup) GET(path string, handlerType ...HandlerInterface) {
	g.Handle(path, http.MethodGet, handlerType...)
}

func (g *RouteGroup) POST(path string, handlerType ...HandlerInterface) {
	g.Handle(path, http.MethodPost, handlerType...)
}

func (g *RouteGroup) PUT(path string, handlerType ...HandlerInterface) {
	g.Handle(path, http.MethodPut, handlerType...)
}

func (g *RouteGroup) DELETE(path string, handlerType ...HandlerInterface) {
	g.Handle(path, http.MethodDelete, handlerType...)
}

// 便捷方法：支持 HandlerFunc 的 HTTP 方法
func (g *RouteGroup) GETFunc(path string, handlerFunc ...HandlerFunc) {
	g.HandleFunc(path, http.MethodGet, handlerFunc...)
}

func (g *RouteGroup) POSTFunc(path string, handlerFunc ...HandlerFunc) {
	g.HandleFunc(path, http.MethodPost, handlerFunc...)
}

func (g *RouteGroup) PUTFunc(path string, handlerFunc ...HandlerFunc) {
	g.HandleFunc(path, http.MethodPut, handlerFunc...)
}

func (g *RouteGroup) DELETEFunc(path string, handlerFunc ...HandlerFunc) {
	g.HandleFunc(path, http.MethodDelete, handlerFunc...)
}

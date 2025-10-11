package rest

import "net/http"

func (g *RouteGroup) GET(path string, handlerType ...HandlerInterface) error {
	return g.Handle(path, http.MethodGet, handlerType...)
}

func (g *RouteGroup) POST(path string, handlerType ...HandlerInterface) error {
	return g.Handle(path, http.MethodPost, handlerType...)
}

func (g *RouteGroup) PUT(path string, handlerType ...HandlerInterface) error {
	return g.Handle(path, http.MethodPut, handlerType...)
}

func (g *RouteGroup) DELETE(path string, handlerType ...HandlerInterface) error {
	return g.Handle(path, http.MethodDelete, handlerType...)
}

// 便捷方法：支持 HandlerFunc 的 HTTP 方法
func (g *RouteGroup) GETFunc(path string, handlerFunc ...HandlerFunc) error {
	g.HandleFunc(path, http.MethodGet, handlerFunc...)
	return nil
}

func (g *RouteGroup) POSTFunc(path string, handlerFunc ...HandlerFunc) error {
	g.HandleFunc(path, http.MethodPost, handlerFunc...)
	return nil
}

func (g *RouteGroup) PUTFunc(path string, handlerFunc ...HandlerFunc) error {
	g.HandleFunc(path, http.MethodPut, handlerFunc...)
	return nil
}

func (g *RouteGroup) DELETEFunc(path string, handlerFunc ...HandlerFunc) error {
	g.HandleFunc(path, http.MethodDelete, handlerFunc...)
	return nil
}

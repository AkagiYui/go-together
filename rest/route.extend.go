package rest

import "net/http"

func (g *RouteGroup) Get(path string, handlerType ...HandlerInterface) error {
	return g.Handle(path, http.MethodGet, handlerType...)
}

func (g *RouteGroup) Post(path string, handlerType ...HandlerInterface) error {
	return g.Handle(path, http.MethodPost, handlerType...)
}

func (g *RouteGroup) Put(path string, handlerType ...HandlerInterface) error {
	return g.Handle(path, http.MethodPut, handlerType...)
}

func (g *RouteGroup) Delete(path string, handlerType ...HandlerInterface) error {
	return g.Handle(path, http.MethodDelete, handlerType...)
}

// 便捷方法：支持 HandlerFunc 的 HTTP 方法
func (g *RouteGroup) GetFunc(path string, handlerFunc ...HandlerFunc) error {
	g.HandleFunc(path, http.MethodGet, handlerFunc...)
	return nil
}

func (g *RouteGroup) PostFunc(path string, handlerFunc ...HandlerFunc) error {
	g.HandleFunc(path, http.MethodPost, handlerFunc...)
	return nil
}

func (g *RouteGroup) PutFunc(path string, handlerFunc ...HandlerFunc) error {
	g.HandleFunc(path, http.MethodPut, handlerFunc...)
	return nil
}

func (g *RouteGroup) DeleteEFunc(path string, handlerFunc ...HandlerFunc) error {
	g.HandleFunc(path, http.MethodDelete, handlerFunc...)
	return nil
}

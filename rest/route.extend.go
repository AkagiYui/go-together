package rest

import "net/http"

// Get 注册 GET 方法的处理器
func (g *RouteGroup) Get(path string, handlers ...HandlerFunc) {
	g.Handle(path, http.MethodGet, handlers...)
}

// Post 注册 POST 方法的处理器
func (g *RouteGroup) Post(path string, handlers ...HandlerFunc) {
	g.Handle(path, http.MethodPost, handlers...)
}

// Put 注册 PUT 方法的处理器
func (g *RouteGroup) Put(path string, handlers ...HandlerFunc) {
	g.Handle(path, http.MethodPut, handlers...)
}

// Delete 注册 DELETE 方法的处理器
func (g *RouteGroup) Delete(path string, handlers ...HandlerFunc) {
	g.Handle(path, http.MethodDelete, handlers...)
}

// Patch 注册 PATCH 方法的处理器
func (g *RouteGroup) Patch(path string, handlers ...HandlerFunc) {
	g.Handle(path, http.MethodPatch, handlers...)
}

// Any 注册任意 HTTP 方法的处理器
func (g *RouteGroup) Any(path string, handlers ...HandlerFunc) {
	g.Handle(path, "", handlers...)
}

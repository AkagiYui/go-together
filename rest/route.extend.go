package rest

import "net/http"

// Get 注册 GET 方法的处理器
func (g *RouteGroup) Get(path string, handlerType ...HandlerInterface) error {
	return g.Handle(path, http.MethodGet, handlerType...)
}

// Post 注册 POST 方法的处理器
func (g *RouteGroup) Post(path string, handlerType ...HandlerInterface) error {
	return g.Handle(path, http.MethodPost, handlerType...)
}

// Put 注册 PUT 方法的处理器
func (g *RouteGroup) Put(path string, handlerType ...HandlerInterface) error {
	return g.Handle(path, http.MethodPut, handlerType...)
}

// Delete 注册 DELETE 方法的处理器
func (g *RouteGroup) Delete(path string, handlerType ...HandlerInterface) error {
	return g.Handle(path, http.MethodDelete, handlerType...)
}

// Patch 注册 PATCH 方法的处理器
func (g *RouteGroup) Patch(path string, handlerType ...HandlerInterface) error {
	return g.Handle(path, http.MethodPatch, handlerType...)
}

// Any 注册任意 HTTP 方法的处理器
func (g *RouteGroup) Any(path string, handlerType ...HandlerInterface) error {
	return g.Handle(path, "", handlerType...)
}

// GetFunc 注册 GET 方法的函数处理器
func (g *RouteGroup) GetFunc(path string, handlerFunc ...HandlerFunc) error {
	g.HandleFunc(path, http.MethodGet, handlerFunc...)
	return nil
}

// PostFunc 注册 POST 方法的函数处理器
func (g *RouteGroup) PostFunc(path string, handlerFunc ...HandlerFunc) error {
	g.HandleFunc(path, http.MethodPost, handlerFunc...)
	return nil
}

// PutFunc 注册 PUT 方法的函数处理器
func (g *RouteGroup) PutFunc(path string, handlerFunc ...HandlerFunc) error {
	g.HandleFunc(path, http.MethodPut, handlerFunc...)
	return nil
}

// DeleteFunc 注册 DELETE 方法的函数处理器
func (g *RouteGroup) DeleteFunc(path string, handlerFunc ...HandlerFunc) error {
	g.HandleFunc(path, http.MethodDelete, handlerFunc...)
	return nil
}

// PatchFunc 注册 PATCH 方法的函数处理器
func (g *RouteGroup) PatchFunc(path string, handlerFunc ...HandlerFunc) error {
	g.HandleFunc(path, http.MethodPatch, handlerFunc...)
	return nil
}

// AnyFunc 注册任意 HTTP 方法的函数处理器
func (g *RouteGroup) AnyFunc(path string, handlerFunc ...HandlerFunc) error {
	g.HandleFunc(path, "", handlerFunc...)
	return nil
}

// GetServ 注册 GET 方法的服务处理器
func (g *RouteGroup) GetServ(path string, handlerType ...ServiceHandlerInterface) error {
	return g.HandleServ(path, http.MethodGet, handlerType...)
}

// PostServ 注册 POST 方法的服务处理器
func (g *RouteGroup) PostServ(path string, handlerType ...ServiceHandlerInterface) error {
	return g.HandleServ(path, http.MethodPost, handlerType...)
}

// PutServ 注册 PUT 方法的服务处理器
func (g *RouteGroup) PutServ(path string, handlerType ...ServiceHandlerInterface) error {
	return g.HandleServ(path, http.MethodPut, handlerType...)
}

// DeleteServ 注册 DELETE 方法的服务处理器
func (g *RouteGroup) DeleteServ(path string, handlerType ...ServiceHandlerInterface) error {
	return g.HandleServ(path, http.MethodDelete, handlerType...)
}

// PatchServ 注册 PATCH 方法的服务处理器
func (g *RouteGroup) PatchServ(path string, handlerType ...ServiceHandlerInterface) error {
	return g.HandleServ(path, http.MethodPatch, handlerType...)
}

// AnyServ 注册任意 HTTP 方法的服务处理器
func (g *RouteGroup) AnyServ(path string, handlerType ...ServiceHandlerInterface) error {
	return g.HandleServ(path, "", handlerType...)
}

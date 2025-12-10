package rest

// HandlerFunc 处理函数类型
type HandlerFunc func(*Context)

// HandlerInterface 处理器接口
type HandlerInterface interface {
	Handle(*Context)
}

// ServiceHandlerInterface 服务处理器接口
type ServiceHandlerInterface interface {
	Do() (any, error)
}

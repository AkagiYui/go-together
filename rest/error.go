package rest

// ErrHandlerNotImplementHandlerInterface 处理器未实现 HandlerInterface 接口的错误
type ErrHandlerNotImplementHandlerInterface struct{}

func (e ErrHandlerNotImplementHandlerInterface) Error() string {
	return "Handler does not implement HandlerInterface"
}

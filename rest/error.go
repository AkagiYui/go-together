package rest

type ErrHandlerNotImplementHandlerInterface struct{}

func (e ErrHandlerNotImplementHandlerInterface) Error() string {
	return "Handler does not implement HandlerInterface"
}

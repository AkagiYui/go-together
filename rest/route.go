package rest

type HandlerFunc func(*Context)

type HandlerInterface interface {
	Handle(*Context)
}

// HandlerFactory 用于创建 handler 实例的工厂
type HandlerFactory struct {
	Path        string
	Method      string
	RunnerChain []HandlerFunc
}

// Handle 接受结构体类型，每次请求时创建新实例
func (g *RouteGroup) Handle(path string, method string, handlerTypes ...HandlerInterface) error {
	factory := HandlerFactory{
		Path:        path,
		Method:      method,
		RunnerChain: nil,
	}

	runners, err := runnersFromHandlers(handlerTypes...)
	if err != nil {
		return err
	}
	factory.RunnerChain = runners
	g.Factories = append(g.Factories, factory)
	return nil
}

func (g *RouteGroup) HandleFunc(path string, method string, handlerFuncs ...HandlerFunc) {
	factory := HandlerFactory{
		Path:        path,
		Method:      method,
		RunnerChain: nil,
	}
	factory.RunnerChain = handlerFuncs
	g.Factories = append(g.Factories, factory)
}

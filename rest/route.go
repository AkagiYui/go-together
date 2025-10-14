package rest

type HandlerFunc func(*Context)

type HandlerInterface interface {
	Handle(*Context)
}

// HandlerFactory 用于创建 handler 实例的工厂
type HandlerFactory struct {
	Path         string
	Method       string
	RunnerChain  []HandlerFunc
	HandlerNames []string // 存储每个 handler 的名称，用于调试输出
}

// Handle 接受结构体类型，每次请求时创建新实例
func (g *RouteGroup) Handle(path string, method string, handlerTypes ...HandlerInterface) error {
	factory := HandlerFactory{
		Path:         path,
		Method:       method,
		RunnerChain:  nil,
		HandlerNames: nil,
	}

	runners, names, err := runnersFromHandlers(handlerTypes...)
	if err != nil {
		return err
	}
	factory.RunnerChain = runners
	factory.HandlerNames = names
	g.Factories = append(g.Factories, factory)
	return nil
}

func (g *RouteGroup) HandleFunc(path string, method string, handlerFuncs ...HandlerFunc) {
	factory := HandlerFactory{
		Path:         path,
		Method:       method,
		RunnerChain:  nil,
		HandlerNames: nil,
	}
	factory.RunnerChain = handlerFuncs
	// 获取函数名称
	factory.HandlerNames = make([]string, len(handlerFuncs))
	for i, f := range handlerFuncs {
		factory.HandlerNames[i] = funcName(f)
	}
	g.Factories = append(g.Factories, factory)
}

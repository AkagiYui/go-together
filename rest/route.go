package rest

type HandlerFunc func(*Context)

type HandlerInterface interface {
	Handle(*Context)
}
type ServiceHandlerInterface interface {
	Do() (any, error)
}

// Validator 接口用于在参数绑定后、业务处理前进行数据校验
// 实现此接口的 handler 会在 Handle 方法调用前自动执行 Validate 方法
type Validator interface {
	Validate() error
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

func (g *RouteGroup) HandleService(path string, method string, handlerTypes ...ServiceHandlerInterface) error {
	return nil
}

func (g *RouteGroup) HandleFunc(path string, method string, handlerFuncs ...HandlerFunc) {
	factory := HandlerFactory{
		Path:         path,
		Method:       method,
		RunnerChain:  nil,
		HandlerNames: nil,
	}

	// TODO 允许空方法列表，添加一个默认的空方法
	if len(handlerFuncs) == 0 {
		handlerFuncs = append(handlerFuncs, func(ctx *Context) {})
	}

	factory.RunnerChain = handlerFuncs
	// 获取函数名称
	factory.HandlerNames = make([]string, len(handlerFuncs))
	for i, f := range handlerFuncs {
		factory.HandlerNames[i] = funcName(f)
	}
	g.Factories = append(g.Factories, factory)
}

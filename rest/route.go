package rest

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

// Handle 注册处理器到指定路径和方法
func (g *RouteGroup) Handle(path string, method string, handlers ...HandlerFunc) {
	factory := HandlerFactory{
		Path:         path,
		Method:       method,
		RunnerChain:  nil,
		HandlerNames: nil,
	}

	// 允许空方法列表，添加一个默认的空方法
	if len(handlers) == 0 {
		handlers = append(handlers, func(_ *Context) {})
	}

	factory.RunnerChain = handlers
	// 获取函数名称
	factory.HandlerNames = make([]string, len(handlers))
	for i, f := range handlers {
		factory.HandlerNames[i] = funcName(f)
	}
	g.Factories = append(g.Factories, factory)
}

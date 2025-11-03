package rest

// RouteGroup 路由组
type RouteGroup struct {
	Factories      []HandlerFactory
	BasePath       string
	ChildGroups    []*RouteGroup
	PreRunnerChain []HandlerFunc
	PreRunnerNames []string // 存储前置 handler 的名称，用于调试输出

	server *Server
}

// NewRouteGroup 创建一个新的路由组
func NewRouteGroup(server *Server, basePath string, preRunnerChain ...HandlerFunc) RouteGroup {
	// 获取 preRunnerChain 的函数名称
	preRunnerNames := make([]string, len(preRunnerChain))
	for i, f := range preRunnerChain {
		preRunnerNames[i] = funcName(f)
	}

	return RouteGroup{
		Factories:      make([]HandlerFactory, 0),
		BasePath:       basePath,
		ChildGroups:    make([]*RouteGroup, 0),
		PreRunnerChain: preRunnerChain,
		PreRunnerNames: preRunnerNames,
		server:         server,
	}
}

// Group 创建子组
func (g *RouteGroup) Group(basePath string, preRunnerChain ...HandlerFunc) *RouteGroup {
	childGroup := NewRouteGroup(g.server, basePath, preRunnerChain...)
	g.ChildGroups = append(g.ChildGroups, &childGroup)
	return &childGroup
}

// UseFunc 为当前组添加前置函数处理器
func (g *RouteGroup) UseFunc(handlers ...HandlerFunc) {
	g.PreRunnerChain = append(g.PreRunnerChain, handlers...)
	// 获取函数名称
	for _, f := range handlers {
		g.PreRunnerNames = append(g.PreRunnerNames, funcName(f))
	}
}

// Use 为当前组添加前置处理器
func (g *RouteGroup) Use(handlers ...HandlerInterface) {
	// 复用与 Handle 相同的结构体 Handler 构造逻辑
	runners, names, err := runnersFromHandlers(handlers...)
	if err != nil {
		// 与 Handle 的错误语义保持一致，这里无法返回 error，只能在编程错误时直接失败
		panic(err)
	}
	g.PreRunnerChain = append(g.PreRunnerChain, runners...)
	g.PreRunnerNames = append(g.PreRunnerNames, names...)
}

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

// Use 为当前组添加前置处理器
func (g *RouteGroup) Use(handlers ...HandlerFunc) {
	g.PreRunnerChain = append(g.PreRunnerChain, handlers...)
	// 获取函数名称
	for _, f := range handlers {
		g.PreRunnerNames = append(g.PreRunnerNames, funcName(f))
	}
}

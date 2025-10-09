package rest

type RouteGroup struct {
	Factories      []HandlerFactory
	BasePath       string
	ChildGroups    []*RouteGroup
	PreRunnerChain []HandlerFunc
}

func NewRouteGroup(basePath string, preRunnerChain ...HandlerFunc) RouteGroup {
	if preRunnerChain == nil {
		preRunnerChain = make([]HandlerFunc, 0)
	}
	return RouteGroup{
		Factories:      make([]HandlerFactory, 0),
		BasePath:       basePath,
		ChildGroups:    make([]*RouteGroup, 0),
		PreRunnerChain: preRunnerChain,
	}
}

// Group 创建子组
func (g *RouteGroup) Group(basePath string, preRunnerChain ...HandlerFunc) *RouteGroup {
	childGroup := NewRouteGroup(g.BasePath+basePath, preRunnerChain...)
	g.ChildGroups = append(g.ChildGroups, &childGroup)
	childGroup.PreRunnerChain = append(childGroup.PreRunnerChain, g.PreRunnerChain...)
	return &childGroup
}

// Use 为当前组添加中间件
func (g *RouteGroup) UseFunc(middlewares ...HandlerFunc) {
	g.PreRunnerChain = append(g.PreRunnerChain, middlewares...)
}

func (g *RouteGroup) Use(middlewares ...HandlerInterface) {
	// 复用与 Handle 相同的结构体 Handler 构造逻辑
	runners, err := runnersFromHandlers(middlewares...)
	if err != nil {
		// 与 Handle 的错误语义保持一致，这里无法返回 error，只能在编程错误时直接失败
		panic(err)
	}
	g.PreRunnerChain = append(g.PreRunnerChain, runners...)
}

package rest

type RouteGroup struct {
	Factories   []HandlerFactory
	BasePath    string
	ChildGroups []*RouteGroup
}

func NewRouteGroup(basePath string) RouteGroup {
	return RouteGroup{
		Factories:   make([]HandlerFactory, 0),
		BasePath:    basePath,
		ChildGroups: make([]*RouteGroup, 0),
	}
}

func (g *RouteGroup) Group(basePath string) *RouteGroup {
	group := NewRouteGroup(g.BasePath + basePath)
	g.ChildGroups = append(g.ChildGroups, &group)
	return &group
}

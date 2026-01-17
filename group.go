package zorya

import "net/http"

// Group is a collection of routes that share a common prefix and set of
// operation modifiers, middlewares, and transformers.
//
// This is useful for grouping related routes together and applying common
// settings to them. For example, you might create a group for all routes that
// require authentication.
type Group struct {
	API
	prefixes     []string
	adapter      Adapter
	modifiers    []func(o *BaseRoute, next func(*BaseRoute))
	middlewares  Middlewares
	transformers []Transformer
	security     *RouteSecurity
}

// groupAdapter is an Adapter wrapper that registers multiple operation handlers
// with the underlying adapter based on the group's prefixes.
type groupAdapter struct {
	Adapter
	group *Group
}

// NewGroup creates a new group of routes with the given prefixes, if any. A
// group enables a collection of operations to have the same prefix and share
// operation modifiers, middlewares, and transformers.
//
//	grp := zorya.NewGroup(api, "/v1")
//	grp.UseMiddleware(authMiddleware)
//
//	zorya.Get(grp, "/users", func(ctx context.Context, input *MyInput) (*MyOutput, error) {
//		// Your code here...
//	})
func NewGroup(api API, prefixes ...string) *Group {
	group := &Group{API: api, prefixes: prefixes}
	group.adapter = &groupAdapter{Adapter: api.Adapter(), group: group}
	if len(prefixes) > 0 {
		group.UseModifier(PrefixModifier(prefixes))
	}

	return group
}

func (g *Group) Adapter() Adapter {
	return g.adapter
}

func (a *groupAdapter) Handle(route *BaseRoute, handler http.HandlerFunc) {
	a.group.ModifyOperation(route, func(route *BaseRoute) {
		a.Adapter.Handle(route, handler)
	})
}

// ModifyOperation runs all operation modifiers in the group on the given
// route, in the order they were added. This is useful for modifying a route
// before it is registered with the router.
func (g *Group) ModifyOperation(route *BaseRoute, next func(*BaseRoute)) {
	g.mergeSecurity(route)

	chain := func(route *BaseRoute) {
		// Call the final handler.
		next(route)
	}

	for i := len(g.modifiers) - 1; i >= 0; i-- {
		// Use an inline func to provide a closure around the index & chain.
		func(i int, n func(*BaseRoute)) {
			chain = func(route *BaseRoute) { g.modifiers[i](route, n) }
		}(i, chain)
	}

	chain(route)
}

// mergeSecurity merges group security with route security.
func (g *Group) mergeSecurity(route *BaseRoute) {
	if g.security == nil {
		return
	}

	if route.Security == nil {
		route.Security = g.copySecurity()

		return
	}

	g.mergeSecurityFields(route.Security)
}

// copySecurity creates a deep copy of the group's security configuration.
func (g *Group) copySecurity() *RouteSecurity {
	return &RouteSecurity{
		Roles:       append([]string(nil), g.security.Roles...),
		Permissions: append([]string(nil), g.security.Permissions...),
		Resource:    g.security.Resource,
		Action:      g.security.Action,
	}
}

// mergeSecurityFields merges group security fields into route security.
func (g *Group) mergeSecurityFields(routeSec *RouteSecurity) {
	if len(routeSec.Roles) == 0 && len(g.security.Roles) > 0 {
		routeSec.Roles = append([]string(nil), g.security.Roles...)
	}
	if len(routeSec.Permissions) == 0 && len(g.security.Permissions) > 0 {
		routeSec.Permissions = append([]string(nil), g.security.Permissions...)
	}
	if routeSec.Resource == "" && g.security.Resource != "" {
		routeSec.Resource = g.security.Resource
	}
	if routeSec.Action == "" && g.security.Action != "" {
		routeSec.Action = g.security.Action
	}
}

// UseModifier adds an operation modifier function to the group that will be run
// on all operations in the group. Use this to modify the operation before it is
// registered with the router. This behaves similar to middleware in that you
// should invoke `next` to continue the chain. Skip it to prevent the operation
// from being registered, and call multiple times for a fan-out effect.
func (g *Group) UseModifier(modifier func(o *BaseRoute, next func(*BaseRoute))) {
	g.modifiers = append(g.modifiers, modifier)
}

// UseSimpleModifier adds an operation modifier function to the group that
// will be run on all operations in the group. Use this to modify the operation
// before it is registered with the router.
func (g *Group) UseSimpleModifier(modifier func(o *BaseRoute)) {
	g.modifiers = append(g.modifiers, func(o *BaseRoute, next func(*BaseRoute)) {
		modifier(o)
		next(o)
	})
}

// UseMiddleware adds one or more standard Go middleware functions to the group.
// Middleware functions take an http.Handler and return an http.Handler.
func (g *Group) UseMiddleware(middlewares ...Middleware) {
	g.middlewares = append(g.middlewares, middlewares...)
}

func (g *Group) Middlewares() Middlewares {
	m := append(Middlewares{}, g.API.Middlewares()...)

	return append(m, g.middlewares...)
}

// UseTransformer adds one or more transformer functions to the group that will
// be run on all responses in the group.
func (g *Group) UseTransformer(transformers ...Transformer) {
	g.transformers = append(g.transformers, transformers...)
}

// WithSecurity sets security requirements for all routes in the group.
func (g *Group) WithSecurity(security *RouteSecurity) {
	g.security = security
}

// UseRoles requires users to have at least one of the specified roles for all routes in the group.
func (g *Group) UseRoles(roles ...string) {
	if g.security == nil {
		g.security = &RouteSecurity{}
	}
	g.security.Roles = append(g.security.Roles, roles...)
}

// UsePermissions requires users to have all specified permissions for all routes in the group.
func (g *Group) UsePermissions(perms ...string) {
	if g.security == nil {
		g.security = &RouteSecurity{}
	}
	g.security.Permissions = append(g.security.Permissions, perms...)
}

// UseResource sets the RBAC resource for all routes in the group.
func (g *Group) UseResource(resource string) {
	if g.security == nil {
		g.security = &RouteSecurity{}
	}
	g.security.Resource = resource
}

// Transform runs all transformers in the group on the response, in the order
// they were added, then chains to the parent API's transformers.
func (g *Group) Transform(r *http.Request, status int, v any) (any, error) {
	// Run group-specific transformers first.
	for _, transformer := range g.transformers {
		var err error
		v, err = transformer(r, status, v)
		if err != nil {
			return v, err
		}
	}

	// Chain to parent API transformers.
	return g.API.Transform(r, status, v)
}

// PrefixModifier provides a fan-out to register one or more operations with
// the given prefix for every one operation added to a group.
func PrefixModifier(prefixes []string) func(o *BaseRoute, next func(*BaseRoute)) {
	return func(o *BaseRoute, next func(*BaseRoute)) {
		for _, prefix := range prefixes {
			modified := *o
			modified.Path = prefix + modified.Path
			next(&modified)
		}
	}
}

package routing

import "context"

// currentRouteKey is the unexported context key under which the request-scoped
// matched route is stored. Using a private zero-size type guarantees no
// collision with other packages' context values.
type currentRouteKey struct{}

// ContextRouteResolver resolves the current route from a context.Context. It
// structurally satisfies foundation.RouteResolver without depending on shared
// router state, so each request reads only its own matched route. Bind it once
// (it is stateless) with foundation.Request.SetRouteResolver.
type ContextRouteResolver struct{}

// WithCurrentRoute returns a copy of ctx carrying route as the request-scoped
// current route. The router calls this after matching so that handlers and
// middleware can read the matched route back with the package-level accessors
// (CurrentRoute, CurrentRouteName, ...) without touching shared router state.
//
// A nil ctx is treated as context.Background(); a nil route is stored as-is and
// read back as nil.
func WithCurrentRoute(ctx context.Context, route *Route) context.Context {
	if ctx == nil {
		ctx = context.Background()
	}

	return context.WithValue(ctx, currentRouteKey{}, route)
}

// CurrentRoute returns the request-scoped route stored in ctx by
// WithCurrentRoute, or nil when none is present (including a nil ctx). This is
// the concurrency-safe replacement for the deprecated Router.Current method.
func CurrentRoute(ctx context.Context) *Route {
	if ctx == nil {
		return nil
	}

	route, _ := ctx.Value(currentRouteKey{}).(*Route)

	return route
}

// CurrentRouteName returns the name of the request-scoped route in ctx, or "".
func CurrentRouteName(ctx context.Context) string {
	route := CurrentRoute(ctx)

	if route == nil {
		return ""
	}

	return route.GetName()
}

// CurrentRouteAction returns the handler action of the request-scoped route in
// ctx, or "".
func CurrentRouteAction(ctx context.Context) string {
	route := CurrentRoute(ctx)

	if route == nil {
		return ""
	}

	return route.GetActionName()
}

// CurrentRouteIs reports whether the request-scoped route's name matches any of
// the given glob patterns. It is the context-scoped counterpart of Router.Is.
func CurrentRouteIs(ctx context.Context, patterns ...string) bool {
	route := CurrentRoute(ctx)

	if route == nil {
		return false
	}

	return route.Named(patterns...)
}

// CurrentRouteNamed is an alias of CurrentRouteIs kept for parity with
// Router.CurrentRouteNamed.
func CurrentRouteNamed(ctx context.Context, patterns ...string) bool {
	return CurrentRouteIs(ctx, patterns...)
}

// CurrentRouteName returns the name of the route stored in ctx, or "".
func (ContextRouteResolver) CurrentRouteName(ctx context.Context) string {
	return CurrentRouteName(ctx)
}

// CurrentRouteAction returns the action of the route stored in ctx, or "".
func (ContextRouteResolver) CurrentRouteAction(ctx context.Context) string {
	return CurrentRouteAction(ctx)
}

// CurrentRouteUses reports whether the request-scoped route's action matches any
// of the given glob patterns. It is the context-scoped counterpart of
// Router.Uses.
func CurrentRouteUses(ctx context.Context, patterns ...string) bool {
	route := CurrentRoute(ctx)

	if route == nil {
		return false
	}

	action := route.GetActionName()

	for _, p := range patterns {
		if matchPattern(p, action) {
			return true
		}
	}

	return false
}

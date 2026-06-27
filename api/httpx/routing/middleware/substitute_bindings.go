// The bedrock pipeline runs each middleware as a function that receives the
// current request and a "next" continuation. The shapes here are designed so
// they can be wrapped trivially by the bedrock pipeline package in M11.
package middleware

import (
	"fmt"

	"github.com/oullin/alloy/api/httpx/routing"
)

// BindingRouter is the minimum router surface SubstituteBindings touches.
// Both [routing.Router] and any future test fake satisfy it.
type BindingRouter interface {
	SubstituteBindings(route *routing.Route) error
}

// SubstituteBindings is the middleware that runs route bindings before the
// route handler executes.
type SubstituteBindings struct{ Router BindingRouter }

// New wraps a router into the middleware.
func New(router BindingRouter) *SubstituteBindings {
	return &SubstituteBindings{Router: router}
}

// Handle is the pipeline entry point: invoke the binders, then call next.
//
// route must be a *routing.Route obtained from the request's route resolver
// (M11 wires this via foundation.Request.SetRouteResolver). Returning an error
// from the binder short-circuits the pipeline; if the route declared a
// missing handler via [routing.Route.Missing] the wrapping pipeline catches
// the error and invokes it instead.
func (s *SubstituteBindings) Handle(request any, route any, next func(any) any) (any, error) {
	if route == nil {
		return nil, fmt.Errorf("substitute bindings: no route on request")
	}

	routingRoute, ok := route.(*routing.Route)

	if !ok {
		return nil, fmt.Errorf("substitute bindings: expected *routing.Route, got %T", route)
	}

	if err := s.Router.SubstituteBindings(routingRoute); err != nil {
		return nil, err
	}

	return next(request), nil
}

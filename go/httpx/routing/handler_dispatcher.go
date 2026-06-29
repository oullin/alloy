package routing

import (
	handlermiddleware "alloy.dev/go/httpx/handlerx/middleware"
)

// HandlerDispatcher resolves and invokes a handler method against a
// matched route, applying reflection-based parameter resolution.
type HandlerDispatcher struct {
	ResolvesRouteDependencies
	FiltersHandlerMiddleware
	container DependencyContainer
}

// NewHandlerDispatcher wires a dispatcher to a container.
func NewHandlerDispatcher(container DependencyContainer) *HandlerDispatcher {
	d := &HandlerDispatcher{container: container}
	d.ResolvesRouteDependencies.Bind(container)

	return d
}

// Dispatch invokes method on handler with the parameters resolved from the
// route. handler must be a value or pointer whose type exposes the method.
func (d *HandlerDispatcher) Dispatch(route *Route, handler any, method string) (any, error) {
	in, m, _, err := d.ResolveClassMethodDependencies(
		route.ParametersWithoutNulls(),
		handler,
		method,
		route.ParameterNames(),
	)

	if err != nil {
		return nil, err
	}

	out := m.Func.Call(in)

	return packReturn(out), nil
}

// GetMiddleware returns the handler's middleware filtered by the
// only/except options for method.
//
// If the handler satisfies [handlermiddleware.Provider] its declared list
// is returned; otherwise the handler's [Handler.GetMiddleware] entries
// (if any) are flattened.
func (d *HandlerDispatcher) GetMiddleware(handler any, method string) []any {
	var entries []handlermiddleware.Entry

	if hm, ok := handler.(handlermiddleware.Provider); ok {
		entries = hm.Middleware()
	}

	out := make([]any, 0, len(entries))

	for _, e := range entries {
		opts := map[string]any{}

		if e.Only != nil {
			opts["only"] = e.Only
		}

		if e.Except != nil {
			opts["except"] = e.Except
		}

		if MethodExcludedByOptions(method, opts) {
			continue
		}

		out = append(out, e.Value)
	}

	return out
}

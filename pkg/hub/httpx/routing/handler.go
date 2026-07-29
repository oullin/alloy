package routing

import handlermiddleware "hara.sh/alloy/httpx/handlerx/middleware"

// Handler is the embeddable base class for routing handlers.
//
// In Go the recommended pattern is to compose a [Handler] into your own
// struct and override [Handler.Middleware] (via the [handlermiddleware.Provider]
// interface) to declare per-method middleware. The base type retains the
// PHP class's public surface (Middleware, GetMiddleware, CallAction) so
// dispatchers can probe for it uniformly.
type Handler struct {
	middleware []handlerMiddlewareEntry
}

type handlerMiddlewareEntry struct {
	Value   any
	Options map[string]any
}

// MiddlewareOptions is the chainable filter helper returned by Use.
type MiddlewareOptions struct{ options map[string]any }

// Only constrains the middleware to the named methods.
func (o *MiddlewareOptions) Only(methods ...string) *MiddlewareOptions {
	o.options["only"] = methods

	return o
}

// Except excludes the middleware from the named methods.
func (o *MiddlewareOptions) Except(methods ...string) *MiddlewareOptions {
	o.options["except"] = methods

	return o
}

// Use registers middleware on the handler and returns a chainable options
// object (Only/Except).
func (c *Handler) Use(middleware any) *MiddlewareOptions {
	options := map[string]any{}
	c.middleware = append(c.middleware, handlerMiddlewareEntry{
		Value:   middleware,
		Options: options,
	})

	return &MiddlewareOptions{options: options}
}

// GetMiddleware returns the middleware registered via [Handler.Use].
func (c *Handler) GetMiddleware() []handlerMiddlewareEntry { return c.middleware }

// Middleware satisfies [handlermiddleware.Provider] for handlers that only
// register middleware imperatively via Use. Override this on your own
// embedding type to return declarative entries.
func (c *Handler) Middleware() []handlermiddleware.Entry {
	out := make([]handlermiddleware.Entry, 0, len(c.middleware))

	for _, e := range c.middleware {
		m := handlermiddleware.Entry{Value: e.Value}

		if only, ok := e.Options["only"].([]string); ok {
			m.Only = only
		}

		if except, ok := e.Options["except"].([]string); ok {
			m.Except = except
		}

		out = append(out, m)
	}

	return out
}

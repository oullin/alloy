package controllers

// Middleware describes controller middleware and method filters.
type Middleware struct {
	Middleware any
	Only       []string
	Except     []string
}

// NewMiddleware constructs a Middleware value with no method filter.

// WithOnly returns a copy that applies only to the named methods.

// WithExcept returns a copy that excludes the named methods.

// HasMiddleware declares controller middleware.
type HasMiddleware interface {
	Middleware() []Middleware
}

func NewMiddleware(m any) Middleware { return Middleware{Middleware: m} }

func (m Middleware) WithOnly(methods ...string) Middleware {
	m.Only = append([]string(nil), methods...)

	return m
}

func (m Middleware) WithExcept(methods ...string) Middleware {
	m.Except = append([]string(nil), methods...)

	return m
}

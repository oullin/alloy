// Package middleware defines handler-attached middleware declarations.
package middleware

// Entry describes handler middleware and method filters.
type Entry struct {
	Value  any
	Only   []string
	Except []string
}

// Provider declares middleware attached to a handler type.
type Provider interface {
	Middleware() []Entry
}

// New constructs an Entry value with no method filter.
func New(value any) Entry { return Entry{Value: value} }

// WithOnly returns a copy that applies only to the named methods.
func (e Entry) WithOnly(methods ...string) Entry {
	e.Only = append([]string(nil), methods...)

	return e
}

// WithExcept returns a copy that excludes the named methods.
func (e Entry) WithExcept(methods ...string) Entry {
	e.Except = append([]string(nil), methods...)

	return e
}

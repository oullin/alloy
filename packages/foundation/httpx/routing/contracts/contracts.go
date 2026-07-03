package contracts

import (
	"reflect"

	"github.com/oullin/alloy/packages/foundation/httpx/routing/contracts/compiler"
)

// CallableDispatcher dispatches callable route actions.
type CallableDispatcher interface {
	Dispatch(route any, callable any) (any, error)
}

// HandlerDispatcher dispatches handler route actions.
type HandlerDispatcher interface {
	Dispatch(route any, handler any, method string) (any, error)
	GetMiddleware(handler any, method string) []any
}

// UrlRoutable exposes route key behavior for URL generation and binding hooks.
type UrlRoutable interface {
	GetRouteKey() any
	GetRouteKeyName() string
	ResolveRouteBinding(value, field string) (any, error)
	ResolveChildRouteBinding(childType, value, field string) (any, error)
}

// URLRequest is the minimum surface UrlGenerator needs from a request.
type URLRequest interface {
	Scheme() string
	Host() string
	URL() string
	Path() string
	Query(name string) string
	QueryString() string
}

// EventDispatcher is the minimal event surface the router needs.
type EventDispatcher interface {
	Dispatch(event any)
}

// ViewFactory is the minimum surface ViewHandler needs from a view layer.
type ViewFactory interface {
	Make(view string, data map[string]any) any
}

// RedirectSessionStore is the minimum session surface Redirector touches.
type RedirectSessionStore interface {
	Flash(key string, value any)
	GetOldInput(key, fallback string) string
	HasOldInput(key string) bool
	FlashInput(input map[string]any)
	Get(key string, fallback any) any
	Put(key string, value any)
}

// BindingContainer is the minimum container surface route binding needs.
type BindingContainer interface {
	Make(abstract string) (any, error)
}

// BindingResolver converts a raw URL value into a bound value.
type BindingResolver func(value string, route any) (any, error)

// ModelInstance is the surface a user-defined model must expose for binding.
type ModelInstance interface {
	ResolveRouteBinding(value, field string) (any, error)
	ResolveSoftDeletableRouteBinding(value, field string) (any, error)
	IsSoftDeletable() bool
}

// DependencyContainer resolves typed route action dependencies.
type DependencyContainer interface {
	MakeFor(t reflect.Type) (any, error)
}

// BackedEnum exposes a scalar backing value for route URLs and signatures.
type BackedEnum interface {
	BackingValue() string
}

// MatchableRoute is the surface validators need from a route.
type MatchableRoute interface {
	Methods() []string
	HttpOnly() bool
	Secure() bool
	Compiled() *compiler.CompiledRoute
}

// MatchableRequest is the surface validators need from a request.
type MatchableRequest interface {
	Method() string
	Host() string
	PathInfo() string
	Secure() bool
}

// ValidatorInterface evaluates one matching dimension for a route/request pair.
type ValidatorInterface interface {
	Matches(route MatchableRoute, request MatchableRequest) bool
}

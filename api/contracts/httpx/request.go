package httpx

// SessionStore is the minimal session interface that httpx needs.
type SessionStore interface {
	Get(key string, fallback any) any
	Put(key string, value any)
	Flash(key string, value any)
	GetOldInput(key string, fallback any) any
	HasOldInput(key string) bool
	FlashInput(values map[string]any)
	Remove(key string) any
}

// RouteResolver provides route information for the current request.
type RouteResolver interface {
	CurrentRouteName() string
	CurrentRouteAction() string
}

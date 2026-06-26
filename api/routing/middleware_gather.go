package routing

import "strings"

// MiddlewareGatherMode controls how route middleware should be exposed.
type MiddlewareGatherMode int

const (
	// MiddlewareHidden hides route middleware, matching route:list's default
	// non-verbose output.
	MiddlewareHidden MiddlewareGatherMode = iota
	// MiddlewareNames returns the route's declared middleware and group names.
	MiddlewareNames
	// MiddlewareExpanded resolves aliases and expands middleware groups.
	MiddlewareExpanded
)

// GatherRouteMiddleware returns the route middleware according to mode.
func (r *Router) GatherRouteMiddleware(route *Route, mode MiddlewareGatherMode) []any {
	if r == nil || route == nil || mode == MiddlewareHidden {
		return nil
	}

	middleware := middlewareList(route.ActionMap["middleware"])
	middleware = excludeMiddleware(middleware, middlewareList(route.ActionMap["excluded_middleware"]))

	if mode == MiddlewareNames {
		return middleware
	}

	resolver := MiddlewareNameResolver{}
	expanded := make([]any, 0, len(middleware))

	for _, item := range middleware {
		expanded = append(expanded, resolver.Resolve(item, r.middleware, r.middlewareGroups)...)
	}

	return []any(NewSortedMiddleware(r.MiddlewarePriority, expanded))
}

func middlewareList(value any) []any {
	switch v := value.(type) {
	case nil:
		return nil
	case []any:
		out := make([]any, 0, len(v))

		for _, item := range v {
			if item != nil {
				out = append(out, item)
			}
		}

		return out
	case []string:
		out := make([]any, 0, len(v))

		for _, item := range v {
			out = append(out, item)
		}

		return out
	case string:
		if v == "" {
			return nil
		}

		return []any{v}
	default:
		return []any{value}
	}
}

func excludeMiddleware(middleware []any, excluded []any) []any {
	if len(middleware) == 0 || len(excluded) == 0 {
		return middleware
	}

	out := make([]any, 0, len(middleware))

	for _, item := range middleware {
		if middlewareExcluded(item, excluded) {
			continue
		}

		out = append(out, item)
	}

	return out
}

func middlewareExcluded(item any, excluded []any) bool {
	itemName, itemOK := item.(string)

	if itemOK {
		itemName = middlewareBaseName(itemName)
	}

	for _, excludedItem := range excluded {
		if item == excludedItem {
			return true
		}

		if excludedName, ok := excludedItem.(string); ok && itemOK && middlewareBaseName(excludedName) == itemName {
			return true
		}
	}

	return false
}

func middlewareBaseName(name string) string {
	if idx := strings.Index(name, ":"); idx >= 0 {
		return name[:idx]
	}

	return name
}

package routing

import (
	"strings"

	"github.com/oullin/alloy/pkg/hub/httpx/routing/matching"
)

// CompiledRouteCollection is the production-mode counterpart to
// [RouteCollection]. In the upstream framework/Symfony it is backed by a dumped matcher; in
// the Go port it stores pre-bound route metadata in a slice and delegates
// matching to the same compiled regexes.
//
// dumped matcher data, opaque to consumers) and an "attributes" payload that
// the dumper produces alongside it. In Go we store a slice of Routes that the
// router built from the cached form; M11 will provide a real cache loader.
type CompiledRouteCollection struct {
	routeCollectionMatcher
	routes     []*Route
	nameList   map[string]*Route
	actionList map[string]*Route

	// index maps each HTTP verb to a bucket holding that verb's routes in
	// registration order plus the precomputed match plans (static exact map and
	// dynamic prefix gates). It turns Get(method) into an O(1) map read and lets
	// the matcher skip regex for static routes. Built once at construction and
	// kept in sync by Add so routes registered after the first dispatch are
	// still visible to the matcher.
	index map[string]*methodBucket
}

// NewCompiledRouteCollection builds a compiled collection from the supplied
// routes. The "attributes" map is accepted for parity with the PHP signature
// but is currently unused; M11's cache loader will populate it.
func NewCompiledRouteCollection(routes []*Route, _ map[string]any) *CompiledRouteCollection {
	c := &CompiledRouteCollection{
		routes:     append([]*Route(nil), routes...),
		nameList:   map[string]*Route{},
		actionList: map[string]*Route{},
		index:      map[string]*methodBucket{},
	}
	c.RefreshNameLookups()
	c.RefreshActionLookups()
	c.rebuildMethodIndex()

	return c
}

// rebuildMethodIndex rebuilds the per-verb match index from the current route
// slice, preserving registration order within each verb.
func (c *CompiledRouteCollection) rebuildMethodIndex() {
	c.index = make(map[string]*methodBucket, len(c.index))

	for _, r := range c.routes {
		c.indexRouteByMethod(r)
	}
}

// indexRouteByMethod appends a single route to each of its verb buckets and
// refreshes the bucket's fast-path metadata.
func (c *CompiledRouteCollection) indexRouteByMethod(route *Route) {
	for _, m := range route.Methods() {
		b := c.index[m]

		if b == nil {
			b = &methodBucket{}
			c.index[m] = b
		}

		b.add(route)
	}
}

// Add appends a route and updates the lookup tables.
func (c *CompiledRouteCollection) Add(route *Route) *Route {
	c.routes = append(c.routes, route)
	c.indexRouteByMethod(route)

	if name := route.GetName(); name != "" {
		if _, ok := c.nameList[name]; !ok {
			c.nameList[name] = route
		}
	}

	if v, ok := route.ActionMap["handler"]; ok {
		if s, ok := v.(string); ok && s != "" {
			s = strings.TrimLeft(s, `\`)

			if _, ok := c.actionList[s]; !ok {
				c.actionList[s] = route
			}
		}
	}

	return route
}

// RefreshNameLookups rebuilds the name index.
func (c *CompiledRouteCollection) RefreshNameLookups() {
	c.nameList = map[string]*Route{}

	for _, r := range c.routes {
		if name := r.GetName(); name != "" {
			if _, ok := c.nameList[name]; !ok {
				c.nameList[name] = r
			}
		}
	}
}

// RefreshActionLookups rebuilds the action index.
func (c *CompiledRouteCollection) RefreshActionLookups() {
	c.actionList = map[string]*Route{}

	for _, r := range c.routes {
		if v, ok := r.ActionMap["handler"]; ok {
			if s, ok := v.(string); ok && s != "" {
				s = strings.TrimLeft(s, `\`)

				if _, ok := c.actionList[s]; !ok {
					c.actionList[s] = r
				}
			}
		}
	}
}

// Match resolves the request against the compiled set. It consults the per-verb
// index (static exact map + dynamic prefix gates) instead of running a regex per
// route, but preserves the exact registration-order/first-match/fallback and
// 405-vs-404 semantics of the linear scan.
func (c *CompiledRouteCollection) Match(request matching.MatchableRequest) (*Route, error) {
	path := normalizeMatchPath(request.PathInfo())

	if matched := c.findInMethod(request.Method(), path, request); matched != nil {
		return matched.cloneForRequest().Bind(toBoundRequest(request))
	}

	others := c.alternateVerbs(path, request)

	if len(others) > 0 {
		return getRouteForMethods(request, others)
	}

	return nil, ErrRouteNotFound
}

// Get returns routes filtered by method, or all routes when method is "".
//
// The method-specific lookup is an O(1) read of the pre-built index; callers
// must treat the returned slice as read-only (same contract as the dev
// [RouteCollection.Get]).
func (c *CompiledRouteCollection) Get(method string) []*Route {
	if method == "" {
		return c.GetRoutes()
	}

	if b := c.index[method]; b != nil {
		return b.routes
	}

	return nil
}

// HasNamedRoute reports whether name is registered.
func (c *CompiledRouteCollection) HasNamedRoute(name string) bool {
	return c.nameList[name] != nil
}

// GetByName returns the named route or nil.
func (c *CompiledRouteCollection) GetByName(name string) *Route { return c.nameList[name] }

// GetByAction returns the route for the given action or nil.
func (c *CompiledRouteCollection) GetByAction(action string) *Route { return c.actionList[action] }

// GetRoutes returns a defensive copy of the internal slice.
func (c *CompiledRouteCollection) GetRoutes() []*Route {
	out := make([]*Route, len(c.routes))
	copy(out, c.routes)

	return out
}

// GetRoutesByMethod groups routes by HTTP verb.
func (c *CompiledRouteCollection) GetRoutesByMethod() map[string][]*Route {
	out := map[string][]*Route{}

	for _, r := range c.routes {
		for _, m := range r.Methods() {
			out[m] = append(out[m], r)
		}
	}

	return out
}

// GetRoutesByName returns the name index.
func (c *CompiledRouteCollection) GetRoutesByName() map[string]*Route {
	out := make(map[string]*Route, len(c.nameList))

	for k, v := range c.nameList {
		out[k] = v
	}

	return out
}

// Count reports the total number of routes.
func (c *CompiledRouteCollection) Count() int { return len(c.routes) }

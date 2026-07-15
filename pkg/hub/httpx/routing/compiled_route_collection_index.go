package routing

import (
	"net/url"
	"strings"

	"github.com/oullin/alloy/pkg/hub/httpx/routing/matching"
)

// methodBucket holds every route registered for one HTTP verb, in registration
// order, alongside precomputed match plans that let the matcher avoid the
// per-request regex scan.
//
// The bucket preserves the linear scan's semantics exactly: routes are still
// iterated in registration order and the first non-fallback match wins, with
// fallbacks deferred. The optimization only changes *how* each candidate is
// tested (string equality for static paths, a prefix gate before the regex for
// dynamic paths), never the order in which candidates are considered.
type methodBucket struct {
	routes []*Route
	plans  []routePlan

	// exact resolves a fully static, unconstrained path to its route in O(1).
	// It is only consulted when fastStatic is true (see below), so it never
	// shadows an earlier-registered dynamic route.
	exact map[string]*Route

	// fastStatic is true when every route in the bucket is a static path with
	// no host/scheme constraint and is not a fallback. In that case a single
	// exact-map lookup reproduces the linear scan's result, because no dynamic
	// route can precede a static one and change the first match.
	fastStatic bool
}

// routePlan is the precomputed matching strategy for a single route, parallel
// to methodBucket.routes.
type routePlan struct {
	static bool   // path has no variables: matches by exact string equality
	exact  string // normalized static path (valid when static)
	prefix string // staticPrefix used to gate the regex for dynamic routes
}

// add appends a route to the bucket and incrementally maintains its fast-path
// metadata in O(1). Once any route disqualifies the bucket (dynamic, fallback,
// or constrained), fastStatic can never become true again under append-only
// registration, so a disqualified bucket skips all bookkeeping. The resulting
// state is identical to recomputing from scratch on every add.
func (b *methodBucket) add(route *Route) {
	b.routes = append(b.routes, route)
	plan := planFor(route)
	b.plans = append(b.plans, plan)

	// Already disqualified by an earlier route: false is sticky.
	if len(b.routes) > 1 && !b.fastStatic {
		return
	}

	if !plan.static || route.IsFallback || !routeUnconstrained(route) {
		b.fastStatic = false
		b.exact = nil

		return
	}

	if b.exact == nil {
		b.exact = make(map[string]*Route)
	}

	b.fastStatic = true

	// First registration wins for a duplicate path, mirroring the linear scan.
	if _, ok := b.exact[plan.exact]; !ok {
		b.exact[plan.exact] = route
	}
}

// find returns the winning route (uncloned) for the normalized path, honoring
// registration order and fallback deferral, or nil.
func (b *methodBucket) find(path string, request matching.MatchableRequest) *Route {
	if b.fastStatic {
		return b.exact[path]
	}

	var fallback *Route

	for i, route := range b.routes {
		if !b.plans[i].matches(route, path, request) {
			continue
		}

		if route.IsFallback {
			if fallback == nil {
				fallback = route
			}

			continue
		}

		return route
	}

	return fallback
}

// matches reports whether the route matches the request, using the cheap plan
// (equality/prefix gate) before any regex, then the scheme and host validators.
func (p routePlan) matches(route *Route, path string, request matching.MatchableRequest) bool {
	if p.static {
		if path != p.exact {
			return false
		}
	} else {
		if p.prefix != "" && !strings.HasPrefix(path, p.prefix) {
			return false
		}

		if !(matching.UriValidator{}).Matches(route, request) {
			return false
		}
	}

	if !(matching.SchemeValidator{}).Matches(route, request) {
		return false
	}

	return (matching.HostValidator{}).Matches(route, request)
}

// findInMethod resolves the request within a single verb's bucket and returns
// the matched route (uncloned), or nil.
func (c *CompiledRouteCollection) findInMethod(method, path string, request matching.MatchableRequest) *Route {
	b := c.index[method]

	if b == nil {
		return nil
	}

	return b.find(path, request)
}

// alternateVerbs returns the verbs (other than the request's own) that respond
// to the path, used to distinguish a 405 from a 404. Each verb is gated by a
// cheap bucket-existence check so an empty verb costs nothing.
func (c *CompiledRouteCollection) alternateVerbs(path string, request matching.MatchableRequest) []string {
	current := request.Method()
	out := make([]string, 0, len(HTTPVerbs))

	for _, m := range HTTPVerbs {
		if m == current {
			continue
		}

		if b := c.index[m]; b != nil && len(b.routes) > 0 && b.find(path, request) != nil {
			out = append(out, m)
		}
	}

	return out
}

// planFor builds the match plan for a route, compiling it if necessary.
func planFor(route *Route) routePlan {
	compiled := route.Compiled()

	if compiled == nil {
		return routePlan{}
	}

	prefix := compiled.StaticPrefix()

	if len(compiled.PathVariables()) == 0 {
		return routePlan{static: true, exact: prefix, prefix: prefix}
	}

	return routePlan{prefix: prefix}
}

// routeUnconstrained reports whether a route's match reduces to path+method
// only (no host or scheme constraint), so an exact-path hit is a definite match.
func routeUnconstrained(route *Route) bool {
	if route.HttpOnly() || route.Secure() {
		return false
	}

	c := route.Compiled()

	return c == nil || c.HostRegex() == ""
}

// normalizeMatchPath mirrors matching.UriValidator's request-path normalization
// so exact/prefix comparisons see the same value the regex would.
func normalizeMatchPath(path string) string {
	path = strings.TrimRight(path, "/")

	if path == "" {
		path = "/"
	}

	if decoded, err := url.PathUnescape(path); err == nil {
		path = decoded
	}

	return path
}

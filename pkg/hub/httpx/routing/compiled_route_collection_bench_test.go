package routing

import (
	"fmt"
	"testing"
)

// benchRoutes builds a realistic route set: a mix of fully static routes and
// dynamic routes spread across several HTTP verbs, mirroring a mid-sized API.
func benchRoutes(static, dynamic int) []*Route {
	routes := make([]*Route, 0, static+dynamic)
	verbs := []string{"GET", "POST", "PUT", "PATCH", "DELETE"}

	for i := 0; i < static; i++ {
		verb := verbs[i%len(verbs)]
		routes = append(routes, NewRoute(verb, fmt.Sprintf("/api/v1/resource%d/list", i), func() {}))
	}

	for i := 0; i < dynamic; i++ {
		verb := verbs[i%len(verbs)]
		routes = append(routes, NewRoute(verb, fmt.Sprintf("/api/v1/entity%d/{id}", i), func() {}))
	}

	return routes
}

func benchmarkMatch(b *testing.B, req fakeRequest, static, dynamic int) {
	c := NewCompiledRouteCollection(benchRoutes(static, dynamic), nil)
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = c.Match(req)
	}
}

// BenchmarkMatchStaticHit measures the common case: a fully static route that
// lands near the end of the registration list.
func BenchmarkMatchStaticHit(b *testing.B) {
	benchmarkMatch(b, fakeRequest{method: "GET", path: "/api/v1/resource195/list"}, 200, 40)
}

// BenchmarkMatchDynamicHit measures a dynamic (parametrised) route hit.
func BenchmarkMatchDynamicHit(b *testing.B) {
	benchmarkMatch(b, fakeRequest{method: "GET", path: "/api/v1/entity35/42"}, 200, 40)
}

// BenchmarkMatchNotFound measures the 404 path (no route matches any verb).
func BenchmarkMatchNotFound(b *testing.B) {
	benchmarkMatch(b, fakeRequest{method: "GET", path: "/nope/does/not/exist"}, 200, 40)
}

// BenchmarkRouteMatches exercises Route.Matches directly (the per-candidate
// validator loop used by the dev RouteCollection). It documents the singleton
// validator set removing the per-call slice allocation.
func BenchmarkRouteMatches(b *testing.B) {
	route := NewRoute("GET", "/api/v1/entity/{id}", func() {})
	req := fakeRequest{method: "GET", path: "/api/v1/entity/42"}
	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		route.Matches(req, true)
	}
}

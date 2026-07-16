package routing

import "testing"

func TestCompiledRouteCollection_MethodIndex(t *testing.T) {
	t.Run("get_returns_routes_for_method_in_registration_order", func(t *testing.T) {
		a := NewRoute("GET", "/a", func() {})
		b := NewRoute("GET", "/b", func() {})
		p := NewRoute("POST", "/a", func() {})
		c := NewCompiledRouteCollection([]*Route{a, b, p}, nil)

		got := c.Get("GET")

		if len(got) != 2 || got[0] != a || got[1] != b {
			t.Fatalf("Get(GET) = %v, want [a b] in order", got)
		}

		if post := c.Get("POST"); len(post) != 1 || post[0] != p {
			t.Fatalf("Get(POST) = %v, want [p]", post)
		}
	})

	// The index is built at construction; routes added afterwards (including
	// after the first dispatch) must remain visible to the matcher.
	t.Run("route_registered_after_dispatch_is_matchable", func(t *testing.T) {
		c := NewCompiledRouteCollection([]*Route{NewRoute("GET", "/users", func() {})}, nil)

		// First dispatch primes any lazily-observed state.
		if _, err := c.Match(fakeRequest{method: "GET", path: "/users"}); err != nil {
			t.Fatalf("initial match: %v", err)
		}

		// Miss for a not-yet-registered path/verb.
		if _, err := c.Match(fakeRequest{method: "GET", path: "/posts"}); err == nil {
			t.Fatal("expected miss for /posts before registration")
		}

		c.Add(NewRoute("GET", "/posts", func() {}))
		c.Add(NewRoute("DELETE", "/users", func() {}))

		if _, err := c.Match(fakeRequest{method: "GET", path: "/posts"}); err != nil {
			t.Fatalf("match after Add(/posts): %v", err)
		}

		if got := c.Get("GET"); len(got) != 2 {
			t.Fatalf("Get(GET) after Add = %d routes, want 2", len(got))
		}

		if _, err := c.Match(fakeRequest{method: "DELETE", path: "/users"}); err != nil {
			t.Fatalf("match after Add(DELETE /users): %v", err)
		}
	})
}

func TestCompiledRouteCollection_IndexedMatchSemantics(t *testing.T) {
	// A dynamic route registered BEFORE an overlapping static route must still
	// win, because matching is first-registered-wins. This guards against the
	// static exact map shadowing an earlier dynamic route.
	t.Run("earlier_dynamic_route_wins_over_later_static", func(t *testing.T) {
		dynamic := NewRoute("GET", "/{slug}", func() {}).Where("slug", ".*")
		static := NewRoute("GET", "/users", func() {})
		c := NewCompiledRouteCollection([]*Route{dynamic, static}, nil)

		got, err := c.Match(fakeRequest{method: "GET", path: "/users"})

		if err != nil {
			t.Fatalf("match: %v", err)
		}

		if got.Uri != dynamic.Uri {
			t.Fatalf("got %q, want earlier dynamic route %q", got.Uri, dynamic.Uri)
		}
	})

	t.Run("earlier_static_route_wins_over_later_dynamic", func(t *testing.T) {
		static := NewRoute("GET", "/users", func() {})
		dynamic := NewRoute("GET", "/{slug}", func() {}).Where("slug", ".*")
		c := NewCompiledRouteCollection([]*Route{static, dynamic}, nil)

		got, err := c.Match(fakeRequest{method: "GET", path: "/users"})

		if err != nil {
			t.Fatalf("match: %v", err)
		}

		if got.Uri != static.Uri {
			t.Fatalf("got %q, want earlier static route %q", got.Uri, static.Uri)
		}
	})

	t.Run("fallback_deferred_across_static_and_dynamic", func(t *testing.T) {
		concrete := NewRoute("GET", "/users", func() {})
		fallback := NewRoute("GET", "/{any}", func() {}).Where("any", ".*").Fallback()
		c := NewCompiledRouteCollection([]*Route{fallback, concrete}, nil)

		// Even though the fallback is registered first and matches /users, the
		// concrete route must win.
		got, err := c.Match(fakeRequest{method: "GET", path: "/users"})

		if err != nil {
			t.Fatalf("match: %v", err)
		}

		if got.Uri != concrete.Uri {
			t.Fatalf("got %q, want concrete %q", got.Uri, concrete.Uri)
		}

		// A path only the fallback matches resolves to the fallback.
		got, err = c.Match(fakeRequest{method: "GET", path: "/anything"})

		if err != nil {
			t.Fatalf("match: %v", err)
		}

		if got.Uri != fallback.Uri {
			t.Fatalf("got %q, want fallback %q", got.Uri, fallback.Uri)
		}
	})

	t.Run("method_not_allowed_vs_not_found", func(t *testing.T) {
		c := NewCompiledRouteCollection([]*Route{NewRoute("GET", "/users", func() {})}, nil)

		if _, err := c.Match(fakeRequest{method: "POST", path: "/users"}); err == nil {
			t.Fatal("expected 405 error for POST /users")
		} else if _, ok := err.(*MethodNotAllowedError); !ok {
			t.Fatalf("want MethodNotAllowedError, got %T", err)
		}

		if _, err := c.Match(fakeRequest{method: "GET", path: "/missing"}); err != ErrRouteNotFound {
			t.Fatalf("want ErrRouteNotFound, got %v", err)
		}
	})

	t.Run("options_returns_allow_header", func(t *testing.T) {
		c := NewCompiledRouteCollection([]*Route{
			NewRoute("GET", "/users", func() {}),
			NewRoute("POST", "/users", func() {}),
		}, nil)

		got, err := c.Match(fakeRequest{method: "OPTIONS", path: "/users"})

		if err != nil {
			t.Fatalf("options match: %v", err)
		}

		if got == nil {
			t.Fatal("expected synthetic OPTIONS route")
		}
	})

	// A static path with a host constraint must NOT take the unconstrained
	// fast path: the host is still validated.
	t.Run("host_constrained_static_route_validates_host", func(t *testing.T) {
		c := NewCompiledRouteCollection([]*Route{
			NewRoute("GET", "/users", func() {}).Domain("api.example.com"),
		}, nil)

		if _, err := c.Match(fakeRequest{method: "GET", host: "api.example.com", path: "/users"}); err != nil {
			t.Fatalf("match with correct host: %v", err)
		}

		if _, err := c.Match(fakeRequest{method: "GET", host: "other.example.com", path: "/users"}); err == nil {
			t.Fatal("expected miss for wrong host")
		}
	})
}

func TestCompiledRouteCollection_ActionLookups(t *testing.T) {
	t.Run("test_get_by_action_normalizes_leading_backslash", func(t *testing.T) {
		route := NewRoute("GET", "/users", map[string]any{
			"handler": "\\App\\Http\\Handlers\\UserHandler@index",
		})
		c := NewCompiledRouteCollection([]*Route{route}, nil)

		if c.GetByAction("App\\Http\\Handlers\\UserHandler@index") != route {
			t.Fatal("GetByAction should normalize leading backslashes on init")
		}

		c.RefreshActionLookups()

		if c.GetByAction("App\\Http\\Handlers\\UserHandler@index") != route {
			t.Fatal("GetByAction should normalize leading backslashes after refresh")
		}
	})
}

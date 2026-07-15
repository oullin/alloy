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

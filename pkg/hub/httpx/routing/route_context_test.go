package routing

import (
	"context"
	"sync"
	"testing"
)

// TestCurrentRouteContextAccessors covers the package-level, context-scoped
// accessors that replace the deprecated process-wide Router.Current* methods.
func TestCurrentRouteContextAccessors(t *testing.T) {
	t.Run("nil context and empty context are safe and empty", func(t *testing.T) {
		//nolint:staticcheck // deliberately exercising the nil-context fallback.
		if got := CurrentRoute(nil); got != nil {
			t.Errorf("CurrentRoute(nil) = %v, want nil", got)
		}

		if got := CurrentRouteName(context.Background()); got != "" {
			t.Errorf("CurrentRouteName(bg) = %q, want empty", got)
		}

		if CurrentRouteIs(context.Background(), "*") {
			t.Error("CurrentRouteIs on empty context should be false")
		}

		if CurrentRouteUses(context.Background(), "*") {
			t.Error("CurrentRouteUses on empty context should be false")
		}
	})

	t.Run("accessors read the stored route", func(t *testing.T) {
		route := NewRoute("GET", "/users", "UserHandler@show")
		route.Name("users.index")

		ctx := WithCurrentRoute(context.Background(), route)

		if CurrentRoute(ctx) != route {
			t.Fatal("CurrentRoute did not return the stored route")
		}

		if got := CurrentRouteName(ctx); got != "users.index" {
			t.Errorf("CurrentRouteName = %q, want users.index", got)
		}

		if got := CurrentRouteAction(ctx); got != "UserHandler@show" {
			t.Errorf("CurrentRouteAction = %q, want UserHandler@show", got)
		}

		if !CurrentRouteIs(ctx, "users.*") {
			t.Error("CurrentRouteIs(users.*) should be true")
		}

		if !CurrentRouteNamed(ctx, "users.index") {
			t.Error("CurrentRouteNamed(users.index) should be true")
		}

		if !CurrentRouteUses(ctx, "UserHandler@*") {
			t.Error("CurrentRouteUses(UserHandler@*) should be true")
		}
	})

	t.Run("nil context defaults to Background in WithCurrentRoute", func(t *testing.T) {
		route := NewRoute("GET", "/x", func() {})

		//nolint:staticcheck // deliberately passing a nil parent context.
		ctx := WithCurrentRoute(nil, route)

		if CurrentRoute(ctx) != route {
			t.Fatal("WithCurrentRoute(nil, route) did not store the route")
		}
	})
}

// TestConcurrentDispatchIsRequestScoped is the concurrency proof from plan 004:
// two goroutines dispatching different routes concurrently must each observe
// their OWN route through the context accessors, never the other's. Handlers
// read the route from the per-request context the router threads through
// dispatch. Run under -race.
func TestConcurrentDispatchIsRequestScoped(t *testing.T) {
	r := NewRouter(nil, nil)

	// Context-accepting handlers return the route name they observe from their
	// own request context.
	r.Get("/alpha", func(ctx context.Context) any {
		return CurrentRouteName(ctx)
	}).Name("route.alpha")

	r.Get("/beta", func(ctx context.Context) any {
		return CurrentRouteName(ctx)
	}).Name("route.beta")

	cases := []struct {
		path string
		want string
	}{
		{"/alpha", "route.alpha"},
		{"/beta", "route.beta"},
	}

	const iterations = 500

	var wg sync.WaitGroup

	for i := 0; i < iterations; i++ {
		for _, tc := range cases {
			wg.Add(1)

			go func(path, want string) {
				defer wg.Done()

				dispatch, err := r.Dispatch(fakeRequest{method: "GET", path: path})

				if err != nil {
					t.Errorf("Dispatch(%q): %v", path, err)

					return
				}

				if got, _ := dispatch.Value.(string); got != want {
					t.Errorf("handler for %q observed route %q, want %q", path, got, want)
				}
			}(tc.path, tc.want)
		}
	}

	wg.Wait()
}

// TestZeroArgHandlersStillDispatch confirms the pre-existing zero-arg handler
// variants keep working unchanged after context threading was added.
func TestZeroArgHandlersStillDispatch(t *testing.T) {
	r := NewRouter(nil, nil)

	called := false
	r.Get("/plain", func() { called = true })
	r.Get("/value", func() any { return 7 })
	r.Get("/err", func() error { return nil })

	if _, err := r.Dispatch(fakeRequest{method: "GET", path: "/plain"}); err != nil {
		t.Fatalf("zero-arg func(): %v", err)
	}

	if !called {
		t.Error("zero-arg func() handler was not invoked")
	}

	dispatch, err := r.Dispatch(fakeRequest{method: "GET", path: "/value"})

	if err != nil {
		t.Fatalf("func() any: %v", err)
	}

	if dispatch.Value != 7 {
		t.Errorf("func() any value = %v, want 7", dispatch.Value)
	}

	if _, err := r.Dispatch(fakeRequest{method: "GET", path: "/err"}); err != nil {
		t.Fatalf("func() error: %v", err)
	}
}

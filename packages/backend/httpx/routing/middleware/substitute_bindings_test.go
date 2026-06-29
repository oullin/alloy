package middleware

import (
	"testing"

	"alloy.dev/backend/httpx/routing"
)

type fakeBindingRouter struct {
	explicit *routing.Route
}

var _ BindingRouter = (*routing.Router)(nil)

func (r *fakeBindingRouter) SubstituteBindings(route *routing.Route) error {
	r.explicit = route

	return nil
}

func TestSubstituteBindingsPassesRoutingRouteToRouter(t *testing.T) {
	router := &fakeBindingRouter{}
	middleware := New(router)
	route := routing.NewRoute("GET", "/users/{user}", func() {})

	value, err := middleware.Handle("request", route, func(request any) any {
		return request
	})

	if err != nil {
		t.Fatal(err)
	}

	if value != "request" {
		t.Fatalf("value = %v, want request", value)
	}

	if router.explicit != route {
		t.Fatal("explicit bindings did not receive route")
	}
}

func TestSubstituteBindingsRejectsNonRoutingRoute(t *testing.T) {
	_, err := New(&fakeBindingRouter{}).Handle("request", "route", func(request any) any {
		return request
	})

	if err == nil {
		t.Fatal("expected non-routing route to return an error")
	}
}

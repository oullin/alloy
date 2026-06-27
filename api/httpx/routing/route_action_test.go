package routing

import "testing"

func TestRouteAction(t *testing.T) {
	t.Run("test_parse_action_with_callable", func(t *testing.T) {
		called := false
		fn := func() { called = true }
		a, err := ParseAction("/x", fn)

		if err != nil {
			t.Fatal(err)
		}

		if a.Uses == nil {
			t.Fatal("uses nil")
		}
		// Sanity: invoke and ensure the original closure runs.
		a.Uses.(func())()

		if !called {
			t.Error("closure not invoked")
		}
	})

	t.Run("test_parse_action_with_handler_string", func(t *testing.T) {
		a, err := ParseAction("/x", "App\\Http\\Handlers\\UserHandler@show")

		if err != nil {
			t.Fatal(err)
		}

		if a.Handler != "App\\Http\\Handlers\\UserHandler@show" {
			t.Errorf("handler = %q", a.Handler)
		}
	})

	t.Run("test_parse_action_invokable_string", func(t *testing.T) {
		a, err := ParseAction("/x", "App\\Http\\Handlers\\Invokable")

		if err != nil {
			t.Fatal(err)
		}

		want := "App\\Http\\Handlers\\Invokable@Invoke"

		if a.Handler != want {
			t.Errorf("handler = %q, want %q", a.Handler, want)
		}
	})

	t.Run("test_parse_action_with_nil_returns_missing_action", func(t *testing.T) {
		a, err := ParseAction("/x", nil)

		if err != nil {
			t.Fatal(err)
		}

		if a.Uses == nil {
			t.Fatal("expected placeholder uses")
		}

		if got := a.Uses.(func() error)(); got == nil {
			t.Error("expected missing-action error")
		}
	})

	t.Run("test_route_get_handler_class", func(t *testing.T) {
		route := NewRoute("GET", "/users/{user}", "App\\Http\\Handlers\\UserHandler@show")

		if got := route.GetHandlerClass(); got != "App\\Http\\Handlers\\UserHandler" {
			t.Errorf("handler class = %q", got)
		}

		if got := route.GetActionMethod(); got != "show" {
			t.Errorf("action method = %q", got)
		}
	})

	t.Run("test_route_flush_handler", func(t *testing.T) {
		route := NewRoute("GET", "/users/{user}", "UserHandler@show")
		route.Handler = &userHandler{}
		route.FlushHandler()

		if route.Handler != nil {
			t.Errorf("handler = %v, want nil", route.Handler)
		}
	})

	t.Run("test_parse_action_with_map", func(t *testing.T) {
		a, err := ParseAction("/x", map[string]any{
			"uses":       func() {},
			"middleware": []any{"auth"},
			"as":         "users.show",
			"prefix":     "api",
		})

		if err != nil {
			t.Fatal(err)
		}

		if a.As != "users.show" || a.Prefix != "api" {
			t.Errorf("a = %+v", a)
		}

		if len(a.Middleware) != 1 || a.Middleware[0] != "auth" {
			t.Errorf("middleware = %v", a.Middleware)
		}
	})

	t.Run("test_contains_serialized_closure_is_false", func(t *testing.T) {
		if ContainsSerializedClosure(&Action{}) {
			t.Error("ContainsSerializedClosure should always return false")
		}
	})
}

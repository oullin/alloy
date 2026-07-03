package routing

import (
	"testing"

	handlermiddleware "github.com/oullin/alloy/packages/foundation/httpx/handlerx/middleware"
)

// tests/Routing/RoutingHandlerAttributeTest.php and the handler
// dispatch parts of RoutingRouteTest.

// userHandler is a fake handler used in dispatch tests.
type userHandler struct {
	Handler
	lastID  int
	lastTag string
}

// authHandler declares middleware via the middleware.Provider interface.
type authHandler struct{ Handler }

func (c *userHandler) Show(id int) string {
	c.lastID = id

	return "show"
}

func (c *userHandler) ShowTagged(tag string) string {
	c.lastTag = tag

	return "tagged:" + tag
}

func (c *authHandler) Middleware() []handlermiddleware.Entry {
	return []handlermiddleware.Entry{
		handlermiddleware.New("auth").WithExcept("Public"),
		handlermiddleware.New("verified").WithOnly("Settings"),
	}
}
func (c *authHandler) Settings() string { return "ok" }
func (c *authHandler) Public() string   { return "ok" }

func TestCallableDispatcher(t *testing.T) {
	t.Run("test_dispatch_func_with_string_param", func(t *testing.T) {
		d := NewCallableDispatcher(nil)
		r := NewRoute("GET", "/users/{user}", func(user string) string { return "u=" + user })
		_, _ = r.Bind(fakeRequest{path: "/users/alice"})
		got, err := d.Dispatch(r, r.ActionMap["uses"])

		if err != nil {
			t.Fatal(err)
		}

		if got != "u=alice" {
			t.Errorf("got %v", got)
		}
	})

	t.Run("test_dispatch_func_with_int_param", func(t *testing.T) {
		d := NewCallableDispatcher(nil)
		r := NewRoute("GET", "/users/{id}", func(id int) int { return id * 2 })
		_, _ = r.Bind(fakeRequest{path: "/users/21"})
		got, err := d.Dispatch(r, r.ActionMap["uses"])

		if err != nil {
			t.Fatal(err)
		}

		if got != 42 {
			t.Errorf("got %v, want 42", got)
		}
	})
}

func TestHandlerDispatcher(t *testing.T) {
	t.Run("test_dispatch_calls_method", func(t *testing.T) {
		d := NewHandlerDispatcher(nil)
		ctrl := &userHandler{}
		r := NewRoute("GET", "/users/{id}", "userHandler@Show")
		_, _ = r.Bind(fakeRequest{path: "/users/7"})
		got, err := d.Dispatch(r, ctrl, "Show")

		if err != nil {
			t.Fatal(err)
		}

		if got != "show" || ctrl.lastID != 7 {
			t.Errorf("got %v, lastID=%d", got, ctrl.lastID)
		}
	})

	t.Run("test_dispatch_string_param", func(t *testing.T) {
		d := NewHandlerDispatcher(nil)
		ctrl := &userHandler{}
		r := NewRoute("GET", "/tags/{tag}", "userHandler@ShowTagged")
		_, _ = r.Bind(fakeRequest{path: "/tags/api"})
		got, err := d.Dispatch(r, ctrl, "ShowTagged")

		if err != nil {
			t.Fatal(err)
		}

		if got != "tagged:api" {
			t.Errorf("got %v", got)
		}
	})

	t.Run("test_dispatch_missing_method", func(t *testing.T) {
		d := NewHandlerDispatcher(nil)
		ctrl := &userHandler{}
		r := NewRoute("GET", "/x", "userHandler@Missing")
		_, err := d.Dispatch(r, ctrl, "Missing")

		var mc *MissingHandlerMethodError

		if err == nil || !asMissing(err, &mc) {
			t.Errorf("expected MissingHandlerMethodError, got %v", err)
		}
	})

	t.Run("test_get_middleware_filters_by_only_except", func(t *testing.T) {
		d := NewHandlerDispatcher(nil)
		ctrl := &authHandler{}

		// Settings → "auth" applies (Public excluded), "verified" applies (only=Settings)
		got := d.GetMiddleware(ctrl, "Settings")

		if len(got) != 2 || got[0] != "auth" || got[1] != "verified" {
			t.Errorf("settings middleware = %v, want [auth verified]", got)
		}

		// public → "auth" excluded, "verified" not in only-list
		got = d.GetMiddleware(ctrl, "Public")

		if len(got) != 0 {
			t.Errorf("public middleware should be empty, got %v", got)
		}
	})

	t.Run("test_handler_middleware_attributes_are_inherited", func(t *testing.T) {
		type inheritedHandler struct{ Handler }

		d := NewHandlerDispatcher(nil)
		ctrl := &inheritedHandler{}
		ctrl.Use("auth").Only("Show")
		ctrl.Use("throttle").Except("Public")

		got := d.GetMiddleware(ctrl, "Show")

		if len(got) != 2 || got[0] != "auth" || got[1] != "throttle" {
			t.Errorf("show middleware = %v", got)
		}

		got = d.GetMiddleware(ctrl, "Public")

		if len(got) != 0 {
			t.Errorf("public middleware = %v, want []", got)
		}
	})

	t.Run("test_handler_middleware_attributes_are_in_declaration_order", func(t *testing.T) {
		type orderedHandler struct{ Handler }

		d := NewHandlerDispatcher(nil)
		ctrl := &orderedHandler{}
		ctrl.Use("first").Only("Show")
		ctrl.Use("second").Only("Show")

		got := d.GetMiddleware(ctrl, "Show")

		if len(got) != 2 || got[0] != "first" || got[1] != "second" {
			t.Errorf("show middleware order = %v", got)
		}
	})
}

// Helper for type-asserting the missing-method error without importing errors.
func asMissing(err error, dst **MissingHandlerMethodError) bool {
	if e, ok := err.(*MissingHandlerMethodError); ok {
		*dst = e

		return true
	}

	return false
}

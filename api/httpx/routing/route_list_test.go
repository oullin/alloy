package routing

import (
	"strings"
	"testing"
)

func TestRouterListRoutesFiltersSortsAndColumns(t *testing.T) {
	t.Parallel()

	router := NewRouter(nil, nil)
	router.Get("/users/{user:slug}", map[string]any{
		"uses":       "UserHandler@show",
		"middleware": []any{"web", "auth"},
	}).Name("users.show")
	router.Post("/posts", map[string]any{
		"uses": "PostHandler@store",
	}).Name("posts.store")

	entries := router.ListRoutes(RouteListOptions{
		Method:         "GET",
		Sort:           "name",
		MiddlewareMode: MiddlewareNames,
	})

	if len(entries) != 1 {
		t.Fatalf("expected 1 GET route, got %d", len(entries))
	}

	if entries[0].URI != "users/{user:slug}" {
		t.Fatalf("uri = %q", entries[0].URI)
	}

	if entries[0].Action != "UserHandler@show" {
		t.Fatalf("action = %q", entries[0].Action)
	}

	if got := strings.Join(entries[0].Middleware, ","); got != "web,auth" {
		t.Fatalf("middleware = %q", got)
	}

	rows := router.RouteListRows(RouteListOptions{
		Method:  "GET",
		Columns: []string{"method,uri"},
	})

	if len(rows) != 1 {
		t.Fatalf("expected 1 row, got %d", len(rows))
	}

	if len(rows[0]) != 2 || rows[0]["method"] != "GET|HEAD" || rows[0]["uri"] != "users/{user:slug}" {
		t.Fatalf("row = %#v", rows[0])
	}
}

func TestRouterListRoutesReverseDefinitionOrder(t *testing.T) {
	t.Parallel()

	router := NewRouter(nil, nil)
	router.Get("/first", func() {}).Name("first")
	router.Get("/second", func() {}).Name("second")

	entries := router.ListRoutes(RouteListOptions{
		Sort:    "definition",
		Reverse: true,
	})

	if len(entries) != 2 {
		t.Fatalf("entries = %d", len(entries))
	}

	if entries[0].Name != "second" || entries[1].Name != "first" {
		t.Fatalf("entries = %#v", entries)
	}
}

func TestRouterListRoutesMethodFilterMatchesMethodTokens(t *testing.T) {
	t.Parallel()

	router := NewRouter(nil, nil)
	router.Get("/users", map[string]any{"uses": "UserHandler@index"}).Name("users.index")

	if entries := router.ListRoutes(RouteListOptions{Method: "ET"}); len(entries) != 0 {
		t.Fatalf("partial method filter matched entries = %#v", entries)
	}

	if entries := router.ListRoutes(RouteListOptions{Method: "HEAD"}); len(entries) != 1 {
		t.Fatalf("HEAD method filter entries = %#v", entries)
	}
}

func TestRouterListRoutesJSONMiddlewareShape(t *testing.T) {
	t.Parallel()

	router := NewRouter(nil, nil)
	router.Get("/users", map[string]any{
		"uses":       "UserHandler@index",
		"middleware": []any{"auth", "verified"},
	}).Name("users.index")

	data, err := router.RouteListJSON(RouteListOptions{
		Columns:        []string{"name", "middleware"},
		MiddlewareMode: MiddlewareNames,
	})

	if err != nil {
		t.Fatalf("RouteListJSON: %v", err)
	}

	json := string(data)

	if !strings.Contains(json, `"name":"users.index"`) {
		t.Fatalf("json missing name: %s", json)
	}

	if !strings.Contains(json, `"middleware":["auth","verified"]`) {
		t.Fatalf("json missing middleware array: %s", json)
	}
}

func TestGatherRouteMiddlewareModes(t *testing.T) {
	t.Parallel()

	router := NewRouter(nil, nil)
	router.AliasMiddleware("auth", "App\\Http\\Middleware\\Authenticate")
	router.MiddlewareGroup("web", []any{"StartSession", "ShareErrors"})

	route := NewRouteRegistrar(router).
		Middleware("web", "auth", "csrf").
		WithoutMiddleware("csrf").
		Get("/dashboard", map[string]any{"uses": "DashboardHandler@index"})

	if got := router.GatherRouteMiddleware(route, MiddlewareHidden); len(got) != 0 {
		t.Fatalf("hidden middleware = %#v", got)
	}

	names := router.GatherRouteMiddleware(route, MiddlewareNames)

	if strings.Join(routeListMiddleware(names), ",") != "web,auth" {
		t.Fatalf("names = %#v", names)
	}

	expanded := routeListMiddleware(router.GatherRouteMiddleware(route, MiddlewareExpanded))

	if strings.Join(expanded, ",") != `StartSession,ShareErrors,App\Http\Middleware\Authenticate` {
		t.Fatalf("expanded = %#v", expanded)
	}
}

func TestGatherRouteMiddlewareNilRouter(t *testing.T) {
	t.Parallel()

	var router *Router
	route := NewRoute("GET", "/dashboard", map[string]any{
		"middleware": []any{"web"},
	})

	if got := router.GatherRouteMiddleware(route, MiddlewareExpanded); got != nil {
		t.Fatalf("middleware = %#v", got)
	}
}

func TestRouteListEntryNilRoute(t *testing.T) {
	t.Parallel()

	got := routeListEntry(nil, nil, MiddlewareNames)

	if got.Domain != "" || got.Method != "" || got.URI != "" || got.Name != "" || got.Action != "" || len(got.Middleware) != 0 || got.Path != "" || got.Vendor {
		t.Fatalf("entry = %#v", got)
	}
}

func TestIntStringHandlesNegativeValues(t *testing.T) {
	t.Parallel()

	if got := intString(-42); got != "-42" {
		t.Fatalf("intString(-42) = %q", got)
	}
}

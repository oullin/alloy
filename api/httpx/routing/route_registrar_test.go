package routing

import "testing"

// the resource-registration portions of RoutingRouteTest.

func TestRouteRegistrar_Fluent(t *testing.T) {
	t.Run("test_middleware_then_get", func(t *testing.T) {
		router := NewRouter(nil, nil)
		registrar := NewRouteRegistrar(router)
		route := registrar.Middleware("auth").Prefix("api").Get("/users", func() {})

		if route.Uri != "api/users" {
			t.Errorf("uri = %q", route.Uri)
		}

		mw, _ := route.ActionMap["middleware"].([]any)

		if len(mw) != 1 || mw[0] != "auth" {
			t.Errorf("middleware = %v", mw)
		}
	})

	t.Run("test_as_prefixes_name", func(t *testing.T) {
		router := NewRouter(nil, nil)
		registrar := NewRouteRegistrar(router)
		route := registrar.As("users.").Get("/users", func() {}).Name("index")

		if route.GetName() != "users.index" {
			t.Errorf("name = %q", route.GetName())
		}
	})

	t.Run("test_route_action_array_sets_handler", func(t *testing.T) {
		router := NewRouter(nil, nil)
		route := router.Get("/users", map[string]any{"uses": "UserHandler@show"})

		if route.ActionMap["handler"] != "UserHandler@show" {
			t.Errorf("handler = %v", route.ActionMap["handler"])
		}
	})

	t.Run("test_domain_attribute", func(t *testing.T) {
		router := NewRouter(nil, nil)
		registrar := NewRouteRegistrar(router)
		route := registrar.Domain("api.example.com").Get("/", func() {})

		if route.GetDomain() != "api.example.com" {
			t.Errorf("domain = %q", route.GetDomain())
		}
	})

	t.Run("test_group_namespace_prefixes_handler_action", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Group(map[string]any{"namespace": "App\\Http\\Handlers"}, func(r *Router) {
			route := r.Get("/users", map[string]any{"uses": "UserHandler@show"})

			if route.ActionMap["handler"] != "App\\Http\\Handlers\\UserHandler@show" {
				t.Errorf("handler = %v", route.ActionMap["handler"])
			}
		})
	})

	t.Run("test_group_domain_and_name_prefix_merge", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Group(map[string]any{
			"domain": "api.example.com",
			"as":     "api.",
		}, func(r *Router) {
			route := r.Get("/", func() {}).Name("index")

			if route.GetDomain() != "api.example.com" {
				t.Errorf("domain = %q", route.GetDomain())
			}

			if route.GetName() != "api.index" {
				t.Errorf("name = %q", route.GetName())
			}
		})
	})

	t.Run("test_group_handler_prefixes_string_action", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Group(map[string]any{"handler": "App\\Http\\Handlers\\UserHandler"}, func(r *Router) {
			route := r.Get("/users", "show")

			if route.ActionMap["handler"] != "App\\Http\\Handlers\\UserHandler@show" {
				t.Errorf("handler = %v", route.ActionMap["handler"])
			}
		})
	})

	t.Run("test_group_handler_does_not_override_explicit_action", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Group(map[string]any{"handler": "App\\Http\\Handlers\\UserHandler"}, func(r *Router) {
			route := r.Get("/users", "OtherHandler@show")

			if route.ActionMap["handler"] != "OtherHandler@show" {
				t.Errorf("handler = %v", route.ActionMap["handler"])
			}
		})
	})

	t.Run("test_middleware_nil_is_ignored", func(t *testing.T) {
		router := NewRouter(nil, nil)
		route := NewRouteRegistrar(router).Middleware(nil).Get("/users", func() {})

		if _, ok := route.ActionMap["middleware"]; ok {
			t.Errorf("middleware = %v, want no middleware entry", route.ActionMap["middleware"])
		}
	})

	t.Run("test_without_middleware_registration", func(t *testing.T) {
		router := NewRouter(nil, nil)
		route := NewRouteRegistrar(router).WithoutMiddleware("csrf").Get("/users", func() {})
		excluded, _ := route.ActionMap["excluded_middleware"].([]any)

		if len(excluded) != 1 || excluded[0] != "csrf" {
			t.Errorf("excluded middleware = %v", excluded)
		}
	})

	t.Run("test_scope_bindings_are_propagated", func(t *testing.T) {
		router := NewRouter(nil, nil)
		route := NewRouteRegistrar(router).ScopeBindings().Get("/users/{user}", func() {})

		if v, ok := route.ActionMap["scope_bindings"].(bool); !ok || !v {
			t.Errorf("scope_bindings = %v", route.ActionMap["scope_bindings"])
		}
	})

	t.Run("test_without_scoped_bindings_are_propagated", func(t *testing.T) {
		router := NewRouter(nil, nil)
		route := NewRouteRegistrar(router).WithoutScopedBindings().Get("/users/{user}", func() {})

		if v, ok := route.ActionMap["scope_bindings"].(bool); !ok || v {
			t.Errorf("scope_bindings = %v, want false", route.ActionMap["scope_bindings"])
		}
	})

	t.Run("test_where_constraints_are_propagated", func(t *testing.T) {
		t.Run("number", func(t *testing.T) {
			router := NewRouter(nil, nil)
			registrar := NewRouteRegistrar(router)
			registrar.WhereNumber("id")
			route := registrar.Get("/users/{id}", func() {})

			if route.Wheres["id"] != "[0-9]+" {
				t.Errorf("wheres = %v", route.Wheres)
			}
		})

		t.Run("alpha", func(t *testing.T) {
			router := NewRouter(nil, nil)
			registrar := NewRouteRegistrar(router)
			registrar.WhereAlpha("slug")
			route := registrar.Get("/posts/{slug}", func() {})

			if route.Wheres["slug"] != "[a-zA-Z]+" {
				t.Errorf("wheres = %v", route.Wheres)
			}
		})

		t.Run("alpha_numeric", func(t *testing.T) {
			router := NewRouter(nil, nil)
			registrar := NewRouteRegistrar(router)
			registrar.WhereAlphaNumeric("code")
			route := registrar.Get("/codes/{code}", func() {})

			if route.Wheres["code"] != "[a-zA-Z0-9]+" {
				t.Errorf("wheres = %v", route.Wheres)
			}
		})

		t.Run("in", func(t *testing.T) {
			router := NewRouter(nil, nil)
			registrar := NewRouteRegistrar(router)
			registrar.WhereIn("role", []string{"admin", "user"})
			route := registrar.Get("/roles/{role}", func() {})

			if route.Wheres["role"] != "admin|user" {
				t.Errorf("wheres = %v", route.Wheres)
			}
		})
	})
}

func TestResourceRegistrar(t *testing.T) {
	t.Run("test_register_emits_seven_routes", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Resource("users", "UserHandler", nil).Register()
		count := router.GetRoutes().(*RouteCollection).Count()

		if count != 7 {
			t.Errorf("count = %d, want 7", count)
		}
	})

	t.Run("test_only_filters_actions", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Resource("users", "UserHandler", nil).Only("index", "show", "destroy").Register()
		count := router.GetRoutes().(*RouteCollection).Count()

		if count != 3 {
			t.Errorf("count = %d, want 3", count)
		}

		assertNamedRoutes(t, router, []string{"users.index", "users.show", "users.destroy"}, nil)
	})

	t.Run("test_except_filters_actions", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Resource("users", "UserHandler", nil).Except("index", "create", "store", "show", "edit").Register()
		count := router.GetRoutes().(*RouteCollection).Count()

		if count != 2 {
			t.Errorf("count = %d, want 2", count)
		}

		assertNamedRoutes(t, router, []string{"users.update", "users.destroy"}, []string{"users.index", "users.show"})
	})

	t.Run("test_only_and_except_filters_actions", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Resource("users", "UserHandler", nil).
			Only("index", "show", "destroy").
			Except("destroy").
			Register()

		if count := router.GetRoutes().(*RouteCollection).Count(); count != 2 {
			t.Errorf("count = %d, want 2", count)
		}

		assertNamedRoutes(t, router, []string{"users.index", "users.show"}, []string{"users.destroy"})
	})

	t.Run("test_api_resource_excludes_create_edit", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.ApiResource("users", "UserHandler", nil).Register()
		count := router.GetRoutes().(*RouteCollection).Count()

		if count != 5 {
			t.Errorf("count = %d, want 5", count)
		}
	})

	t.Run("test_api_resource_except_keeps_api_exclusions", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.ApiResource("users", "UserHandler", nil).Except("index", "show", "store").Register()

		if count := router.GetRoutes().(*RouteCollection).Count(); count != 2 {
			t.Errorf("count = %d, want 2", count)
		}

		assertNamedRoutes(t, router,
			[]string{"users.update", "users.destroy"},
			[]string{"users.index", "users.store", "users.show", "users.create", "users.edit"},
		)
	})

	t.Run("test_api_resources_with_except_option", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.ApiResources(map[string]string{
			"resource-one": "OneHandler",
			"resource-two": "TwoHandler",
		}, map[string]any{"except": []string{"create", "show"}})

		if count := router.GetRoutes().(*RouteCollection).Count(); count != 8 {
			t.Errorf("count = %d, want 8", count)
		}

		for _, resource := range []string{"resource-one", "resource-two"} {
			assertNamedRoutes(t, router,
				[]string{resource + ".index", resource + ".store", resource + ".update", resource + ".destroy"},
				[]string{resource + ".create", resource + ".show", resource + ".edit"},
			)
		}
	})

	t.Run("test_api_resources_with_only_option", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.ApiResources(map[string]string{
			"resource-one": "OneHandler",
			"resource-two": "TwoHandler",
		}, map[string]any{"only": []string{"index", "show"}})

		if count := router.GetRoutes().(*RouteCollection).Count(); count != 4 {
			t.Errorf("count = %d, want 4", count)
		}

		for _, resource := range []string{"resource-one", "resource-two"} {
			assertNamedRoutes(t, router,
				[]string{resource + ".index", resource + ".show"},
				[]string{resource + ".store", resource + ".update", resource + ".destroy", resource + ".create", resource + ".edit"},
			)
		}
	})

	t.Run("test_resources_without_option_registers_each_resource", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Resources(map[string]string{
			"users": "UserHandler",
			"posts": "PostHandler",
		}, nil)

		if count := router.GetRoutes().(*RouteCollection).Count(); count != 14 {
			t.Errorf("count = %d, want 14", count)
		}

		assertNamedRoutes(t, router, []string{"users.index", "users.destroy", "posts.index", "posts.destroy"}, nil)
	})

	t.Run("test_show_uri_uses_singular_param", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Resource("users", "UserHandler", nil).Only("show").Register()
		routes := router.GetRoutes().GetRoutes()

		if routes[0].Uri != "users/{user}" {
			t.Errorf("uri = %q", routes[0].Uri)
		}
	})

	t.Run("test_route_names_default", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Resource("users", "UserHandler", nil).Only("index").Register()
		route := router.GetRoutes().GetRoutes()[0]

		if route.GetName() != "users.index" {
			t.Errorf("name = %q", route.GetName())
		}
	})

	t.Run("test_registered_resource_routes_are_returned_as_collection", func(t *testing.T) {
		router := NewRouter(nil, nil)
		resource := router.Resource("users", "UserHandler", nil).Register()

		if resource.Count() != 7 {
			t.Errorf("resource count = %d, want 7", resource.Count())
		}

		assertNamedRoutes(t, router,
			[]string{"users.index", "users.create", "users.store", "users.show", "users.edit", "users.update", "users.destroy"},
			nil,
		)
	})

	t.Run("test_singleton_emits_three_routes", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Singleton("profile", "ProfileHandler", nil).Register()
		count := router.GetRoutes().(*RouteCollection).Count()

		if count != 3 {
			t.Errorf("count = %d, want 3", count)
		}
	})

	t.Run("test_singleton_creatable_adds_create_store_destroy", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Singleton("profile", "ProfileHandler", nil).Creatable().Register()
		count := router.GetRoutes().(*RouteCollection).Count()

		if count != 6 {
			t.Errorf("count = %d, want 6", count)
		}
	})

	t.Run("test_singleton_creatable_can_exclude_destroy", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Singleton("profile", "ProfileHandler", nil).Creatable().Except("destroy").Register()

		if count := router.GetRoutes().(*RouteCollection).Count(); count != 5 {
			t.Errorf("count = %d, want 5", count)
		}

		assertNamedRoutes(t, router,
			[]string{"profile.create", "profile.store", "profile.show", "profile.edit", "profile.update"},
			[]string{"profile.destroy"},
		)
	})

	t.Run("test_api_singleton_excludes_edit", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.ApiSingleton("profile", "ProfileHandler", nil).Register()

		if count := router.GetRoutes().(*RouteCollection).Count(); count != 2 {
			t.Errorf("count = %d, want 2", count)
		}

		assertNamedRoutes(t, router, []string{"profile.show", "profile.update"}, []string{"profile.edit"})
	})

	t.Run("test_creatable_api_singleton_excludes_create_edit", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.ApiSingleton("profile", "ProfileHandler", nil).Creatable().Register()

		if count := router.GetRoutes().(*RouteCollection).Count(); count != 4 {
			t.Errorf("count = %d, want 4", count)
		}

		assertNamedRoutes(t, router,
			[]string{"profile.store", "profile.show", "profile.update", "profile.destroy"},
			[]string{"profile.create", "profile.edit"},
		)
	})

	t.Run("test_creatable_api_singleton_can_exclude_destroy", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.ApiSingleton("profile", "ProfileHandler", nil).Creatable().Except("destroy").Register()

		if count := router.GetRoutes().(*RouteCollection).Count(); count != 3 {
			t.Errorf("count = %d, want 3", count)
		}

		assertNamedRoutes(t, router,
			[]string{"profile.store", "profile.show", "profile.update"},
			[]string{"profile.create", "profile.edit", "profile.destroy"},
		)
	})

	t.Run("test_nested_resource_uri", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Resource("users.posts", "PostHandler", nil).Only("show").Register()
		route := router.GetRoutes().GetRoutes()[0]

		if route.Uri != "users/{user}/posts/{post}" {
			t.Errorf("uri = %q", route.Uri)
		}
	})

	t.Run("test_resource_parameters_override_wildcards", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Resource("users.posts", "PostHandler", nil).
			Parameters(map[string]string{"users": "account", "posts": "article"}).
			Only("show").
			Register()
		route := router.GetRoutes().GetRoutes()[0]

		if route.Uri != "users/{account}/posts/{article}" {
			t.Errorf("uri = %q", route.Uri)
		}
	})

	t.Run("test_resource_middleware_is_applied", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Resource("users", "UserHandler", nil).Middleware("auth").Only("index").Register()
		route := router.GetRoutes().GetRoutes()[0]

		mw, _ := route.ActionMap["middleware"].([]any)

		if len(mw) != 1 || mw[0] != "auth" {
			t.Errorf("middleware = %v", mw)
		}
	})

	t.Run("test_singleton_destroyable_adds_destroy_route", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Singleton("profile", "ProfileHandler", nil).Destroyable().Register()

		if count := router.GetRoutes().(*RouteCollection).Count(); count != 4 {
			t.Errorf("count = %d, want 4", count)
		}
	})

	t.Run("test_api_singleton_destroyable_adds_destroy_route", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.ApiSingleton("profile", "ProfileHandler", nil).Destroyable().Register()

		if count := router.GetRoutes().(*RouteCollection).Count(); count != 3 {
			t.Errorf("count = %d, want 3", count)
		}
	})

	t.Run("test_singleton_can_be_only_creatable", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Singleton("profile", "ProfileHandler", nil).Creatable().Only("create", "store").Register()

		if count := router.GetRoutes().(*RouteCollection).Count(); count != 2 {
			t.Errorf("count = %d, want 2", count)
		}

		assertNamedRoutes(t, router, []string{"profile.create", "profile.store"}, []string{"profile.show"})
	})

	t.Run("test_api_singleton_can_be_only_creatable", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.ApiSingleton("profile", "ProfileHandler", nil).Creatable().Only("store").Register()

		if count := router.GetRoutes().(*RouteCollection).Count(); count != 1 {
			t.Errorf("count = %d, want 1", count)
		}

		assertNamedRoutes(t, router, []string{"profile.store"}, []string{"profile.show", "profile.create"})
	})

	t.Run("test_singleton_rejects_unsupported_methods", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.Singleton("profile", "ProfileHandler", nil).Only("index", "store", "create", "destroy").Register()

		if count := router.GetRoutes().(*RouteCollection).Count(); count != 0 {
			t.Errorf("count = %d, want 0", count)
		}

		router.ApiSingleton("account", "AccountHandler", nil).Only("index", "store", "create", "destroy").Register()

		if count := router.GetRoutes().(*RouteCollection).Count(); count != 0 {
			t.Errorf("count after api singleton = %d, want 0", count)
		}
	})

	t.Run("test_api_singleton_can_explicitly_include_edit", func(t *testing.T) {
		router := NewRouter(nil, nil)
		router.ApiSingleton("profile", "ProfileHandler", nil).Only("edit").Register()

		if count := router.GetRoutes().(*RouteCollection).Count(); count != 1 {
			t.Errorf("count = %d, want 1", count)
		}

		assertNamedRoutes(t, router, []string{"profile.edit"}, []string{"profile.show", "profile.update"})
	})
}

func assertNamedRoutes(t *testing.T, router *Router, present []string, absent []string) {
	t.Helper()

	for _, name := range present {
		if !router.GetRoutes().HasNamedRoute(name) {
			t.Errorf("expected route %q to be registered", name)
		}
	}

	for _, name := range absent {
		if router.GetRoutes().HasNamedRoute(name) {
			t.Errorf("expected route %q to be absent", name)
		}
	}
}

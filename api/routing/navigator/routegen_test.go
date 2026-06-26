package navigator_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oullin/alloy/routing"
	"github.com/oullin/alloy/routing/navigator"
)

// ─────────────────────────────────────────────────────────────────────────────
// Helpers
// ─────────────────────────────────────────────────────────────────────────────

// generateTo runs Generate() into a temp directory and returns the base path.
func generateTo(t *testing.T, routes []*navigator.RouteInfo, opts navigator.Options) string {
	t.Helper()

	dir := t.TempDir()
	opts.Path = dir

	if err := navigator.Generate(routes, opts); err != nil {
		t.Fatalf("Generate: %v", err)
	}

	return dir
}

// readFile reads a file relative to base and fails the test if absent.
func readFile(t *testing.T, base, rel string) string {
	t.Helper()

	b, err := os.ReadFile(filepath.Join(base, filepath.FromSlash(rel)))

	if err != nil {
		t.Fatalf("readFile(%q): %v", rel, err)
	}

	return string(b)
}

// assertContains checks that content contains the expected substring.
func assertContains(t *testing.T, content, want string) {
	t.Helper()

	if !strings.Contains(content, want) {
		t.Errorf("expected content to contain:\n  %q\ngot:\n%s", want, content)
	}
}

// assertNotContains checks that content does NOT contain the given substring.
func assertNotContains(t *testing.T, content, unwanted string) {
	t.Helper()

	if strings.Contains(content, unwanted) {
		t.Errorf("expected content NOT to contain %q", unwanted)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Workbench route fixtures (mirror Expose workbench)
// ─────────────────────────────────────────────────────────────────────────────

func postControllerRoutes() []*navigator.RouteInfo {
	base := "App\\Http\\Controllers\\PostController"

	return []*navigator.RouteInfo{
		{URI: "/posts", Methods: []string{"get", "head"}, Controller: base + "@index"},
		{URI: "/posts/create", Methods: []string{"get", "head"}, Controller: base + "@create"},
		{URI: "/posts", Methods: []string{"post"}, Controller: base + "@store"},
		{
			URI: "/posts/{post}", Methods: []string{"get", "head"},
			Controller: base + "@show",
			Params:     []navigator.Param{{Name: "post"}},
		},
		{
			URI: "/posts/{post}/edit", Methods: []string{"get", "head"},
			Controller: base + "@edit",
			Params:     []navigator.Param{{Name: "post"}},
		},
		{
			URI: "/posts/{post}", Methods: []string{"put", "patch"},
			Controller: base + "@update",
			Params:     []navigator.Param{{Name: "post"}},
		},
		{
			URI: "/posts/{post}", Methods: []string{"delete"},
			Controller: base + "@destroy",
			Params:     []navigator.Param{{Name: "post"}},
		},
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// PostController tests
// ─────────────────────────────────────────────────────────────────────────────

func TestPostControllerGeneration(t *testing.T) {
	t.Parallel()

	dir := generateTo(t, postControllerRoutes(), navigator.Options{})
	content := readFile(t, dir, "actions/App/Http/Controllers/PostController.ts")

	// index — no params, GET
	assertContains(t, content, `export const index`)
	assertContains(t, content, `url: index.url(options)`)
	assertContains(t, content, `method: "get"`)
	assertContains(t, content, `index.definition = {`)
	assertContains(t, content, `methods: ["get","head"]`)
	assertContains(t, content, `url: "/posts"`)
	assertContains(t, content, `index.url = `)
	assertContains(t, content, `index.get = `)
	assertContains(t, content, `index.head = `)

	// show — one required param
	assertContains(t, content, `export const show`)
	assertContains(t, content, `show.url = `)
	assertContains(t, content, `.replace("{post}"`)
	assertContains(t, content, `show.definition`)
	assertContains(t, content, `url: "/posts/{post}"`)

	// destroy — DELETE
	assertContains(t, content, `export const destroy`)
	assertContains(t, content, `method: "delete"`)

	// update — PUT/PATCH
	assertContains(t, content, `export const update`)
	assertContains(t, content, `method: "put"`)

	// store — POST
	assertContains(t, content, `export const store`)
	assertContains(t, content, `method: "post"`)

	// Default export (controller object)
	assertContains(t, content, `export default PostController`)
}

func TestAnonymousMiddlewareClosureRoutesDoNotBlockGeneration(t *testing.T) {
	t.Parallel()

	routes := []*navigator.RouteInfo{
		{URI: "/closure", Methods: []string{"get", "head"}, Controller: "Closure"},
		{URI: "/posts", Methods: []string{"get", "head"}, Controller: "App\\Http\\Controllers\\PostController@index"},
	}

	dir := generateTo(t, routes, navigator.Options{})
	content := readFile(t, dir, "actions/App/Http/Controllers/PostController.ts")

	assertContains(t, content, `url: "/posts"`)

	if _, err := os.Stat(filepath.Join(dir, "actions/Closure.ts")); !os.IsNotExist(err) {
		t.Fatalf("closure middleware route should not create a Closure action file")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// InvokableController tests
// ─────────────────────────────────────────────────────────────────────────────

func TestInvokableControllerGeneration(t *testing.T) {
	t.Parallel()

	routes := []*navigator.RouteInfo{
		{
			URI:         "/invokable-controller",
			Methods:     []string{"get", "head"},
			Controller:  "App\\Http\\Controllers\\InvokableController@Invoke",
			IsInvokable: true,
		},
	}

	dir := generateTo(t, routes, navigator.Options{})
	content := readFile(t, dir, "actions/App/Http/Controllers/InvokableController.ts")

	// Must use default export (not a named export).
	assertContains(t, content, `export default InvokableController`)
	assertContains(t, content, `url: "/invokable-controller"`)
	assertContains(t, content, `method: "get"`)
	// Should NOT have "export const InvokableController"
	assertNotContains(t, content, `export const InvokableController`)
}

func TestInvokablePlusControllerGeneration(t *testing.T) {
	t.Parallel()

	routes := []*navigator.RouteInfo{
		{
			URI:         "/invokable-plus",
			Methods:     []string{"get", "head"},
			Controller:  "App\\Http\\Controllers\\InvokablePlusController@Invoke",
			IsInvokable: true,
		},
		{
			URI:        "/invokable-plus/download",
			Methods:    []string{"post"},
			Controller: "App\\Http\\Controllers\\InvokablePlusController@download",
		},
	}

	dir := generateTo(t, routes, navigator.Options{})
	content := readFile(t, dir, "actions/App/Http/Controllers/InvokablePlusController.ts")

	assertContains(t, content, `const InvokablePlusController = (`)
	assertContains(t, content, `const download = (`)
	assertContains(t, content, `InvokablePlusController.download = download`)
	assertContains(t, content, `export default InvokablePlusController`)
	assertNotContains(t, content, `export const InvokablePlusController`)
}

// ─────────────────────────────────────────────────────────────────────────────
// OptionalController tests
// ─────────────────────────────────────────────────────────────────────────────

func TestOptionalControllerGeneration(t *testing.T) {
	t.Parallel()

	routes := []*navigator.RouteInfo{
		{
			URI:        "/optional/{parameter?}",
			Methods:    []string{"get", "head"},
			Controller: "App\\Http\\Controllers\\OptionalController@optional",
			Params:     []navigator.Param{{Name: "parameter", Optional: true}},
		},
		{
			URI:        "/many-optional/{one?}/{two?}/{three?}",
			Methods:    []string{"get", "head"},
			Controller: "App\\Http\\Controllers\\OptionalController@manyOptional",
			Params: []navigator.Param{
				{Name: "one", Optional: true},
				{Name: "two", Optional: true},
				{Name: "three", Optional: true},
			},
		},
	}

	dir := generateTo(t, routes, navigator.Options{})
	content := readFile(t, dir, "actions/App/Http/Controllers/OptionalController.ts")

	// optional param — definition should contain {parameter?}
	assertContains(t, content, `url: "/optional/{parameter?}"`)
	// validateParameters should be called
	assertContains(t, content, `validateParameters(`)
	// Optional chaining in replace
	assertContains(t, content, `.replace("{parameter?}"`)

	// manyOptional
	assertContains(t, content, `url: "/many-optional/{one?}/{two?}/{three?}"`)
	assertContains(t, content, `validateParameters(args, ["one", "two", "three"])`)
	assertContains(t, content, `.replace("{one?}", parsedArgs.one?.toString() ?? '')`)
	assertContains(t, content, `.replace("{two?}", parsedArgs.two?.toString() ?? '')`)
	assertContains(t, content, `.replace("{three?}", parsedArgs.three?.toString() ?? '')`)
}

func TestEmptyRouteExposure(t *testing.T) {
	t.Parallel()

	routes := []*navigator.RouteInfo{
		{
			URI:        "",
			Methods:    []string{"get", "head"},
			Controller: "App\\Http\\Controllers\\EmptyRouteController@index",
		},
	}

	dir := generateTo(t, routes, navigator.Options{})
	content := readFile(t, dir, "actions/App/Http/Controllers/EmptyRouteController.ts")

	assertContains(t, content, `url: ""`)
	assertNotContains(t, content, `url: "/"`)
}

// ─────────────────────────────────────────────────────────────────────────────
// ModelBindingController tests
// ─────────────────────────────────────────────────────────────────────────────

func TestModelBindingControllerGeneration(t *testing.T) {
	t.Parallel()

	routes := []*navigator.RouteInfo{
		{
			URI:        "/users/{user}",
			Methods:    []string{"get", "head"},
			Controller: "App\\Http\\Controllers\\ModelBindingController@show",
			Params:     []navigator.Param{{Name: "user"}},
		},
	}

	dir := generateTo(t, routes, navigator.Options{})
	content := readFile(t, dir, "actions/App/Http/Controllers/ModelBindingController.ts")

	assertContains(t, content, `url: "/users/{user}"`)
	// Primitive shorthand support.
	assertContains(t, content, `if (typeof args === 'string' || typeof args === 'number')`)
	// Array support.
	assertContains(t, content, `if (Array.isArray(args))`)
}

func TestCamelCaseRouteParameterGeneration(t *testing.T) {
	t.Parallel()

	routes := []*navigator.RouteInfo{
		{
			URI:        "/profiles/{userProfile}",
			Methods:    []string{"get", "head"},
			Controller: "App\\Http\\Controllers\\CamelCaseRouteParameterController@show",
			Params:     []navigator.Param{{Name: "userProfile", Key: "uuid"}},
		},
	}

	dir := generateTo(t, routes, navigator.Options{})
	content := readFile(t, dir, "actions/App/Http/Controllers/CamelCaseRouteParameterController.ts")

	assertContains(t, content, `url: "/profiles/{userProfile}"`)
	assertContains(t, content, `"uuid" in args`)
	assertContains(t, content, `args.userProfile.uuid`)
	assertContains(t, content, `.replace("{userProfile}", parsedArgs.userProfile.toString())`)
}

func TestHyphenatedRouteParameterGenerationUsesSafeNames(t *testing.T) {
	t.Parallel()

	routes := []*navigator.RouteInfo{
		{
			URI:        "/profiles/{user-profile}",
			Methods:    []string{"get", "head"},
			Controller: "App\\Http\\Controllers\\HyphenatedRouteParameterController@show",
			Params:     []navigator.Param{{Name: "user-profile", Key: "uuid"}},
		},
	}

	dir := generateTo(t, routes, navigator.Options{})
	content := readFile(t, dir, "actions/App/Http/Controllers/HyphenatedRouteParameterController.ts")

	assertContains(t, content, `url: "/profiles/{user-profile}"`)
	assertContains(t, content, `{ userProfile: args }`)
	assertContains(t, content, `userProfile: typeof args.userProfile === 'object'`)
	assertContains(t, content, `args.userProfile.uuid`)
	assertContains(t, content, `.replace("{user-profile}", parsedArgs.userProfile.toString())`)
	assertNotContains(t, content, `args.user-profile`)
}

// ─────────────────────────────────────────────────────────────────────────────
// KeyController tests
// ─────────────────────────────────────────────────────────────────────────────

func TestKeyControllerGeneration(t *testing.T) {
	t.Parallel()

	routes := []*navigator.RouteInfo{
		{
			URI:        "/keys/{key}/edit",
			Methods:    []string{"get", "head"},
			Controller: "App\\Http\\Controllers\\KeyController@edit",
			Params:     []navigator.Param{{Name: "key", Key: "uuid"}},
		},
	}

	dir := generateTo(t, routes, navigator.Options{})
	content := readFile(t, dir, "actions/App/Http/Controllers/KeyController.ts")

	assertContains(t, content, `url: "/keys/{key}/edit"`)
	// Custom key resolution: check for uuid field.
	assertContains(t, content, `.uuid`)
	// definition should show url pattern.
	assertContains(t, content, `url: "/keys/{key}/edit"`)
}

// ─────────────────────────────────────────────────────────────────────────────
// DisallowedMethodNames tests
// ─────────────────────────────────────────────────────────────────────────────

func TestDisallowedMethodNamesGeneration(t *testing.T) {
	t.Parallel()

	base := "App\\Http\\Controllers\\DisallowedMethodNameController"
	routes := []*navigator.RouteInfo{
		{URI: "/disallowed/delete", Methods: []string{"get", "head"}, Controller: base + "@delete"},
		{URI: "/disallowed/404", Methods: []string{"get", "head"}, Controller: base + "@404"},
		{URI: "/disallowed/2fa", Methods: []string{"get", "head"}, Controller: base + "@2fa"},
		{URI: "/disallowed/default", Methods: []string{"get", "head"}, Controller: base + "@default"},
	}

	dir := generateTo(t, routes, navigator.Options{})
	content := readFile(t, dir, "actions/App/Http/Controllers/DisallowedMethodNameController.ts")

	// Reserved word "delete" → "deleteMethod"
	assertContains(t, content, `const deleteMethod`)
	assertContains(t, content, `url: "/disallowed/delete"`)

	// Leading number "404" → "method404"
	assertContains(t, content, `const method404`)
	assertContains(t, content, `url: "/disallowed/404"`)

	// Reserved word "default" → "defaultMethod"
	assertContains(t, content, `const defaultMethod`)

	// Controller object must have original name aliases.
	assertContains(t, content, `delete: deleteMethod`)
	assertContains(t, content, `404: method404`)

	// Default export
	assertContains(t, content, `export default DisallowedMethodNameController`)
}

func TestMethodNameCollisionGeneration(t *testing.T) {
	t.Parallel()

	routes := []*navigator.RouteInfo{
		{
			URI:        "/method-collision/{post}",
			Methods:    []string{"get", "head"},
			Controller: "App\\Http\\Controllers\\MethodNameCollisionController@options",
			Params:     []navigator.Param{{Name: "post"}},
		},
	}

	dir := generateTo(t, routes, navigator.Options{})
	content := readFile(t, dir, "actions/App/Http/Controllers/MethodNameCollisionController.ts")

	assertContains(t, content, `export const options = (args: {`)
	assertContains(t, content, `routeOptions?: RouteQueryOptions`)
	assertContains(t, content, `queryParams(routeOptions)`)
	assertNotContains(t, content, `queryParams(options)`)
}

func TestParameterNameCollisionGeneration(t *testing.T) {
	t.Parallel()

	routes := []*navigator.RouteInfo{
		{
			URI:        "/parameter-names/{args}/{options}/{parsedArgs}",
			Methods:    []string{"get", "head"},
			Controller: "App\\Http\\Controllers\\ParamaterNameController@show",
			Params: []navigator.Param{
				{Name: "args"},
				{Name: "options"},
				{Name: "parsedArgs"},
			},
		},
	}

	dir := generateTo(t, routes, navigator.Options{})
	content := readFile(t, dir, "actions/App/Http/Controllers/ParamaterNameController.ts")

	assertContains(t, content, `url: "/parameter-names/{args}/{options}/{parsedArgs}"`)
	assertContains(t, content, `.replace("{args}", parsedArgs.args.toString())`)
	assertContains(t, content, `.replace("{options}", parsedArgs.options.toString())`)
	assertContains(t, content, `.replace("{parsedArgs}", parsedArgs.parsedArgs.toString())`)
}

// ─────────────────────────────────────────────────────────────────────────────
// TwoRoutesSameAction tests
// ─────────────────────────────────────────────────────────────────────────────

func TestTwoRoutesSameActionGeneration(t *testing.T) {
	t.Parallel()

	base := "App\\Http\\Controllers\\TwoRoutesSameActionController"
	routes := []*navigator.RouteInfo{
		{URI: "/two-routes-one-action-1", Methods: []string{"get", "head"}, Controller: base + "@same"},
		{URI: "/two-routes-one-action-2", Methods: []string{"get", "head"}, Controller: base + "@same"},
	}

	dir := generateTo(t, routes, navigator.Options{})
	content := readFile(t, dir, "actions/App/Http/Controllers/TwoRoutesSameActionController.ts")

	// Keyed dictionary with URI strings as keys.
	assertContains(t, content, `"/two-routes-one-action-1"`)
	assertContains(t, content, `"/two-routes-one-action-2"`)
	assertContains(t, content, `export const same`)
}

// ─────────────────────────────────────────────────────────────────────────────
// DomainController tests
// ─────────────────────────────────────────────────────────────────────────────

func TestDomainControllerGeneration(t *testing.T) {
	t.Parallel()

	base := "App\\Http\\Controllers\\DomainController"
	routes := []*navigator.RouteInfo{
		{
			URI: "/fixed-domain/{param}", Methods: []string{"get", "head"},
			Controller: base + "@fixedDomain",
			Domain:     "example.test",
			Scheme:     "//",
			Params:     []navigator.Param{{Name: "param"}},
		},
		{
			URI: "/default-parameters-domain/{param}", Methods: []string{"get", "head"},
			Controller: base + "@defaultParametersDomain",
			Domain:     "{defaultDomain?}.au",
			Scheme:     "//",
			Params: []navigator.Param{
				{Name: "defaultDomain", Optional: true},
				{Name: "param"},
			},
		},
	}

	dir := generateTo(t, routes, navigator.Options{})
	content := readFile(t, dir, "actions/App/Http/Controllers/DomainController.ts")

	// Fixed domain URLs include the domain.
	assertContains(t, content, `//example.test/fixed-domain/{param}`)

	// Dynamic domain URL includes the domain placeholder.
	assertContains(t, content, `//{defaultDomain?}.au/default-parameters-domain/{param}`)
	assertNotContains(t, content, `:8080`)
}

// ─────────────────────────────────────────────────────────────────────────────
// Named routes tests
// ─────────────────────────────────────────────────────────────────────────────

func TestNamedRoutesGeneration(t *testing.T) {
	t.Parallel()

	routes := []*navigator.RouteInfo{
		{
			URI: "/posts/{post}/edit", Methods: []string{"get", "head"},
			Name:       "posts.edit",
			Controller: "App\\Http\\Controllers\\PostController@edit",
			Params:     []navigator.Param{{Name: "post"}},
		},
		{
			URI: "/dashboard", Methods: []string{"get", "head"},
			Name:       "dashboard",
			Controller: "App\\Http\\Controllers\\DashboardController@index",
		},
		{
			URI:         "/named-invokable-controller",
			Methods:     []string{"get", "head"},
			Name:        "invokable",
			Controller:  "App\\Http\\Controllers\\InvokableController@Invoke",
			IsInvokable: true,
		},
		{
			URI:        "/invalid-js-name",
			Methods:    []string{"get", "head"},
			Name:       "invalid_js_name",
			Controller: "App\\Http\\Controllers\\SomeController@invalidJsName",
		},
	}

	dir := generateTo(t, routes, navigator.Options{SkipActions: true})

	// posts/index.ts should export "edit"
	postsContent := readFile(t, dir, "routes/posts/index.ts")
	assertContains(t, postsContent, `export const edit`)
	assertContains(t, postsContent, `url: "/posts/{post}/edit"`)

	// root index.ts should have dashboard
	rootContent := readFile(t, dir, "routes/index.ts")
	assertContains(t, rootContent, `dashboard`)
}

func TestNamespacedAndStorageRoutesGeneration(t *testing.T) {
	t.Parallel()

	routes := []*navigator.RouteInfo{
		{
			URI:        "/admin/reports",
			Methods:    []string{"get", "head"},
			Name:       "admin::reports.index",
			Controller: "App\\Http\\Controllers\\Admin\\ReportsController@index",
		},
		{
			URI:     "/storage/{path}",
			Methods: []string{"get", "head"},
			Name:    "storage.local",
			Params:  []navigator.Param{{Name: "path"}},
		},
	}

	dir := generateTo(t, routes, navigator.Options{})
	content := readFile(t, dir, "routes/namespaced/admin/reports/index.ts")
	assertContains(t, content, `export const index`)
	assertContains(t, content, `url: "/admin/reports"`)

	storage := readFile(t, dir, "routes/storage/index.ts")
	assertContains(t, storage, `export const local`)
	assertContains(t, storage, `url: "/storage/{path}"`)
}

// ─────────────────────────────────────────────────────────────────────────────
// UrlDefaults tests
// ─────────────────────────────────────────────────────────────────────────────

func TestUrlDefaultsControllerGeneration(t *testing.T) {
	t.Parallel()

	base := "App\\Http\\Controllers\\UrlDefaultsController"
	routes := []*navigator.RouteInfo{
		{
			URI:        "/with-defaults/{locale}",
			Methods:    []string{"post"},
			Controller: base + "@onlyDefaults",
			Params: []navigator.Param{
				{Name: "locale", Optional: true, Default: "en"},
			},
			Defaults: map[string]string{"locale": "en"},
		},
	}

	dir := generateTo(t, routes, navigator.Options{})
	content := readFile(t, dir, "actions/App/Http/Controllers/UrlDefaultsController.ts")

	// Default value is used in parsedArgs.
	assertContains(t, content, `?? "en"`)
	// URI shows optional marker because locale has a default.
	assertContains(t, content, `{locale?}`)
}

func TestFromRouteCollectionKeepsMultiDigitIntegerDefaults(t *testing.T) {
	t.Parallel()

	router := routing.NewRouter(nil, nil)
	router.Get("/archive/{year}", func() {}).Defaults("year", 2026)
	routes := navigator.FromRouteCollection(router.GetRoutes(), navigator.AdapterOptions{})

	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}

	if len(routes[0].Params) != 1 {
		t.Fatalf("expected 1 param, got %d", len(routes[0].Params))
	}

	if routes[0].Params[0].Default != "2026" {
		t.Fatalf("default = %q, want 2026", routes[0].Params[0].Default)
	}
}

func TestFromRouteCollectionUsesForcedRootAndDefaults(t *testing.T) {
	t.Parallel()

	router := routing.NewRouter(nil, nil)
	router.Get("/{locale}/posts/{post:slug}", "App\\Http\\Controllers\\PostController@show").Name("posts.show")

	routes := navigator.FromRouteCollection(router.GetRoutes(), navigator.AdapterOptions{
		ForcedRoot: "https://example.test/app",
		Defaults:   map[string]string{"locale": "en"},
	})

	if len(routes) != 1 {
		t.Fatalf("expected 1 route, got %d", len(routes))
	}

	if routes[0].Domain != "example.test" {
		t.Fatalf("domain = %q", routes[0].Domain)
	}

	if routes[0].Scheme != "https://" {
		t.Fatalf("scheme = %q", routes[0].Scheme)
	}

	if routes[0].BasePath != "/app" {
		t.Fatalf("basePath = %q", routes[0].BasePath)
	}

	if routes[0].Params[0].Default != "en" || !routes[0].Params[0].Optional {
		t.Fatalf("locale param = %#v", routes[0].Params[0])
	}

	if routes[0].Params[1].Key != "slug" {
		t.Fatalf("post key = %q", routes[0].Params[1].Key)
	}
}

func TestGenerateFromRouter(t *testing.T) {
	t.Parallel()

	router := routing.NewRouter(nil, nil)
	router.Get("/posts/{post:slug}", "App\\Http\\Controllers\\PostController@show").Name("posts.show")

	dir := t.TempDir()

	if err := navigator.GenerateFromRouter(router, navigator.Options{Path: dir}, navigator.AdapterOptions{}); err != nil {
		t.Fatalf("GenerateFromRouter: %v", err)
	}

	action := readFile(t, dir, "actions/App/Http/Controllers/PostController.ts")
	named := readFile(t, dir, "routes/posts/index.ts")
	runtime := readFile(t, dir, "expose/index.ts")

	assertContains(t, action, `from './../../../../expose'`)
	assertContains(t, action, `url: "/posts/{post}"`)
	assertContains(t, action, `args.post.slug`)
	assertContains(t, named, `export const show`)
	assertContains(t, runtime, `export const queryParams`)
}

// ─────────────────────────────────────────────────────────────────────────────
// WithForm option tests
// ─────────────────────────────────────────────────────────────────────────────

func TestWithFormOption(t *testing.T) {
	t.Parallel()

	routes := []*navigator.RouteInfo{
		{
			URI: "/posts", Methods: []string{"post"},
			Controller: "App\\Http\\Controllers\\PostController@store",
		},
		{
			URI: "/posts/{post}", Methods: []string{"put", "patch"},
			Controller: "App\\Http\\Controllers\\PostController@update",
			Params:     []navigator.Param{{Name: "post"}},
		},
	}

	dir := generateTo(t, routes, navigator.Options{WithForm: true})
	content := readFile(t, dir, "actions/App/Http/Controllers/PostController.ts")

	// Form helpers must be present.
	assertContains(t, content, `.form = `)
	assertContains(t, content, `RouteFormDefinition`)
	assertContains(t, content, `action:`)
	// _method spoofing for non-GET verbs.
	assertContains(t, content, `_method`)
}

// ─────────────────────────────────────────────────────────────────────────────
// AppUrl base-path tests
// ─────────────────────────────────────────────────────────────────────────────

func TestAppURLPathPrefix(t *testing.T) {
	t.Parallel()

	routes := postControllerRoutes()
	// Apply base path — simulates APP_URL=http://localhost:8081/v2.
	for _, r := range routes {
		r.BasePath = "/v2"
	}

	dir := generateTo(t, routes, navigator.Options{})
	content := readFile(t, dir, "actions/App/Http/Controllers/PostController.ts")

	assertContains(t, content, `url: "/v2/posts"`)
	assertContains(t, content, `url: "/v2/posts/{post}"`)
}

// ─────────────────────────────────────────────────────────────────────────────
// Expose runtime utility
// ─────────────────────────────────────────────────────────────────────────────

func TestExposeRuntimeUtility(t *testing.T) {
	t.Parallel()

	dir := generateTo(t, nil, navigator.Options{SkipActions: true, SkipRoutes: true})
	content := readFile(t, dir, "expose/index.ts")

	// Must contain the core runtime functions.
	assertContains(t, content, `export const queryParams`)
	assertContains(t, content, `export const setUrlDefaults`)
	assertContains(t, content, `export const applyUrlDefaults`)
	assertContains(t, content, `export const validateParameters`)
	assertContains(t, content, `export type QueryParams`)
	assertContains(t, content, `export type RouteDefinition`)
	assertContains(t, content, `export type RouteFormDefinition`)
	assertContains(t, content, `export type RouteQueryOptions`)
}

func TestExposeRuntimeDirectoryCompatibility(t *testing.T) {
	t.Parallel()

	dir := generateTo(t, nil, navigator.Options{
		SkipActions:      true,
		SkipRoutes:       true,
		RuntimeDirectory: "routegen",
	})
	content := readFile(t, dir, "routegen/index.ts")

	assertContains(t, content, `export const queryParams`)
}

func TestExposeRuntimeDirectoryImportsUseForwardSlashes(t *testing.T) {
	t.Parallel()

	dir := generateTo(t, postControllerRoutes(), navigator.Options{
		RuntimeDirectory: `runtime\expose`,
	})
	content := readFile(t, dir, "actions/App/Http/Controllers/PostController.ts")

	assertContains(t, content, `from './../../../../runtime/expose'`)
}

func TestExposeRuntimeQueryParamsSource(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile(filepath.FromSlash("resources/expose.ts"))

	if err != nil {
		t.Fatal(err)
	}

	content := string(data)

	assertContains(t, content, `return '1';`)
	assertContains(t, content, `return '0';`)
	assertContains(t, content, "params.append(`${paramKey}[]`, getValue(v))")
	assertContains(t, content, `addNestedParams(value, paramKey, params)`)
	assertContains(t, content, `window.location.search`)
	assertContains(t, content, `params.delete(key)`)
	assertContains(t, content, `export const setUrlDefaults = (params: UrlDefaults | (() => UrlDefaults))`)
	assertContains(t, content, `export const applyUrlDefaults`)
	assertContains(t, content, `export const validateParameters`)
}

// ─────────────────────────────────────────────────────────────────────────────
// SkipActions / SkipRoutes options
// ─────────────────────────────────────────────────────────────────────────────

func TestSkipOptions(t *testing.T) {
	t.Parallel()

	routes := []*navigator.RouteInfo{
		{
			URI: "/posts", Methods: []string{"get"},
			Name:       "posts.index",
			Controller: "App\\Http\\Controllers\\PostController@index",
		},
	}

	t.Run("skip_actions", func(t *testing.T) {
		t.Parallel()
		dir := generateTo(t, routes, navigator.Options{SkipActions: true})

		if _, err := os.Stat(filepath.Join(dir, "actions")); !os.IsNotExist(err) {
			t.Error("actions/ directory should not exist when SkipActions=true")
		}
	})

	t.Run("skip_routes", func(t *testing.T) {
		t.Parallel()
		dir := generateTo(t, routes, navigator.Options{SkipRoutes: true})

		if _, err := os.Stat(filepath.Join(dir, "routes")); !os.IsNotExist(err) {
			t.Error("routes/ directory should not exist when SkipRoutes=true")
		}
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// Barrel files
// ─────────────────────────────────────────────────────────────────────────────

func TestBarrelFilesGeneration(t *testing.T) {
	t.Parallel()

	routes := []*navigator.RouteInfo{
		{URI: "/posts", Methods: []string{"get"}, Controller: "App\\Http\\Controllers\\PostController@index"},
		{URI: "/users", Methods: []string{"get"}, Controller: "App\\Http\\Controllers\\UserController@index"},
	}

	dir := generateTo(t, routes, navigator.Options{SkipRoutes: true})

	// Controller-level barrel.
	indexContent := readFile(t, dir, "actions/App/Http/Controllers/index.ts")
	assertContains(t, indexContent, `import PostController from './PostController'`)
	assertContains(t, indexContent, `import UserController from './UserController'`)

	// Top-level barrel.
	rootIndex := readFile(t, dir, "actions/index.ts")
	assertContains(t, rootIndex, `export default`)
}

func TestRepeatedNamespaceControllerGeneration(t *testing.T) {
	t.Parallel()

	routes := []*navigator.RouteInfo{
		{URI: "/admin/repeated", Methods: []string{"get"}, Controller: "App\\Http\\Controllers\\Admin\\Admin\\RepeatedNamespaceController@index"},
		{URI: "/admin/users", Methods: []string{"get"}, Controller: "App\\Http\\Controllers\\Admin\\UserController@index"},
	}

	dir := generateTo(t, routes, navigator.Options{SkipRoutes: true})
	content := readFile(t, dir, "actions/App/Http/Controllers/Admin/index.ts")

	assertContains(t, content, `import Admin from './Admin'`)
	assertContains(t, content, `import UserController from './UserController'`)
	assertContains(t, content, `const Admin = {`)
}

package routing_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"slices"
	"sort"
	"testing"

	"hara.sh/alloy/httpx/routing"
	"hara.sh/alloy/httpx/routing/compiler"
)

type routingFixture struct {
	SchemaVersion int                  `json:"schemaVersion"`
	Cases         []routingFixtureCase `json:"cases"`
}

type routingFixtureCase struct {
	ID       string               `json:"id"`
	Note     string               `json:"note"`
	Type     string               `json:"type"`
	Route    *fixtureRouteSpec    `json:"route"`
	Routes   []fixtureRouteSpec   `json:"routes"`
	Request  *fixtureRequestSpec  `json:"request"`
	Params   map[string]any       `json:"params"`
	Expected *fixtureExpectedSpec `json:"expected"`
	Error    *fixtureErrorSpec    `json:"error"`
}

type fixtureRouteSpec struct {
	Method       string            `json:"method"`
	Path         string            `json:"path"`
	Host         string            `json:"host"`
	Name         string            `json:"name"`
	Defaults     map[string]any    `json:"defaults"`
	Requirements map[string]string `json:"requirements"`
}

type fixtureRequestSpec struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Host   string `json:"host"`
	Secure bool   `json:"secure"`
}

type fixtureExpectedSpec struct {
	StaticPrefix string            `json:"staticPrefix"`
	Variables    []string          `json:"variables"`
	Name         string            `json:"name"`
	Params       map[string]string `json:"params"`
	URL          string            `json:"url"`
}

type fixtureErrorSpec struct {
	Code    string   `json:"code"`
	Allowed []string `json:"allowed"`
}

type fixtureSourceRoute struct {
	path         string
	host         string
	defaults     map[string]any
	requirements map[string]string
}

type fixtureMatchableRequest struct {
	method string
	path   string
	host   string
	secure bool
}

func (f fixtureSourceRoute) Path() string { return f.path }

func (f fixtureSourceRoute) Host() string { return f.host }

func (f fixtureSourceRoute) Requirements() map[string]string { return f.requirements }

func (f fixtureSourceRoute) HasDefault(name string) bool {
	_, ok := f.defaults[name]

	return ok
}

func (r fixtureMatchableRequest) Method() string { return r.method }

func (r fixtureMatchableRequest) PathInfo() string { return r.path }

func (r fixtureMatchableRequest) Path() string { return r.path }

func (r fixtureMatchableRequest) DecodedPath() string { return r.path }

func (r fixtureMatchableRequest) Host() string { return r.host }

func (r fixtureMatchableRequest) Secure() bool { return r.secure }

func (r fixtureMatchableRequest) Context() context.Context { return context.Background() }

func (r fixtureMatchableRequest) URL() string { return r.path }

func (r fixtureMatchableRequest) Scheme() string {
	if r.secure {
		return "https"
	}

	return "http"
}

func (r fixtureMatchableRequest) QueryString() string { return "" }

func (r fixtureMatchableRequest) Query(key string) string { return "" }

func TestRoutingConformance(t *testing.T) {
	t.Parallel()
	fixture := loadRoutingConformance(t)

	for _, tc := range fixture.Cases {
		t.Run(tc.ID, func(t *testing.T) {
			switch tc.Type {
			case "compile":
				reqs := tc.Route.Requirements

				if reqs == nil {
					reqs = map[string]string{}
				}

				defs := tc.Route.Defaults

				if defs == nil {
					defs = map[string]any{}
				}

				sr := fixtureSourceRoute{
					path:         tc.Route.Path,
					host:         tc.Route.Host,
					defaults:     defs,
					requirements: reqs,
				}

				compiled, err := compiler.Compile(sr)

				if err != nil {
					t.Fatalf("unexpected compile error: %v", err)
				}

				if tc.Expected.StaticPrefix != "" && compiled.StaticPrefix() != tc.Expected.StaticPrefix {
					t.Errorf("static prefix = %q, want %q", compiled.StaticPrefix(), tc.Expected.StaticPrefix)
				}

				if tc.Expected.Variables != nil {
					if !slices.Equal(compiled.Variables(), tc.Expected.Variables) {
						t.Errorf("variables = %v, want %v", compiled.Variables(), tc.Expected.Variables)
					}
				}

			case "match":
				collection := routing.NewRouteCollection()

				for _, rspec := range tc.Routes {
					r := routing.NewRoute(rspec.Method, rspec.Path, func() string { return "ok" })

					if rspec.Name != "" {
						r.Name(rspec.Name)
					}

					collection.Add(r)
				}

				reqHost := tc.Request.Host

				if reqHost == "" {
					reqHost = "localhost"
				}

				req := fixtureMatchableRequest{
					method: tc.Request.Method,
					path:   tc.Request.Path,
					host:   reqHost,
					secure: tc.Request.Secure,
				}

				matched, err := collection.Match(req)

				if tc.Error != nil {
					if tc.Error.Code == "ROUTE_NOT_FOUND" {
						if !errors.Is(err, routing.ErrRouteNotFound) {
							t.Fatalf("error = %v, want ErrRouteNotFound", err)
						}
					} else if tc.Error.Code == "METHOD_NOT_ALLOWED" {
						var mna *routing.MethodNotAllowedError

						if !errors.As(err, &mna) {
							t.Fatalf("error = %v, want MethodNotAllowedError", err)
						}

						sortedGot := append([]string(nil), mna.Allowed...)
						sortedWant := append([]string(nil), tc.Error.Allowed...)

						sort.Strings(sortedGot)

						sort.Strings(sortedWant)

						if !reflect.DeepEqual(sortedGot, sortedWant) {
							t.Errorf("allowed methods = %v, want %v", sortedGot, sortedWant)
						}
					}

					return
				}

				if err != nil {
					t.Fatalf("unexpected match error: %v", err)
				}

				if matched.GetName() != tc.Expected.Name {
					t.Errorf("matched route name = %q, want %q", matched.GetName(), tc.Expected.Name)
				}

				if tc.Expected.Params != nil {
					gotParams := matched.Parameters

					if gotParams == nil {
						gotParams = map[string]string{}
					}

					if !reflect.DeepEqual(gotParams, tc.Expected.Params) {
						t.Errorf("params = %v, want %v", gotParams, tc.Expected.Params)
					}
				}

			case "generate":
				collection := routing.NewRouteCollection()
				r := routing.NewRoute("GET", tc.Route.Path, func() string { return "ok" })
				r.Name(tc.Route.Name)
				collection.Add(r)

				gen := routing.NewUrlGenerator(collection, nil, "")
				url, err := gen.Route(tc.Route.Name, tc.Params, false)

				if err != nil {
					t.Fatalf("unexpected generate error: %v", err)
				}

				if url != tc.Expected.URL {
					t.Errorf("generated url = %q, want %q", url, tc.Expected.URL)
				}
			}
		})
	}
}

func loadRoutingConformance(t *testing.T) routingFixture {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)

	if !ok {
		t.Fatal("cannot resolve conformance test path")
	}

	data, err := os.ReadFile(filepath.Join(filepath.Dir(file), "..", "..", "..", "..", "conformance", "routing.json"))

	if err != nil {
		t.Fatal(err)
	}

	var fixture routingFixture

	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}

	return fixture
}

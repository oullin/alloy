// Package generator generates fully-typed, importable TypeScript functions
// for your Go routes.
//
// It is a Go library for RouteGen output, producing identical
// TypeScript output so the same vitest test suite can validate generated files.
//
// Usage:
//
//	routes := generator.FromRouteCollection(router.GetRoutes(), generator.AdapterOptions{})
//	err := generator.Generate(routes, generator.Options{
//	    Path:     "resources/js",
//	    WithForm: true,
//	})
//
// The generator writes three output directories:
//
//   - {Path}/actions/   — per-controller TypeScript files, one function per route
//   - {Path}/routes/    — named-route helpers grouped by name prefix
//   - {Path}/routegen/ — the runtime TypeScript utility (index.ts)
package generator

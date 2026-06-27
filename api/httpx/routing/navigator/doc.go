// Package navigator generates fully-typed, importable TypeScript functions for
// Alloy routes.
//
// It exposes the backend route table to frontend applications using a
// Laravel Expose-compatible actions/, routes/, and runtime helper layout.
//
// Usage:
//
//	routes := navigator.FromRouteCollection(router.GetRoutes(), navigator.AdapterOptions{})
//	err := navigator.Generate(routes, navigator.Options{
//	    Path:     "resources/js",
//	    WithForm: true,
//	})
//
// The package writes three output directories:
//
//   - {Path}/actions/   — per-handler TypeScript files, one function per route
//   - {Path}/routes/    — named-route helpers grouped by name prefix
//   - {Path}/expose/ — the runtime TypeScript utility (index.ts)
package navigator

// Package expose generates fully-typed, importable TypeScript functions for
// Alloy routes.
//
// It exposes the backend route table to frontend applications using a
// Laravel Expose-compatible actions/, routes/, and runtime helper layout.
//
// Usage:
//
//	routes := expose.FromRouteCollection(router.GetRoutes(), expose.AdapterOptions{})
//	err := expose.Generate(routes, expose.Options{
//	    Path:     "resources/js",
//	    WithForm: true,
//	})
//
// The package writes three output directories:
//
//   - {Path}/actions/   — per-controller TypeScript files, one function per route
//   - {Path}/routes/    — named-route helpers grouped by name prefix
//   - {Path}/expose/ — the runtime TypeScript utility (index.ts)
package expose

package navigator

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Options controls what TypeScript route exposure produces.
type Options struct {
	// Path is the root output directory. Expose writes actions/,
	// routes/, and the runtime helper directory inside it.
	// Defaults to "resources/js" when empty.
	Path string

	// RuntimeDirectory is the runtime helper directory written under Path.
	// It defaults to "expose". Use "routegen" for compatibility with
	// older generated Alloy output.
	RuntimeDirectory string

	// SkipActions disables generation of the actions/ directory.
	SkipActions bool

	// SkipRoutes disables generation of the routes/ directory.
	SkipRoutes bool

	// WithForm enables form-helper generation (RouteFormDefinition exports
	// and .form property on each action function).
	WithForm bool
}

// Generate writes TypeScript route helpers for all provided routes.
//
//	{Path}/actions/   — per-controller action functions
//	{Path}/routes/    — named-route helpers
//	{Path}/{RuntimeDirectory}/ — runtime TypeScript utility (index.ts)
func Generate(routes []*RouteInfo, opts Options) error {
	if opts.Path == "" {
		opts.Path = filepath.Join("resources", "js")
	}

	if strings.TrimSpace(opts.RuntimeDirectory) == "" {
		opts.RuntimeDirectory = "expose"
	}

	g := newEmitter(opts)

	if !opts.SkipActions {
		actionsBase := filepath.Join(opts.Path, "actions")

		if err := os.RemoveAll(actionsBase); err != nil {
			return fmt.Errorf("expose: clearing actions dir: %w", err)
		}

		g.generateActions(routes, actionsBase)

		if err := g.flush(actionsBase); err != nil {
			return fmt.Errorf("expose: writing actions: %w", err)
		}
	}

	if !opts.SkipRoutes {
		routesBase := filepath.Join(opts.Path, "routes")

		if err := os.RemoveAll(routesBase); err != nil {
			return fmt.Errorf("expose: clearing routes dir: %w", err)
		}

		g.generateRoutes(routes, routesBase)

		if err := g.flush(routesBase); err != nil {
			return fmt.Errorf("expose: writing routes: %w", err)
		}
	}

	// Always copy the runtime utility.
	runtimeDir := filepath.Join(opts.Path, opts.RuntimeDirectory)

	if err := writeExposeTS(runtimeDir); err != nil {
		return fmt.Errorf("expose: writing runtime utility: %w", err)
	}

	return nil
}

func (o Options) runtimeDirectory() string {
	if strings.TrimSpace(o.RuntimeDirectory) == "" {
		return "expose"
	}

	return strings.ReplaceAll(o.RuntimeDirectory, "\\", "/")
}

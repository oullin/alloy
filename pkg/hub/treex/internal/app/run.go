package app

import (
	"context"
	"fmt"
	"time"

	"hara.sh/alloy/treex/config"
	"hara.sh/alloy/treex/internal/diskwalk"
	"hara.sh/alloy/treex/internal/gitwork"
	"hara.sh/alloy/treex/internal/inventory"
	"hara.sh/alloy/treex/internal/plan"
	"hara.sh/alloy/treex/internal/sweep"
	"hara.sh/alloy/treex/report"
)

// session is one resolved run: configuration, providers, and the collaborators
// built from them. Building it is the only place these are assembled, so scan,
// clean, and doctor cannot drift apart in how they interpret a flag.
type session struct {
	options   options
	config    config.Config
	path      string
	providers []config.Resolved
	roots     map[string]string
	scanner   inventory.Scanner
	policy    gitwork.Policy
	warnings  []string
}

// elapsedPrecision keeps a reported duration readable; a scan measured to the
// nanosecond tells nobody anything.
const elapsedPrecision = 10 * time.Millisecond

func (d *deps) session(opts options, categories []config.Category) (*session, error) {
	loader := config.Loader{Home: d.home, Env: d.env, Explicit: opts.configPath}

	cfg, path, err := loader.Load()

	if err != nil {
		return nil, err
	}

	cfg = opts.applyTo(cfg)

	selected, err := cfg.Select(opts.providers)

	if err != nil {
		return nil, err
	}

	mount := ""

	if opts.root != "" {
		mount = config.Expand(opts.root, d.home)
	}

	resolved := config.ResolveAll(selected, d.home, mount)
	roots := make(map[string]string, len(resolved))

	for _, provider := range resolved {
		roots[provider.Name] = provider.Root
	}

	policy := gitwork.Policy{
		RequireCleanWorktree: cfg.Safety.RequireCleanWorktree,
		RequirePushedCommits: cfg.Safety.RequirePushedCommits,
		RequireNoStash:       cfg.Safety.RequireNoStash,
		ProtectedBranches:    cfg.Safety.ProtectedBranches,
		ProtectPaths:         cfg.Safety.ProtectedPaths(d.home, mount),
		Force:                opts.force,
	}

	runner := gitwork.Runner{Timeout: cfg.Safety.StatusTimeout.Std()}

	session := &session{
		options:   opts,
		config:    cfg,
		path:      path,
		providers: resolved,
		roots:     roots,
		policy:    policy,
	}

	session.scanner = inventory.Scanner{
		Config:     cfg,
		Providers:  resolved,
		Categories: categories,
		Pool:       diskwalk.NewPool(cfg.Defaults.Jobs),
		Runner:     runner,
		Policy:     policy,
		Inspector: gitwork.Inspector{
			Runner:       runner,
			ArtifactRoot: cfg.Artifacts.ArtifactRoot,
		},
		Errors: func(path string, err error) {
			session.warn(fmt.Sprintf("%s: %v", path, err))
		},
	}

	return session, nil
}

// warn records a non-fatal problem. They are collected rather than printed as
// they happen so they do not interleave with the report on a busy scan.
func (s *session) warn(message string) {
	// A scan of a home directory can hit hundreds of unreadable paths; showing
	// every one would bury the report it is attached to.
	const maxWarnings = 10

	if len(s.warnings) >= maxWarnings {
		return
	}

	s.warnings = append(s.warnings, message)
}

func (s *session) guard() sweep.Guard {
	roots := make([]string, 0, len(s.providers))

	for _, provider := range s.providers {
		roots = append(roots, provider.Jail)
	}

	return sweep.Guard{Roots: roots, Protected: s.policy.ProtectPaths}
}

func (s *session) projection(version string, mode report.Mode) projection {
	return projection{version: version, config: s.path, mode: mode, roots: s.roots}
}

// render writes a report and returns the process exit code it implies.
func (d *deps) render(session *session, out report.Report, format report.Format) int {
	out.Warnings = session.warnings

	renderer := report.Renderer{
		Verbose: session.options.verbose,
		Explain: session.options.explain,
		Limit:   session.options.limit,
	}

	if err := renderer.Render(d.stdout, format, out); err != nil {
		return d.reportError(err)
	}

	return out.ExitCode()
}

// scanFor runs the enumerate-inspect-measure pipeline and builds a plan.
func (s *session) scanFor(ctx context.Context) (inventory.Inventory, plan.Plan, error) {
	found, err := s.scanner.Scan(ctx)

	if err != nil {
		return inventory.Inventory{}, plan.Plan{}, err
	}

	return found, plan.Build(found, s.options.limit), nil
}

// pruner builds a sweeper limited to registry repair, which touches no user
// data and so needs none of the path guarding a removal does.
func (s *session) pruner() sweep.Sweeper {
	return sweep.Sweeper{Guard: s.guard(), Runner: s.scanner.Runner}
}

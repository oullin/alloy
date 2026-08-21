package config

import "time"

// Default returns the built-in configuration: every agent tool treex knows
// about, with the roots and subdirectories observed in the wild. A user's file
// is layered on top of this rather than replacing it, so disabling one
// provider or adding one cache directory is a three-line file.
//
// Providers whose debris is large and unambiguous ship enabled. Ones that
// currently hold little (copilot, factory) or whose contents are ambiguous
// ship disabled, as do all toolchain caches: ~/.gradle is regenerable, but it
// is not agent debris and sweeping it should be a deliberate opt-in.
func Default() Config {
	return Config{
		Version: Version,
		Defaults: Defaults{
			Jobs:         0,
			MaxDepth:     4,
			OlderThan:    Duration(7 * 24 * time.Hour),
			MinSize:      0,
			ApparentSize: false,
		},
		Safety: Safety{
			RequireCleanWorktree: true,
			RequirePushedCommits: true,
			RequireNoStash:       true,
			StatusTimeout:        Duration(30 * time.Second),
			ProtectedBranches:    []string{"main", "master", "develop"},
			ProtectPaths:         []string{"~/Sites", "~/Documents", "~/Desktop", "~/Downloads"},
		},
		Artifacts: Artifacts{
			MaxDepth: 6,
			Names: []string{
				"node_modules", "dist", "build", ".turbo", ".next", ".nuxt",
				".svelte-kit", "target", "vendor", "Pods", ".expo", ".venv",
				"venv", "__pycache__", ".pytest_cache", ".gradle", ".dart_tool",
				"DerivedData", ".parcel-cache", ".cache", ".mypy_cache",
			},
		},
		Providers: defaultProviders(),
	}
}

func defaultProviders() []Provider {
	return []Provider{
		{
			Name:    "claude",
			Kind:    KindAgent,
			Enabled: true,
			Root:    "~/.claude",
			// worktrees holds linked git worktrees both directly and one level
			// down, grouped by originating repository.
			Worktrees: []TreeSource{
				{Path: "worktrees", Depth: 3},
				{Path: "work", Depth: 2},
			},
			Caches: []string{
				"browser-artifacts", "image-cache", "paste-cache", "cache",
				"shell-snapshots", "session-env",
			},
			Sessions: []string{
				"projects", "telemetry", "file-history", "history.jsonl", "backups", "debug",
			},
		},
		{
			Name:    "codex",
			Kind:    KindAgent,
			Enabled: true,
			Root:    "~/.codex",
			// Depth 3 is load-bearing: worktrees/codex is itself a directory of
			// worktrees, so the interesting entries sit two levels down.
			Worktrees: []TreeSource{
				{Path: "worktrees", Depth: 3},
				{Path: "work", Depth: 2},
			},
			Caches: []string{
				"cache", "runtimes", "android", "maestro", "cocoapods-gems",
				"node_repl", "computer-use", "plugins", "tmp", ".tmp",
			},
			Sessions: []string{
				"archived_sessions", "sessions", "sqlite", "logs_2.sqlite",
				"log", "history.jsonl", "ambient-suggestions", "visualizations",
			},
			Orphans: []OrphanRule{
				{
					Name:     "stale-global-state-temps",
					Match:    `^\.\.codex-global-state\.json(\.bak)?\.tmp-\d+-[0-9a-fA-F-]+$`,
					Liveness: LivenessNone,
				},
			},
		},
		{
			Name:    "cursor",
			Kind:    KindAgent,
			Enabled: true,
			Root:    "~/.cursor",
			Caches:  []string{"extensions", "plugins"},
			Sessions: []string{
				"chats", "acp-sessions", "ai-tracking", "projects",
			},
		},
		{
			Name:    "gemini",
			Kind:    KindAgent,
			Enabled: true,
			Root:    "~/.gemini",
			Worktrees: []TreeSource{
				{Path: "worktrees", Depth: 3},
				{Path: "work", Depth: 2},
			},
			Caches:   []string{"cache", "tmp", "native", "antigravity", "antigravity-cli"},
			Sessions: []string{"history"},
		},
		{
			Name:     "grok",
			Kind:     KindAgent,
			Enabled:  true,
			Root:     "~/.grok",
			Caches:   []string{"bundled", "downloads", "memtrace"},
			Sessions: []string{"sessions"},
		},
		{
			Name:    "agent-browser",
			Kind:    KindAgent,
			Enabled: true,
			Root:    "~/.agent-browser",
			Caches:  []string{"browsers"},
			// Each headless browser session leaves a set of files named after
			// its pid. Once that process is gone the whole set is dead weight.
			Orphans: []OrphanRule{
				{
					Name:     "dead-browser-sessions",
					Match:    `^(?P<session>.+)-(?P<pid>\d+)\.(config|engine|pid|sock|stream|version)$`,
					Liveness: LivenessPID,
					Group:    "pid",
				},
			},
		},
		{Name: "copilot", Kind: KindAgent, Enabled: false, Root: "~/.copilot"},
		{Name: "factory", Kind: KindAgent, Enabled: false, Root: "~/.factory"},
		{Name: "pi", Kind: KindAgent, Enabled: false, Root: "~/.pi"},

		{
			Name:    "gradle",
			Kind:    KindToolchain,
			Enabled: false,
			Root:    "~/.gradle",
			Caches:  []string{"caches", "daemon", "native", "wrapper"},
		},
		{Name: "expo", Kind: KindToolchain, Enabled: false, Root: "~/.expo", Caches: []string{"."}},
		{Name: "vite-plus", Kind: KindToolchain, Enabled: false, Root: "~/.vite-plus", Caches: []string{"."}},
	}
}

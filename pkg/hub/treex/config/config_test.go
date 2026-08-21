package config_test

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"hara.sh/alloy/treex/config"
)

func TestParseSizeReadsUnitSuffixes(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want int64
	}{
		{"1024", 1024},
		{"1KB", 1000},
		{"1KiB", 1024},
		{"500MB", 500 * 1000 * 1000},
		{"2G", 2 << 30},
		{" 1.5 GiB ", 1610612736},
	}

	for _, tc := range cases {
		got, err := config.ParseSize(tc.in)

		if err != nil {
			t.Fatalf("ParseSize(%q): %v", tc.in, err)
		}

		if got.Bytes() != tc.want {
			t.Fatalf("ParseSize(%q) = %d, want %d", tc.in, got.Bytes(), tc.want)
		}
	}
}

func TestParseSizeRejectsNonsense(t *testing.T) {
	t.Parallel()

	for _, in := range []string{"", "big", "-5MB", "MB"} {
		if _, err := config.ParseSize(in); err == nil {
			t.Fatalf("ParseSize(%q) err = nil, want a failure", in)
		}
	}
}

// Days and weeks are the only units anyone reaches for when talking about
// stale worktrees, and time.ParseDuration accepts neither.
func TestParseDurationSupportsDaysAndWeeks(t *testing.T) {
	t.Parallel()

	cases := []struct {
		in   string
		want time.Duration
	}{
		{"90m", 90 * time.Minute},
		{"7d", 7 * 24 * time.Hour},
		{"2w", 14 * 24 * time.Hour},
	}

	for _, tc := range cases {
		got, err := config.ParseDuration(tc.in)

		if err != nil {
			t.Fatalf("ParseDuration(%q): %v", tc.in, err)
		}

		if got.Std() != tc.want {
			t.Fatalf("ParseDuration(%q) = %v, want %v", tc.in, got.Std(), tc.want)
		}
	}
}

func TestDefaultShipsAKnownAgentCatalog(t *testing.T) {
	t.Parallel()

	cfg := config.Default()

	if err := config.Validate(&cfg); err != nil {
		t.Fatalf("the built-in configuration must be valid: %v", err)
	}

	for _, name := range []string{"claude", "codex", "cursor", "gemini", "agent-browser"} {
		if _, err := cfg.Select([]string{name}); err != nil {
			t.Fatalf("Select(%q): %v", name, err)
		}
	}
}

func TestDefaultLeavesToolchainCachesOptIn(t *testing.T) {
	t.Parallel()

	// A gradle cache is regenerable, but it is not agent debris and sweeping it
	// should be a deliberate choice.
	for _, provider := range config.Default().Providers {
		if provider.Kind == config.KindToolchain && provider.Enabled {
			t.Fatalf("toolchain provider %q ships enabled, want opt-in", provider.Name)
		}
	}
}

func TestLoadFallsBackToTheBuiltInsWhenNoFileExists(t *testing.T) {
	t.Parallel()

	loader := config.Loader{Home: t.TempDir(), Env: emptyEnv}

	cfg, path, err := loader.Load()

	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	if path != "" {
		t.Fatalf("path = %q, want empty", path)
	}

	if len(cfg.Providers) == 0 {
		t.Fatal("providers are empty, want the built-in catalog")
	}
}

func TestLoadRejectsAMissingExplicitFile(t *testing.T) {
	t.Parallel()

	loader := config.Loader{
		Home:     t.TempDir(),
		Env:      emptyEnv,
		Explicit: filepath.Join(t.TempDir(), "absent.yml"),
	}

	// A default location that does not exist is normal; an explicit --config
	// that does not exist is a mistake worth surfacing.
	if _, _, err := loader.Load(); err == nil {
		t.Fatal("Load err = nil, want a missing-file failure")
	}
}

func TestLoadPatchesABuiltInProviderRatherThanReplacingIt(t *testing.T) {
	t.Parallel()

	cfg := load(t, `
version: 1
providers:
    - name: codex
      caches: [cache, my-extra-cache]
`)

	codex := find(t, cfg, "codex")

	if len(codex.Caches) != 2 {
		t.Fatalf("Caches = %v, want the file's list", codex.Caches)
	}

	// The fields the file said nothing about must survive.
	if len(codex.Worktrees) == 0 {
		t.Fatal("Worktrees were dropped, want the built-in value preserved")
	}

	if len(codex.Sessions) == 0 {
		t.Fatal("Sessions were dropped, want the built-in value preserved")
	}
}

func TestLoadHonoursAnExplicitFalse(t *testing.T) {
	t.Parallel()

	cfg := load(t, `
version: 1
providers:
    - name: cursor
      enabled: false
`)

	if find(t, cfg, "cursor").Enabled {
		t.Fatal("cursor is enabled, want the explicit false honoured")
	}
}

// A boolean that defaults to true and an absent key look identical after
// unmarshalling. Getting this wrong would silently re-enable a safety gate the
// user turned off — or worse, leave one off that they never touched.
func TestLoadHonoursAnExplicitlyDisabledSafetyGate(t *testing.T) {
	t.Parallel()

	cfg := load(t, `
version: 1
safety:
    require-pushed-commits: false
`)

	if cfg.Safety.RequirePushedCommits {
		t.Fatal("RequirePushedCommits = true, want the explicit false honoured")
	}

	// The gates the file said nothing about keep their protective defaults.
	if !cfg.Safety.RequireCleanWorktree {
		t.Fatal("RequireCleanWorktree = false, want the default preserved")
	}
}

func TestLoadMergesProtectedPathsRatherThanReplacingThem(t *testing.T) {
	t.Parallel()

	cfg := load(t, `
version: 1
safety:
    protect-paths:
        - ~/Precious
`)

	found := false
	guarded := false

	for _, path := range cfg.Safety.ProtectPaths {
		if path == "~/Precious" {
			found = true
		}

		if path == "~/Sites" {
			guarded = true
		}
	}

	if !found {
		t.Fatalf("ProtectPaths = %v, want the addition", cfg.Safety.ProtectPaths)
	}

	// Naming one extra directory must not quietly drop the built-in guards.
	if !guarded {
		t.Fatalf("ProtectPaths = %v, want the built-in guards preserved", cfg.Safety.ProtectPaths)
	}
}

func TestLoadAcceptsANewProvider(t *testing.T) {
	t.Parallel()

	cfg := load(t, `
version: 1
providers:
    - name: my-agent
      root: ~/.my-agent
      worktrees:
          - { path: worktrees, depth: 2 }
`)

	provider := find(t, cfg, "my-agent")

	if !provider.Enabled {
		t.Fatal("a provider written out in full should be enabled")
	}

	if provider.Kind != config.KindAgent {
		t.Fatalf("Kind = %q, want %q", provider.Kind, config.KindAgent)
	}
}

func TestLoadRejectsAnOrphanRuleThatCannotFindAPID(t *testing.T) {
	t.Parallel()

	_, err := loadRaw(t, `
version: 1
providers:
    - name: broken
      root: ~/.broken
      orphans:
          - name: no-pid
            match: '^session-.*$'
            liveness: pid
            group: pid
`)

	// A pid rule that captures no pid would treat every match as dead, and for
	// a browser session directory that means deleting a running browser's
	// socket out from under it.
	if err == nil {
		t.Fatal("Load err = nil, want the unusable rule rejected")
	}
}

func TestLoadRejectsAnUnknownSchemaVersion(t *testing.T) {
	t.Parallel()

	if _, err := loadRaw(t, "version: 99\n"); err == nil {
		t.Fatal("Load err = nil, want the unknown version rejected")
	}
}

func TestSelectRejectsAnUnknownProviderName(t *testing.T) {
	t.Parallel()

	// A typo'd --providers must not look like "nothing to clean".
	if _, err := config.Default().Select([]string{"clod"}); err == nil {
		t.Fatal("Select err = nil, want an unknown name rejected")
	}
}

func TestSelectEnablesAProviderNamedExplicitly(t *testing.T) {
	t.Parallel()

	selected, err := config.Default().Select([]string{"gradle"})

	if err != nil {
		t.Fatalf("Select: %v", err)
	}

	if len(selected) != 1 || !selected[0].Enabled {
		t.Fatal("naming a disabled provider should switch it on for the run")
	}
}

// Session data is the one class treex cannot regenerate, so it is only swept
// when asked for by name.
func TestParseCategoriesExcludesSessionsByDefault(t *testing.T) {
	t.Parallel()

	categories, err := config.ParseCategories(nil)

	if err != nil {
		t.Fatalf("ParseCategories: %v", err)
	}

	for _, category := range categories {
		if category == config.CategorySession {
			t.Fatal("sessions are in the default selection, want them opt-in")
		}
	}
}

func TestParseCategoriesRejectsAnUnknownName(t *testing.T) {
	t.Parallel()

	if _, err := config.ParseCategories([]string{"everything"}); err == nil {
		t.Fatal("ParseCategories err = nil, want an unknown category rejected")
	}
}

func TestArtifactRootMatchesOnlyTheFirstSegment(t *testing.T) {
	t.Parallel()

	artifacts := config.Default().Artifacts

	if !artifacts.ArtifactRoot("node_modules/left-pad/index.js") {
		t.Fatal("ArtifactRoot = false for a path under node_modules, want true")
	}

	// A source file that merely mentions the name is real work.
	if artifacts.ArtifactRoot("src/node_modules.md") {
		t.Fatal("ArtifactRoot = true for a source file, want false")
	}
}

func TestResolveExpandsHomeAndMount(t *testing.T) {
	t.Parallel()

	provider := config.Provider{Name: "claude", Root: "~/.claude"}

	if got, want := provider.Resolve("/home/u", "").Root, filepath.Clean("/home/u/.claude"); got != want {
		t.Fatalf("Root = %q, want %q", got, want)
	}

	// The container sees the host home under a mount point, but the provider
	// keeps its familiar name in the report.
	if got, want := provider.Resolve("/home/u", "/host").Root, filepath.Clean("/host/home/u/.claude"); got != want {
		t.Fatalf("mounted Root = %q, want %q", got, want)
	}
}

func load(t *testing.T, contents string) config.Config {
	t.Helper()

	cfg, err := loadRaw(t, contents)

	if err != nil {
		t.Fatalf("Load: %v", err)
	}

	return cfg
}

func loadRaw(t *testing.T, contents string) (config.Config, error) {
	t.Helper()

	home := t.TempDir()
	path := filepath.Join(home, ".treex.yml")

	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write config: %v", err)
	}

	cfg, _, err := config.Loader{Home: home, Env: emptyEnv, Explicit: path}.Load()

	return cfg, err
}

func find(t *testing.T, cfg config.Config, name string) config.Provider {
	t.Helper()

	for _, provider := range cfg.Providers {
		if provider.Name == name {
			return provider
		}
	}

	t.Fatalf("provider %q not found", name)

	return config.Provider{}
}

func emptyEnv(string) string {
	return ""
}

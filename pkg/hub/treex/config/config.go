package config

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

// Kind separates the agent tools treex was written for from general toolchain
// caches, which are opt-in: ~/.gradle is regenerable but it is not agent
// debris, and sweeping it should be a deliberate choice.
type Kind string

// Category is one class of reclaimable thing. Every candidate carries exactly
// one, and --categories filters on it.
type Category string

// Config is the whole configuration document.
type Config struct {
	Version   int        `mapstructure:"version"`
	Defaults  Defaults   `mapstructure:"defaults"`
	Safety    Safety     `mapstructure:"safety"`
	Artifacts Artifacts  `mapstructure:"artifacts"`
	Providers []Provider `mapstructure:"providers"`
}

// Defaults are the knobs that apply to every command unless a flag overrides
// them.
type Defaults struct {
	// Jobs is the walker's worker count. Zero means derive it from NumCPU;
	// see diskwalk.NewPool, which oversubscribes because the walk is bound by
	// IO latency rather than CPU.
	Jobs int `mapstructure:"jobs"`

	// MaxDepth bounds how deep worktree discovery descends before giving up.
	MaxDepth int `mapstructure:"max-depth"`

	// OlderThan compares against the newest mtime anywhere in a candidate: a
	// worktree touched an hour ago is live work whatever its git state says.
	OlderThan Duration `mapstructure:"older-than"`

	// MinSize skips candidates too small to be worth reporting.
	MinSize Size `mapstructure:"min-size"`

	// ApparentSize switches byte accounting from allocated blocks to file
	// sizes. The default (false) matches du and reflects what actually frees.
	ApparentSize bool `mapstructure:"apparent-size"`

	// FollowSymlinks is off by default: following them would double-count and,
	// worse, could walk a link out of a provider root.
	FollowSymlinks bool `mapstructure:"follow-symlinks"`
}

// Safety is the set of gates a candidate must pass before treex will delete
// it. Every field defaults to the conservative choice.
type Safety struct {
	RequireCleanWorktree bool     `mapstructure:"require-clean-worktree"`
	RequirePushedCommits bool     `mapstructure:"require-pushed-commits"`
	RequireNoStash       bool     `mapstructure:"require-no-stash"`
	StatusTimeout        Duration `mapstructure:"status-timeout"`
	ProtectedBranches    []string `mapstructure:"protected-branches"`
	ProtectPaths         []string `mapstructure:"protect-paths"`
}

// Artifacts names the regenerable build directories treex recognises inside a
// worktree. The list does double duty: it is what CategoryArtifact sweeps, and
// it is what the dirty-worktree check ignores when deciding whether a worktree
// holds real uncommitted work.
type Artifacts struct {
	Names    []string `mapstructure:"names"`
	MaxDepth int      `mapstructure:"max-depth"`
}

// Provider is one agent tool or toolchain treex knows how to sweep.
type Provider struct {
	Name    string `mapstructure:"name"`
	Kind    Kind   `mapstructure:"kind"`
	Enabled bool   `mapstructure:"enabled"`
	Root    string `mapstructure:"root"`

	// Override replaces a built-in provider outright instead of patching it
	// field by field. Without it, a file that lists only a root still inherits
	// the built-in's cache and session directories.
	Override bool `mapstructure:"override"`

	Worktrees []TreeSource `mapstructure:"worktrees"`
	Caches    []string     `mapstructure:"caches"`
	Sessions  []string     `mapstructure:"sessions"`
	Orphans   []OrphanRule `mapstructure:"orphans"`
}

// TreeSource is a directory under a provider root that holds worktrees, and
// how deep to look. Depth is not uniform across agents: ~/.claude/worktrees
// holds worktrees directly and one level down, while ~/.codex/worktrees nests
// a whole second tier under worktrees/codex.
type TreeSource struct {
	Path  string `mapstructure:"path"`
	Depth int    `mapstructure:"depth"`
}

// OrphanRule matches leftover files from processes that have exited. Match is
// a regular expression against the base name; when Liveness is "pid" the named
// Group capture is read as a process id and the file is only a candidate when
// that process is gone.
type OrphanRule struct {
	Name     string `mapstructure:"name"`
	Match    string `mapstructure:"match"`
	Liveness string `mapstructure:"liveness"`
	Group    string `mapstructure:"group"`

	pattern *regexp.Regexp
}

// Version is the only schema version treex accepts. It exists so a future
// incompatible change can be rejected with a clear message rather than
// silently mis-parsed into a plan that deletes the wrong thing.
const Version = 1

const (
	// KindAgent is an AI coding agent's state directory.
	KindAgent Kind = "agent"

	// KindToolchain is a language or build toolchain cache.
	KindToolchain Kind = "toolchain"
)

const (
	// CategoryWorktree is a whole agent worktree or clone.
	CategoryWorktree Category = "worktree"

	// CategoryArtifact is a regenerable build directory inside a worktree that
	// is being kept. This is the safe daily driver: it never touches git state.
	CategoryArtifact Category = "artifact"

	// CategoryCache is an agent's scratch or cache directory.
	CategoryCache Category = "cache"

	// CategorySession is conversation history, transcripts, and logs. Removing
	// these loses a record that cannot be regenerated, unlike everything else
	// treex sweeps, so it is never included unless asked for by name.
	CategorySession Category = "session"

	// CategoryOrphan is a leftover file from a process that is no longer
	// running, such as a dead browser session's socket.
	CategoryOrphan Category = "orphan"
)

// Categories lists every category in report order.
var Categories = []Category{
	CategoryWorktree,
	CategoryArtifact,
	CategoryCache,
	CategorySession,
	CategoryOrphan,
}

// Pattern returns the compiled expression. Validate compiles every rule at load
// time, so a rule reaching the scanner is always ready to use.
func (o OrphanRule) Pattern() *regexp.Regexp {
	return o.pattern
}

// PID extracts the process id a name encodes, reporting false when the rule
// does not use pid liveness or the name does not match.
func (o OrphanRule) PID(name string) (int, bool) {
	if o.Liveness != LivenessPID || o.pattern == nil {
		return 0, false
	}

	match := o.pattern.FindStringSubmatch(name)

	if match == nil {
		return 0, false
	}

	index := o.pattern.SubexpIndex(o.Group)

	if index < 0 || index >= len(match) {
		return 0, false
	}

	pid := 0

	if _, err := fmt.Sscanf(match[index], "%d", &pid); err != nil {
		return 0, false
	}

	return pid, true
}

const (
	// LivenessPID makes a match conditional on the captured process being gone.
	LivenessPID = "pid"

	// LivenessNone treats every match as an orphan.
	LivenessNone = "none"
)

// IsArtifact reports whether a directory name is a regenerable build artifact.
func (a Artifacts) IsArtifact(name string) bool {
	return slices.Contains(a.Names, name)
}

// ArtifactRoot reports whether the first segment of a repository-relative path
// is an artifact directory. The dirty-worktree check leans on this: without it
// an untracked node_modules makes every worktree look like it holds real work,
// and treex reclaims nothing.
func (a Artifacts) ArtifactRoot(relative string) bool {
	cleaned := strings.TrimPrefix(strings.ReplaceAll(relative, "\\", "/"), "./")
	segment, _, _ := strings.Cut(cleaned, "/")

	return segment != "" && a.IsArtifact(segment)
}

// Enabled returns the providers that are switched on, in catalog order.
func (c Config) Enabled() []Provider {
	out := make([]Provider, 0, len(c.Providers))

	for _, provider := range c.Providers {
		if provider.Enabled {
			out = append(out, provider)
		}
	}

	return out
}

// Select narrows the enabled providers to the named ones. An empty selection
// means all of them; an unknown name is an error rather than a silent no-op,
// because a typo'd --providers would otherwise look like "nothing to clean".
func (c Config) Select(names []string) ([]Provider, error) {
	enabled := c.Enabled()

	if len(names) == 0 {
		return enabled, nil
	}

	known := make(map[string]Provider, len(c.Providers))

	for _, provider := range c.Providers {
		known[provider.Name] = provider
	}

	out := make([]Provider, 0, len(names))

	for _, name := range names {
		provider, ok := known[name]

		if !ok {
			return nil, fmt.Errorf("%w: no provider named %q", ErrInvalidProvider, name)
		}

		// An explicit --providers is an instruction, so it overrides the
		// enabled flag: naming a disabled provider switches it on for this run.
		provider.Enabled = true

		out = append(out, provider)
	}

	return out, nil
}

// ParseCategories reads a --categories value, defaulting to everything except
// sessions. Session data is the one class treex cannot regenerate, so it is
// only ever swept when named explicitly.
func ParseCategories(raw []string) ([]Category, error) {
	if len(raw) == 0 {
		return []Category{CategoryWorktree, CategoryArtifact, CategoryCache, CategoryOrphan}, nil
	}

	out := make([]Category, 0, len(raw))

	for _, item := range raw {
		category := Category(strings.ToLower(strings.TrimSpace(item)))

		if !slices.Contains(Categories, category) {
			return nil, fmt.Errorf("%w: %q", ErrUnknownCategory, item)
		}

		if !slices.Contains(out, category) {
			out = append(out, category)
		}
	}

	return out, nil
}

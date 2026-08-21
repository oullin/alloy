package gitwork

import (
	"context"
	"path/filepath"
	"strings"
)

// Entry is one record from a repository's worktree registry.
type Entry struct {
	Path     string
	Head     string
	Branch   string
	Bare     bool
	Detached bool
	Locked   bool

	// LockReason is git's explanation, empty when a worktree is locked without
	// one.
	LockReason string

	// Prunable means git considers the registry entry stale, almost always
	// because the working tree it names is gone. These are the entries a
	// previous rm -rf left behind, and repairing them is what doctor does.
	Prunable       bool
	PrunableReason string
}

// Registry is a repository's worktree list, keyed by cleaned path.
type Registry struct {
	Repo    string
	Entries map[string]Entry
}

// LoadRegistry reads the worktree list for one repository.
//
// It is deliberately per-repository rather than per-worktree: a scan of a few
// hundred worktrees typically spans only a few dozen repositories, so asking
// once per repository turns hundreds of git invocations into dozens.
func LoadRegistry(ctx context.Context, runner Runner, repo string) (Registry, error) {
	out, err := runner.Run(ctx, repo, "worktree", "list", "--porcelain")

	if err != nil {
		return Registry{}, err
	}

	return Registry{Repo: repo, Entries: parseRegistry(out)}, nil
}

// Lookup finds the entry for a working tree path.
func (r Registry) Lookup(path string) (Entry, bool) {
	entry, ok := r.Entries[filepath.Clean(path)]

	return entry, ok
}

// Prunable returns the registry entries git considers stale, which are exactly
// the ones a manual delete left behind.
func (r Registry) Prunable() []Entry {
	out := make([]Entry, 0)

	for _, entry := range r.Entries {
		if entry.Prunable {
			out = append(out, entry)
		}
	}

	return out
}

// parseRegistry reads git worktree list --porcelain: blank-line separated
// records of "key value" lines, where valueless keys are flags.
func parseRegistry(out string) map[string]Entry {
	entries := make(map[string]Entry)

	var current Entry

	flush := func() {
		if current.Path != "" {
			entries[filepath.Clean(current.Path)] = current
		}

		current = Entry{}
	}

	for line := range strings.SplitSeq(strings.ReplaceAll(out, "\x00", "\n"), "\n") {
		trimmed := strings.TrimRight(line, "\r")

		if trimmed == "" {
			flush()

			continue
		}

		key, value, _ := strings.Cut(trimmed, " ")

		switch key {
		case "worktree":
			flush()
			current.Path = value
		case "HEAD":
			current.Head = value
		case "branch":
			current.Branch = strings.TrimPrefix(value, "refs/heads/")
		case "bare":
			current.Bare = true
		case "detached":
			current.Detached = true
		case "locked":
			current.Locked = true
			current.LockReason = value
		case "prunable":
			current.Prunable = true
			current.PrunableReason = value
		}
	}

	flush()

	return entries
}

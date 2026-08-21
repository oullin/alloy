package gitwork

import (
	"context"
	"os"
	"path/filepath"
	"strings"
)

// Worktree is everything treex knows about one candidate directory.
type Worktree struct {
	Path string
	Kind Kind

	// CommonDir is the shared git directory; MainRepo is the repository whose
	// registry owns this worktree.
	CommonDir string
	MainRepo  string

	Branch   string
	Head     string
	Detached bool

	Locked     bool
	LockReason string

	Prunable       bool
	PrunableReason string

	// Dirty counts tracked modifications; Untracked counts untracked paths that
	// are not recognised build artifacts.
	Dirty     int
	Untracked int

	// Unpushed counts commits reachable from HEAD but from no remote-tracking
	// ref. This is the figure that most often blocks a removal.
	Unpushed int

	// Stashes is the repository's stash count. It is shared across every linked
	// worktree of a repository, because refs/stash lives in the common dir.
	Stashes int

	// Alternates means the object store borrows from another repository, so
	// removing this tree frees less than it measures.
	Alternates bool

	// Broken carries the reason a classification failed.
	Broken string

	// Err records an inspection that could not complete. Any candidate with a
	// set Err is treated as unsafe: an unanswered question is never a yes.
	Err error
}

// Inspector answers the questions that decide whether a worktree can go.
type Inspector struct {
	Runner Runner

	// ArtifactRoot reports whether an untracked path is a regenerable build
	// directory. Supplying it is what makes the tool useful: without it, the
	// untracked node_modules in every worktree makes every worktree look like
	// it holds uncommitted work, and nothing is ever reclaimable.
	ArtifactRoot func(relative string) bool
}

// Inspect gathers the git facts about one candidate directory.
func (i Inspector) Inspect(ctx context.Context, path string) Worktree {
	path = filepath.Clean(path)
	classification := Classify(path)

	tree := Worktree{Path: path, Kind: classification.Kind, Broken: classification.Reason}

	switch classification.Kind {
	case KindPlain:
		return tree
	case KindBroken:
		// Nothing further can be asked of a repository that cannot be opened,
		// and nothing needs to be: a broken pointer means the registry entry is
		// already gone, so removal cannot make anything worse.
		return tree
	}

	common, err := i.Runner.Line(ctx, path, "rev-parse", "--path-format=absolute", "--git-common-dir")

	if err != nil {
		tree.Err = err

		return tree
	}

	tree.CommonDir = filepath.Clean(common)
	tree.MainRepo = filepath.Dir(tree.CommonDir)
	tree.Alternates = fileExists(filepath.Join(tree.CommonDir, "objects", "info", "alternates"))

	i.readHead(ctx, path, &tree)
	i.readStatus(ctx, path, &tree)

	tree.Unpushed = i.countUnpushed(ctx, path, &tree)
	tree.Stashes = i.countStashes(ctx, path)

	return tree
}

// Apply folds a registry entry's flags into an inspected worktree. Lock and
// prunable state live only in the parent repository's registry, so they arrive
// separately from everything read inside the worktree itself.
func (w *Worktree) Apply(entry Entry) {
	w.Locked = entry.Locked
	w.LockReason = entry.LockReason
	w.Prunable = entry.Prunable
	w.PrunableReason = entry.PrunableReason

	if w.Branch == "" {
		w.Branch = entry.Branch
	}

	if w.Head == "" {
		w.Head = entry.Head
	}
}

// IsMainWorktree reports whether this path is the primary working tree of a
// repository that has linked worktrees hanging off it. Removing such a
// directory would orphan every one of them, so it is never allowed.
//
// The check applies only to linked worktrees. A clone is by definition its own
// main working tree, so asking the question of one would answer yes every time
// and make every clone permanently unremovable — which would rule out the
// largest category of agent debris there is. What protects a clone is that it
// must lie inside a provider root and outside every protected path, both of
// which are enforced independently.
func (w Worktree) IsMainWorktree() bool {
	if w.Kind != KindLinked {
		return false
	}

	return w.MainRepo != "" && filepath.Clean(w.Path) == filepath.Clean(w.MainRepo)
}

// HasWork reports whether anything here would be lost by deleting the tree.
func (w Worktree) HasWork() bool {
	return w.Dirty > 0 || w.Untracked > 0 || w.Unpushed > 0
}

func (i Inspector) readHead(ctx context.Context, path string, tree *Worktree) {
	if head, err := i.Runner.Line(ctx, path, "rev-parse", "HEAD"); err == nil {
		tree.Head = head
	}

	branch, err := i.Runner.Line(ctx, path, "rev-parse", "--abbrev-ref", "HEAD")

	if err != nil {
		return
	}

	if branch == "HEAD" {
		tree.Detached = true

		return
	}

	tree.Branch = branch
}

// readStatus counts uncommitted work, ignoring regenerable build directories.
func (i Inspector) readStatus(ctx context.Context, path string, tree *Worktree) {
	out, err := i.Runner.Run(ctx, path,
		"--no-optional-locks", "status", "--porcelain=v2", "-z",
		"--untracked-files=normal", "--ignore-submodules=dirty",
	)

	if err != nil {
		tree.Err = err

		return
	}

	for record := range strings.SplitSeq(out, "\x00") {
		if record == "" {
			continue
		}

		switch record[0] {
		case '?':
			relative := strings.TrimSpace(record[1:])

			// --untracked-files=normal collapses an untracked directory into a
			// single record, so a 1.7GB node_modules costs one comparison here
			// rather than two hundred thousand.
			if i.ArtifactRoot != nil && i.ArtifactRoot(relative) {
				continue
			}

			tree.Untracked++
		case '1', '2', 'u':
			tree.Dirty++
		}
	}
}

// countUnpushed counts commits that exist nowhere but here.
//
// rev-list HEAD --not --remotes answers exactly the question that matters: are
// there commits reachable from HEAD that no remote-tracking ref knows about? It
// works on a detached HEAD, and when remote refs are stale it over-reports,
// which is the direction to err in.
func (i Inspector) countUnpushed(ctx context.Context, path string, tree *Worktree) int {
	if _, err := i.Runner.Line(ctx, path, "rev-parse", "--verify", "HEAD"); err != nil {
		// An unborn HEAD has no commits to lose.
		return 0
	}

	count, err := i.Runner.Count(ctx, path, "rev-list", "--count", "HEAD", "--not", "--remotes")

	if err != nil {
		tree.Err = err

		return 0
	}

	return count
}

func (i Inspector) countStashes(ctx context.Context, path string) int {
	count, err := i.Runner.Count(ctx, path, "rev-list", "--walk-reflogs", "--count", "refs/stash")

	if err != nil {
		// No stash ref at all is the overwhelmingly common case and git exits
		// non-zero for it, so this is not worth surfacing as an error.
		return 0
	}

	return count
}

func fileExists(path string) bool {
	_, err := os.Stat(path)

	return err == nil
}

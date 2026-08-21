package diskwalk

import (
	"io/fs"
	"sync"
	"time"
)

// Decision is what a Pruner says about a directory the walk has reached.
type Decision int

// Pruner decides how the walk treats each directory it meets. It is defined
// here but implemented by the caller: diskwalk knows nothing about worktrees or
// build artifacts.
//
// Prune is called from every worker concurrently and must be safe for that.
type Pruner interface {
	Prune(path string, entry fs.DirEntry, depth int) Decision
}

// PrunerFunc adapts a function to Pruner.
type PrunerFunc func(path string, entry fs.DirEntry, depth int) Decision

// Leaf is a marked subtree inside a root: a node_modules, a dist, a Pods.
type Leaf struct {
	Path  string
	Name  string
	Bytes int64
	Files int64
}

// Result is the measurement of one root.
type Result struct {
	// Root is the directory that was measured.
	Root string

	// Bytes is the whole tree, leaves included.
	Bytes int64

	Files int64
	Dirs  int64

	// Newest is the most recent modification time anywhere in the tree. It is
	// what staleness filters compare against: a worktree touched an hour ago is
	// live work regardless of what its git state says.
	Newest time.Time

	// Leaves are the marked subtrees, in discovery order.
	Leaves []Leaf

	// Errors counts directories that could not be read. They are reported
	// through Walker.Errors and never abort the walk, because one unreadable
	// directory should not cost the other two hundred gigabytes of the report.
	Errors int
}

// Walker is the configuration for a measurement pass.
type Walker struct {
	// Pruner decides which directories to enter and which to mark. A nil
	// Pruner descends everywhere and marks nothing.
	Pruner Pruner

	// Apparent switches from allocated blocks to file sizes. The default is
	// blocks, which is what du reports and what actually frees on deletion;
	// apparent sizes over-report sparse files and under-report the block floor
	// that dominates a directory of many tiny files.
	Apparent bool

	// FollowSymlinks is off by default. Following them would double-count and,
	// worse, could walk a link right out of the tree being measured.
	FollowSymlinks bool

	// Dedupe skips files already counted through another hard link.
	Dedupe bool

	// MaxDepth bounds descent; zero means unlimited.
	MaxDepth int

	// Errors receives non-fatal read failures. It may be called concurrently.
	Errors func(path string, err error)

	// Progress is ticked as the walk proceeds, for a live counter on a TTY. It
	// may be called concurrently.
	Progress func(files, bytes int64)
}

// accumulator collects one root's totals. Workers touch it from several
// goroutines, so every field is guarded; the mutex is held only for the
// arithmetic, never across a syscall.
type accumulator struct {
	mu     sync.Mutex
	root   string
	bytes  int64
	files  int64
	dirs   int64
	errs   int
	newest time.Time
	leaves []Leaf
}

const (
	// DecisionDescend walks into the directory, attributing what it finds to
	// whatever the current attribution target is.
	DecisionDescend Decision = iota

	// DecisionMark walks into the directory but attributes everything below it
	// to a new leaf as well as to the root. This is how one pass over a
	// worktree yields both its total size and the per-node_modules breakdown
	// inside it, with no second walk.
	DecisionMark

	// DecisionSkip ignores the directory entirely.
	DecisionSkip
)

// Prune implements Pruner.
func (f PrunerFunc) Prune(path string, entry fs.DirEntry, depth int) Decision {
	return f(path, entry, depth)
}

func (a *accumulator) addFile(leaf int, size int64, mod time.Time) {
	a.mu.Lock()

	defer a.mu.Unlock()

	a.bytes += size
	a.files++

	if mod.After(a.newest) {
		a.newest = mod
	}

	if leaf >= 0 && leaf < len(a.leaves) {
		a.leaves[leaf].Bytes += size
		a.leaves[leaf].Files++
	}
}

func (a *accumulator) addDir(mod time.Time) {
	a.mu.Lock()

	defer a.mu.Unlock()

	a.dirs++

	if mod.After(a.newest) {
		a.newest = mod
	}
}

func (a *accumulator) addError() {
	a.mu.Lock()

	defer a.mu.Unlock()

	a.errs++
}

// mark registers a new leaf and returns its index. Nested marks are not
// created: a node_modules inside a node_modules rolls up into the outer one,
// which is what anyone reading the report expects to see.
func (a *accumulator) mark(path, name string) int {
	a.mu.Lock()

	defer a.mu.Unlock()

	a.leaves = append(a.leaves, Leaf{Path: path, Name: name})

	return len(a.leaves) - 1
}

func (a *accumulator) result() Result {
	a.mu.Lock()

	defer a.mu.Unlock()

	return Result{
		Root:   a.root,
		Bytes:  a.bytes,
		Files:  a.files,
		Dirs:   a.dirs,
		Newest: a.newest,
		Leaves: a.leaves,
		Errors: a.errs,
	}
}

package gitwork

import "errors"

var (
	// ErrNotAWorktree reports a path with no git state at all.
	ErrNotAWorktree = errors.New("gitwork: not a worktree")

	// ErrBrokenGitdir reports a .git file whose pointer does not resolve to a
	// matching administrative directory.
	ErrBrokenGitdir = errors.New("gitwork: gitdir pointer is broken")

	// ErrMainWorktree reports an attempt to treat a repository's own working
	// tree as a disposable worktree. It is never removable.
	ErrMainWorktree = errors.New("gitwork: path is the main working tree")

	// ErrTimeout reports that a git command did not finish in time. It is
	// always treated as unsafe: an unanswered status question must never be
	// read as "clean".
	ErrTimeout = errors.New("gitwork: command timed out")
)

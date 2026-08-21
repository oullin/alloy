// Package gitwork inspects git worktrees and decides whether removing one
// would destroy work.
//
// The distinction that drives everything here is between a linked worktree and
// a clone. A linked worktree's .git is a file pointing into a parent
// repository's administrative directory, and deleting the directory leaves a
// dangling registry entry behind — the state git calls "prunable", and the
// reason this package exists. A clone's .git is a real directory and owns
// everything it needs, so it can simply be removed.
//
// Classification costs a single lstat and no subprocess, which matters when a
// scan has a few hundred candidates. Everything more expensive is done once per
// parent repository rather than once per worktree.
package gitwork

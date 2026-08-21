// Package sweep is the only code in treex that deletes anything.
//
// It is deliberately small and deliberately suspicious. Every action is
// re-validated by Guard immediately before the unlink, independently of
// whatever the plan concluded, because a plan is a snapshot and minutes may
// have passed since it was made. A linked worktree is always removed through
// git rather than by unlinking its directory, so treex never creates the
// dangling registry entries it exists to clean up.
package sweep

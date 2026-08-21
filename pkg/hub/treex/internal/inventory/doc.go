// Package inventory finds what is on disk and how big it is.
//
// Discovery runs in three passes so that sizing, by far the most expensive
// part, happens exactly once. First every candidate is enumerated without
// measuring anything; then the git facts are gathered, grouped by parent
// repository so a registry is read once rather than once per worktree; then
// every candidate is measured in a single walk that also yields the build
// artifacts inside it.
//
// Worktree discovery descends until it finds a directory containing a .git and
// then stops. That one rule handles every layout treex meets in the wild,
// including the second tier of worktrees some agents nest under a per-repository
// directory, without any per-agent special casing.
package inventory

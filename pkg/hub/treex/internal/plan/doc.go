// Package plan turns a measured inventory into an ordered list of actions.
//
// It is the last place where anything can be reconsidered before deletion
// starts, and it does three things that matter. Artifacts nested inside a
// worktree that is itself being removed are dropped as redundant, so their
// bytes are not counted twice. Actions are ordered largest first, so a run cut
// short by an interrupt or a limit has still reclaimed the most it could.
// Registry repairs are appended last and deduplicated per repository, because
// pruning before the removals would simply re-dangle the entries.
package plan

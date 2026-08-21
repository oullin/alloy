// Package treex is the root of the treex command, a git-aware reclaimer for
// the disk space AI coding agents leave behind: worktrees, the build artifacts
// inside them, agent scratch directories, and session logs.
//
// The module is deliberately split so that the destructive half is small and
// heavily guarded. config and report are the public contract — the YAML schema
// users write and the JSON shape other tools parse. Everything that decides
// what may be deleted lives under internal, and every unlink is funnelled
// through internal/sweep, which re-validates each path against internal/gitwork
// immediately before acting.
package treex

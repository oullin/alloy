package inventory

import (
	"time"

	"hara.sh/alloy/treex/config"
	"hara.sh/alloy/treex/internal/diskwalk"
	"hara.sh/alloy/treex/internal/gitwork"
)

// Candidate is one reclaimable thing: a worktree, a build directory inside a
// worktree, a cache, a session store, or an orphaned file.
type Candidate struct {
	// Provider is the agent this belongs to, for grouping in the report.
	Provider string

	// Category decides which --categories selection includes it.
	Category config.Category

	// Path is absolute and cleaned.
	Path string

	// Label is how the path is shown, relative to the provider root, so a
	// report stays readable at a hundred entries.
	Label string

	Bytes  int64
	Files  int64
	Newest time.Time

	// Worktree carries the git facts, set only for CategoryWorktree.
	Worktree *gitwork.Worktree

	// Repo is the parent repository whose registry owns this worktree.
	Repo string

	// Risk and Reasons are filled in by the policy.
	Risk    gitwork.Risk
	Reasons []string

	// Leaves are the build artifacts found inside a worktree. They become
	// artifact candidates in their own right, which is what lets treex reclaim
	// from a worktree it has refused to delete.
	Leaves []diskwalk.Leaf

	// Key guards against acting on a path that has been replaced since the
	// scan measured it.
	Key diskwalk.FileKey
}

// Removable reports whether the policy cleared this candidate.
func (c Candidate) Removable() bool {
	return c.Risk == gitwork.RiskSafe
}

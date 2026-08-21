package plan

import (
	"path/filepath"
	"sort"
	"strings"

	"hara.sh/alloy/treex/config"
	"hara.sh/alloy/treex/internal/diskwalk"
	"hara.sh/alloy/treex/internal/gitwork"
	"hara.sh/alloy/treex/internal/inventory"
)

// Kind is what an action does.
type Kind string

// Action is one thing treex will do.
type Action struct {
	Kind     Kind
	Provider string
	Category config.Category
	Path     string
	Label    string
	Bytes    int64
	Repo     string

	// Worktree carries the git facts, so the sweeper knows whether to call git
	// or to unlink directly.
	Worktree *gitwork.Worktree

	Reasons []string
	Key     diskwalk.FileKey
}

// Plan is the ordered work, plus everything that was refused.
type Plan struct {
	Actions []Action
	Blocked []Action

	// Repairs are repositories carrying registry entries whose working tree is
	// already gone.
	Repairs map[string][]gitwork.Entry
}

const (
	// KindRemoveWorktree removes a whole worktree, through git where the
	// worktree is linked.
	KindRemoveWorktree Kind = "remove-worktree"

	// KindRemoveTree removes a directory outright: an artifact, a cache, a
	// session store.
	KindRemoveTree Kind = "remove-tree"

	// KindRemoveFile removes a single orphaned file.
	KindRemoveFile Kind = "remove-file"

	// KindPruneRegistry repairs a repository's worktree registry.
	KindPruneRegistry Kind = "prune-registry"
)

// Bytes totals the plan's reclaimable size.
func (p Plan) Bytes() int64 {
	return total(p.Actions)
}

// BlockedBytes totals what was refused.
func (p Plan) BlockedBytes() int64 {
	return total(p.Blocked)
}

func total(actions []Action) int64 {
	var sum int64

	for _, action := range actions {
		sum += action.Bytes
	}

	return sum
}

// Build turns an inventory into a plan.
func Build(found inventory.Inventory, limit int) Plan {
	actions := make([]Action, 0, len(found.Candidates))
	blocked := make([]Action, 0)

	removed := removedWorktrees(found.Candidates)

	for _, candidate := range found.Candidates {
		action := actionOf(candidate)

		if !candidate.Removable() {
			blocked = append(blocked, action)

			continue
		}

		// An artifact inside a worktree that is going away entirely is already
		// accounted for by that removal; listing it separately would both
		// double-count the bytes and try to delete a path twice.
		if candidate.Category == config.CategoryArtifact && insideAny(candidate.Path, removed) {
			continue
		}

		actions = append(actions, action)
	}

	// Largest first, so an interrupted run has reclaimed as much as it could.
	sort.SliceStable(actions, func(i, j int) bool {
		return actions[i].Bytes > actions[j].Bytes
	})

	if limit > 0 && len(actions) > limit {
		actions = actions[:limit]
	}

	sort.SliceStable(blocked, func(i, j int) bool {
		return blocked[i].Bytes > blocked[j].Bytes
	})

	return Plan{
		Actions: append(actions, repairs(found, actions)...),
		Blocked: blocked,
		Repairs: found.Prunable,
	}
}

// repairs appends one prune per repository the plan touched, or that already
// carries stale entries. They go last because pruning before the removals would
// leave the entries dangling again.
func repairs(found inventory.Inventory, actions []Action) []Action {
	seen := make(map[string]struct{})
	out := make([]Action, 0)

	for _, action := range actions {
		if action.Kind != KindRemoveWorktree || action.Repo == "" {
			continue
		}

		seen[action.Repo] = struct{}{}
	}

	for repo := range found.Prunable {
		seen[repo] = struct{}{}
	}

	repos := make([]string, 0, len(seen))

	for repo := range seen {
		repos = append(repos, repo)
	}

	sort.Strings(repos)

	for _, repo := range repos {
		out = append(out, Action{
			Kind:  KindPruneRegistry,
			Path:  repo,
			Label: repo,
			Repo:  repo,
		})
	}

	return out
}

func actionOf(candidate inventory.Candidate) Action {
	action := Action{
		Provider: candidate.Provider,
		Category: candidate.Category,
		Path:     candidate.Path,
		Label:    candidate.Label,
		Bytes:    candidate.Bytes,
		Repo:     candidate.Repo,
		Worktree: candidate.Worktree,
		Reasons:  candidate.Reasons,
		Key:      candidate.Key,
	}

	switch candidate.Category {
	case config.CategoryWorktree:
		action.Kind = KindRemoveWorktree
	case config.CategoryOrphan:
		action.Kind = KindRemoveFile
	default:
		action.Kind = KindRemoveTree
	}

	return action
}

func removedWorktrees(candidates []inventory.Candidate) []string {
	out := make([]string, 0)

	for _, candidate := range candidates {
		if candidate.Category == config.CategoryWorktree && candidate.Removable() {
			out = append(out, filepath.Clean(candidate.Path))
		}
	}

	return out
}

func insideAny(path string, roots []string) bool {
	cleaned := filepath.Clean(path)

	for _, root := range roots {
		if cleaned == root || strings.HasPrefix(cleaned, root+string(filepath.Separator)) {
			return true
		}
	}

	return false
}

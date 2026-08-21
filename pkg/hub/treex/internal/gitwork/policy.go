package gitwork

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
)

// Risk is how safe a candidate is to remove.
type Risk string

// Policy is the configured set of gates, resolved to absolute paths.
type Policy struct {
	RequireCleanWorktree bool
	RequirePushedCommits bool
	RequireNoStash       bool
	ProtectedBranches    []string
	ProtectPaths         []string

	// Force relaxes the configured gates. It never relaxes the structural ones
	// below, which is the whole point of separating them: --force is for a user
	// who accepts losing their own uncommitted work, not for one who wants to
	// delete a repository's main working tree.
	Force bool
}

const (
	// RiskSafe means nothing would be lost.
	RiskSafe Risk = "safe"

	// RiskBlocked means removal is refused. The reasons say why.
	RiskBlocked Risk = "blocked"
)

// Verdict decides whether a worktree may be removed and explains itself.
//
// The reasons are returned rather than logged because they are the product:
// "blocked" on its own is useless, whereas "111 commits on no remote" tells the
// user exactly what to do about it.
func (p Policy) Verdict(tree Worktree) (Risk, []string) {
	reasons := make([]string, 0, 4)

	// Structural refusals come first and ignore Force entirely.
	if tree.IsMainWorktree() {
		return RiskBlocked, []string{"this is the repository's main working tree"}
	}

	if protected := p.protects(tree.Path); protected != "" {
		return RiskBlocked, []string{fmt.Sprintf("path is inside protected %s", protected)}
	}

	if tree.Kind == KindBroken && tree.Broken == "the gitdir pointer names a different working tree" {
		return RiskBlocked, []string{tree.Broken}
	}

	if tree.Err != nil {
		return RiskBlocked, []string{fmt.Sprintf("git inspection failed: %v", tree.Err)}
	}

	// A broken pointer means the registry entry is already gone, so there is
	// nothing left to damage and nothing left to lose.
	if tree.Kind == KindBroken {
		return RiskSafe, []string{"the parent repository has already forgotten this worktree"}
	}

	if tree.Locked {
		reasons = append(reasons, lockReason(tree))
	}

	if p.RequireCleanWorktree && !p.Force {
		if tree.Dirty > 0 {
			reasons = append(reasons, fmt.Sprintf("%d uncommitted change(s)", tree.Dirty))
		}

		if tree.Untracked > 0 {
			reasons = append(reasons, fmt.Sprintf("%d untracked path(s)", tree.Untracked))
		}
	}

	if p.RequirePushedCommits && !p.Force && tree.Unpushed > 0 {
		reasons = append(reasons, fmt.Sprintf("%d commit(s) not on any remote", tree.Unpushed))
	}

	if p.RequireNoStash && !p.Force && tree.Stashes > 0 {
		// refs/stash lives in the common dir, so every linked worktree of a
		// repository reports the same count. Blocking ten worktrees on one
		// stash helps nobody, so it is surfaced as context there and enforced
		// only where the stash would actually be destroyed.
		if tree.Kind == KindClone {
			reasons = append(reasons, fmt.Sprintf("%d stash entr(y/ies)", tree.Stashes))
		} else {
			reasons = append(reasons, fmt.Sprintf("note: %d stash entr(y/ies) shared with %s", tree.Stashes, tree.MainRepo))
		}
	}

	if branch := strings.TrimSpace(tree.Branch); branch != "" && slices.Contains(p.ProtectedBranches, branch) && !p.Force {
		reasons = append(reasons, fmt.Sprintf("branch %s is protected", branch))
	}

	if blocking := blockingReasons(reasons); len(blocking) > 0 {
		return RiskBlocked, reasons
	}

	return RiskSafe, reasons
}

// blockingReasons filters out the advisory notes, which are shown to the user
// but do not by themselves refuse a removal.
func blockingReasons(reasons []string) []string {
	out := make([]string, 0, len(reasons))

	for _, reason := range reasons {
		if strings.HasPrefix(reason, "note: ") {
			continue
		}

		out = append(out, reason)
	}

	return out
}

// protects returns the protected ancestor of path, if any.
func (p Policy) protects(path string) string {
	cleaned := filepath.Clean(path)

	for _, protected := range p.ProtectPaths {
		guard := filepath.Clean(protected)

		if cleaned == guard || strings.HasPrefix(cleaned, guard+string(filepath.Separator)) {
			return guard
		}
	}

	return ""
}

func lockReason(tree Worktree) string {
	if strings.TrimSpace(tree.LockReason) == "" {
		return "worktree is locked"
	}

	return fmt.Sprintf("worktree is locked: %s", tree.LockReason)
}

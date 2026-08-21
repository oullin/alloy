package gitwork_test

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"hara.sh/alloy/treex/internal/gitwork"
)

func defaultPolicy() gitwork.Policy {
	return gitwork.Policy{
		RequireCleanWorktree: true,
		RequirePushedCommits: true,
		RequireNoStash:       true,
		ProtectedBranches:    []string{"main", "master"},
	}
}

func TestVerdictAllowsACleanPushedWorktree(t *testing.T) {
	t.Parallel()

	risk, reasons := defaultPolicy().Verdict(gitwork.Worktree{
		Path:     "/home/u/.codex/worktrees/done",
		Kind:     gitwork.KindClone,
		MainRepo: "/home/u/.codex/worktrees/done",
		Branch:   "feature",
	})

	if risk != gitwork.RiskSafe {
		t.Fatalf("Risk = %q, want %q (reasons: %v)", risk, gitwork.RiskSafe, reasons)
	}
}

func TestVerdictBlocksUnpushedCommits(t *testing.T) {
	t.Parallel()

	risk, reasons := defaultPolicy().Verdict(gitwork.Worktree{
		Path:     "/home/u/.claude/worktrees/plan-179",
		Kind:     gitwork.KindLinked,
		MainRepo: "/home/u/Sites/agents",
		Branch:   "plan/179",
		Unpushed: 111,
	})

	if risk != gitwork.RiskBlocked {
		t.Fatalf("Risk = %q, want %q", risk, gitwork.RiskBlocked)
	}

	if !mentions(reasons, "111 commit") {
		t.Fatalf("reasons = %v, want the unpushed count named", reasons)
	}
}

func TestVerdictBlocksUncommittedChanges(t *testing.T) {
	t.Parallel()

	risk, reasons := defaultPolicy().Verdict(gitwork.Worktree{
		Path:     "/home/u/.codex/worktrees/ledger",
		Kind:     gitwork.KindClone,
		MainRepo: "/home/u/.codex/worktrees/ledger/.git",
		Branch:   "chore/reconcile",
		Dirty:    4,
	})

	if risk != gitwork.RiskBlocked {
		t.Fatalf("Risk = %q, want %q", risk, gitwork.RiskBlocked)
	}

	if !mentions(reasons, "4 uncommitted") {
		t.Fatalf("reasons = %v, want the change count named", reasons)
	}
}

func TestVerdictNeverRemovesTheMainWorkingTree(t *testing.T) {
	t.Parallel()

	policy := defaultPolicy()
	policy.Force = true

	// A repository with linked worktrees hanging off it: removing the primary
	// working tree would orphan every one of them.
	risk, reasons := policy.Verdict(gitwork.Worktree{
		Path:     "/home/u/Sites/alloy",
		Kind:     gitwork.KindLinked,
		MainRepo: "/home/u/Sites/alloy",
	})

	// Force relaxes the gates a user owns; it must not relax a structural one.
	if risk != gitwork.RiskBlocked {
		t.Fatalf("Risk = %q with Force, want %q", risk, gitwork.RiskBlocked)
	}

	if !mentions(reasons, "main working tree") {
		t.Fatalf("reasons = %v, want the main working tree named", reasons)
	}
}

// A clone is always its own main working tree, so applying that refusal to one
// would make every clone permanently unremovable — and clones are the single
// largest category of agent debris there is.
func TestVerdictStillAllowsACleanClone(t *testing.T) {
	t.Parallel()

	clone := "/home/u/.codex/worktrees/finished"

	risk, reasons := defaultPolicy().Verdict(gitwork.Worktree{
		Path:     clone,
		Kind:     gitwork.KindClone,
		MainRepo: clone,
		Branch:   "feature",
	})

	if risk != gitwork.RiskSafe {
		t.Fatalf("Risk = %q, want %q (reasons: %v)", risk, gitwork.RiskSafe, reasons)
	}
}

func TestVerdictNeverTouchesAProtectedPath(t *testing.T) {
	t.Parallel()

	policy := defaultPolicy()
	policy.Force = true
	policy.ProtectPaths = []string{filepath.Clean("/home/u/Sites")}

	risk, reasons := policy.Verdict(gitwork.Worktree{
		Path:     "/home/u/Sites/alloy/nested",
		Kind:     gitwork.KindClone,
		MainRepo: "/home/u/Sites/alloy/nested/.git",
	})

	if risk != gitwork.RiskBlocked {
		t.Fatalf("Risk = %q with Force, want %q", risk, gitwork.RiskBlocked)
	}

	if !mentions(reasons, "protected") {
		t.Fatalf("reasons = %v, want the protected path named", reasons)
	}
}

func TestVerdictBlocksAProtectedBranch(t *testing.T) {
	t.Parallel()

	risk, _ := defaultPolicy().Verdict(gitwork.Worktree{
		Path:     "/home/u/.codex/worktrees/trunk",
		Kind:     gitwork.KindClone,
		MainRepo: "/home/u/.codex/worktrees/trunk/.git",
		Branch:   "main",
	})

	if risk != gitwork.RiskBlocked {
		t.Fatalf("Risk = %q, want %q", risk, gitwork.RiskBlocked)
	}
}

func TestVerdictForceReleasesTheGatesAUserOwns(t *testing.T) {
	t.Parallel()

	policy := defaultPolicy()
	policy.Force = true

	risk, _ := policy.Verdict(gitwork.Worktree{
		Path:     "/home/u/.codex/worktrees/ledger",
		Kind:     gitwork.KindClone,
		MainRepo: "/home/u/.codex/worktrees/ledger/.git",
		Branch:   "chore/reconcile",
		Dirty:    4,
		Unpushed: 9,
	})

	if risk != gitwork.RiskSafe {
		t.Fatalf("Risk = %q with Force, want %q", risk, gitwork.RiskSafe)
	}
}

func TestVerdictTreatsAFailedInspectionAsUnsafe(t *testing.T) {
	t.Parallel()

	risk, reasons := defaultPolicy().Verdict(gitwork.Worktree{
		Path:     "/home/u/.codex/worktrees/stalled",
		Kind:     gitwork.KindClone,
		MainRepo: "/home/u/.codex/worktrees/stalled/.git",
		Err:      errors.New("timed out"),
	})

	// An unanswered question is never a yes.
	if risk != gitwork.RiskBlocked {
		t.Fatalf("Risk = %q, want %q", risk, gitwork.RiskBlocked)
	}

	if !mentions(reasons, "inspection failed") {
		t.Fatalf("reasons = %v, want the failure surfaced", reasons)
	}
}

func TestVerdictAllowsAWorktreeItsParentHasForgotten(t *testing.T) {
	t.Parallel()

	risk, _ := defaultPolicy().Verdict(gitwork.Worktree{
		Path:   "/home/u/.claude/worktrees/ghost",
		Kind:   gitwork.KindBroken,
		Broken: "the parent repository's worktree entry is gone",
	})

	// There is no registry left to corrupt and no commits left to reach.
	if risk != gitwork.RiskSafe {
		t.Fatalf("Risk = %q, want %q", risk, gitwork.RiskSafe)
	}
}

func TestVerdictBlocksAGitdirPointerMismatch(t *testing.T) {
	t.Parallel()

	policy := defaultPolicy()
	policy.Force = true

	risk, _ := policy.Verdict(gitwork.Worktree{
		Path:   "/home/u/.claude/worktrees/confused",
		Kind:   gitwork.KindBroken,
		Broken: "the gitdir pointer names a different working tree",
	})

	if risk != gitwork.RiskBlocked {
		t.Fatalf("Risk = %q with Force, want %q", risk, gitwork.RiskBlocked)
	}
}

// A stash lives in the common directory, so it is shared by every linked
// worktree of a repository. Blocking ten of them on one stash would make the
// tool useless on exactly the repositories that need it most.
func TestVerdictDoesNotBlockLinkedWorktreesOnASharedStash(t *testing.T) {
	t.Parallel()

	risk, reasons := defaultPolicy().Verdict(gitwork.Worktree{
		Path:     "/home/u/.claude/worktrees/one",
		Kind:     gitwork.KindLinked,
		MainRepo: "/home/u/Sites/agents",
		Branch:   "feature",
		Stashes:  2,
	})

	if risk != gitwork.RiskSafe {
		t.Fatalf("Risk = %q, want %q (reasons: %v)", risk, gitwork.RiskSafe, reasons)
	}

	if !mentions(reasons, "stash") {
		t.Fatalf("reasons = %v, want the shared stash surfaced as context", reasons)
	}
}

func TestVerdictBlocksACloneWithItsOwnStash(t *testing.T) {
	t.Parallel()

	risk, _ := defaultPolicy().Verdict(gitwork.Worktree{
		Path:     "/home/u/.codex/worktrees/stashed",
		Kind:     gitwork.KindClone,
		MainRepo: "/home/u/.codex/worktrees/stashed/.git",
		Branch:   "feature",
		Stashes:  1,
	})

	if risk != gitwork.RiskBlocked {
		t.Fatalf("Risk = %q, want %q", risk, gitwork.RiskBlocked)
	}
}

func mentions(reasons []string, want string) bool {
	for _, reason := range reasons {
		if strings.Contains(reason, want) {
			return true
		}
	}

	return false
}

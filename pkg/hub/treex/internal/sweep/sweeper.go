package sweep

import (
	"context"
	"fmt"
	"os"

	"hara.sh/alloy/treex/internal/gitwork"
	"hara.sh/alloy/treex/internal/plan"
)

// Outcome records what happened to one action.
type Outcome struct {
	Action plan.Action
	Bytes  int64
	Err    error
}

// Result is a whole sweep.
type Result struct {
	Outcomes []Outcome
	Removed  int64
	Failed   int
}

// Sweeper executes a plan.
type Sweeper struct {
	Guard  Guard
	Runner gitwork.Runner

	// DryRun reports what would happen without touching anything. It is the
	// default everywhere: --apply is the opt-in, not --dry-run.
	DryRun bool

	// Force is passed to git worktree remove, which otherwise refuses a tree
	// with local modifications. It is only ever set when the policy has already
	// cleared the removal, or when the user explicitly accepted the loss.
	Force bool

	// Progress is called before each action so a long sweep is legible.
	Progress func(action plan.Action)
}

// Done reports whether the action succeeded.
func (o Outcome) Done() bool {
	return o.Err == nil
}

// Run executes every action, stopping only if the context is cancelled.
//
// A single failure does not abort the sweep: one worktree git refuses to remove
// should not cost the two hundred gigabytes behind it in the queue.
func (s Sweeper) Run(ctx context.Context, work plan.Plan) Result {
	result := Result{Outcomes: make([]Outcome, 0, len(work.Actions))}

	for _, action := range work.Actions {
		if ctx.Err() != nil {
			break
		}

		if s.Progress != nil {
			s.Progress(action)
		}

		outcome := Outcome{Action: action}

		if err := s.execute(ctx, action); err != nil {
			outcome.Err = err
			result.Failed++
		} else {
			outcome.Bytes = action.Bytes
			result.Removed += action.Bytes
		}

		result.Outcomes = append(result.Outcomes, outcome)
	}

	return result
}

func (s Sweeper) execute(ctx context.Context, action plan.Action) error {
	// A registry prune touches no user data and has no path to guard.
	if action.Kind == plan.KindPruneRegistry {
		return s.prune(ctx, action)
	}

	if err := s.Guard.Check(action); err != nil {
		return err
	}

	if s.DryRun {
		return nil
	}

	if action.Kind == plan.KindRemoveWorktree {
		return s.removeWorktree(ctx, action)
	}

	return remove(action.Path)
}

// removeWorktree deletes a worktree the way its kind requires.
//
// This is the whole reason treex exists rather than a shell alias. A linked
// worktree must go through git, because unlinking its directory leaves the
// parent repository holding a registry entry that points at nothing — the exact
// damage this tool was written to clean up.
func (s Sweeper) removeWorktree(ctx context.Context, action plan.Action) error {
	tree := action.Worktree

	if tree == nil || tree.Kind != gitwork.KindLinked {
		if err := remove(action.Path); err != nil {
			return err
		}

		// A clone owns its own git directory, but a worktree whose pointer was
		// broken may still have a live parent carrying a stale entry.
		if tree != nil && tree.Kind == gitwork.KindBroken && tree.MainRepo != "" {
			_, _ = s.Runner.Run(ctx, tree.MainRepo, "worktree", "prune")
		}

		return nil
	}

	args := []string{"worktree", "remove"}

	if s.Force {
		args = append(args, "--force")
	}

	if _, err := s.Runner.Run(ctx, tree.MainRepo, append(args, action.Path)...); err != nil {
		// git refuses a worktree it considers dirty. Falling back to an unlink
		// would be the corruption this tool exists to prevent, so the failure
		// is surfaced and the tree is left alone.
		return fmt.Errorf("remove linked worktree %s: %w", action.Path, err)
	}

	return nil
}

func (s Sweeper) prune(ctx context.Context, action plan.Action) error {
	if s.DryRun || action.Repo == "" {
		return nil
	}

	if _, err := s.Runner.Run(ctx, action.Repo, "worktree", "prune"); err != nil {
		return fmt.Errorf("prune worktree registry in %s: %w", action.Repo, err)
	}

	return nil
}

func remove(path string) error {
	if err := os.RemoveAll(path); err != nil {
		return fmt.Errorf("remove %s: %w", path, err)
	}

	return nil
}

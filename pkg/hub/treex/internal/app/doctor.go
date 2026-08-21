package app

import (
	"context"
	"fmt"

	"hara.sh/alloy/treex/config"
	"hara.sh/alloy/treex/internal/console"
	"hara.sh/alloy/treex/internal/gitwork"
	"hara.sh/alloy/treex/internal/inventory"
	"hara.sh/alloy/treex/internal/plan"
)

// runDoctor reports the state a scan cannot act on: registry entries pointing
// at working trees that are already gone, and worktrees holding work that will
// keep them blocked until someone deals with it.
//
// It is the answer to the most common first reaction to treex, which is "why is
// so much of this blocked?".
func (d *deps) runDoctor(ctx context.Context, args []string) int {
	opts, err := parseOptions(args)

	if err != nil {
		return d.usageError(err)
	}

	session, err := d.session(opts, []config.Category{config.CategoryWorktree})

	if err != nil {
		return d.reportError(err)
	}

	found, err := session.scanner.Scan(ctx)

	if err != nil {
		return d.reportError(err)
	}

	printer := console.NewPrinter(d.stdout, console.DetectColor(d.stdout))

	d.reportPrunable(ctx, printer, session, found.Prunable, opts.apply)
	d.reportBlocked(printer, found.Candidates)

	return 0
}

func (d *deps) reportPrunable(
	ctx context.Context,
	printer *console.Printer,
	session *session,
	prunable map[string][]gitwork.Entry,
	apply bool,
) {
	printer.Section("registry health")

	if len(prunable) == 0 {
		printer.Success("no stale worktree entries")

		return
	}

	for repo, entries := range prunable {
		printer.Warning(fmt.Sprintf("%s carries %d stale entr(y/ies)", repo, len(entries)))

		for _, entry := range entries {
			printer.Dim(fmt.Sprintf("    %s", entry.Path))
		}

		if !apply {
			continue
		}

		sweeper := session.pruner()

		result := sweeper.Run(ctx, plan.Plan{
			Actions: []plan.Action{{Kind: plan.KindPruneRegistry, Path: repo, Repo: repo}},
		})

		if result.Failed > 0 {
			printer.Warning(fmt.Sprintf("    could not prune %s", repo))

			continue
		}

		printer.Success(fmt.Sprintf("    pruned %s", repo))
	}

	if !apply {
		printer.Dim("run treex doctor --apply to repair these")
	}
}

func (d *deps) reportBlocked(printer *console.Printer, candidates []inventory.Candidate) {
	printer.Section("worktrees holding work")

	blocked := 0

	for _, candidate := range candidates {
		if candidate.Removable() || candidate.Worktree == nil {
			continue
		}

		blocked++

		printer.Warning(fmt.Sprintf("%s", candidate.Label))

		for _, reason := range candidate.Reasons {
			printer.Dim(fmt.Sprintf("    %s", reason))
		}
	}

	if blocked == 0 {
		printer.Success("nothing is holding unpushed or uncommitted work")

		return
	}

	printer.Blank()
	printer.Dim("these are excluded from clean until the work is pushed or discarded;")
	printer.Dim("treex clean --categories artifact still reclaims the build output inside them")
}

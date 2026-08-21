package app

import (
	"context"
	"errors"
	"fmt"

	"hara.sh/alloy/treex/config"
	"hara.sh/alloy/treex/internal/console"
	"hara.sh/alloy/treex/internal/plan"
	"hara.sh/alloy/treex/internal/sweep"
	"hara.sh/alloy/treex/report"
)

var (
	errApplyOnScan = errors.New("scan never deletes; use treex clean --apply")

	// errNeedsConfirmation is what stops a scripted or piped clean from
	// deleting without anyone having said yes. Dry-run is the default, --apply
	// is the opt-in, and outside a terminal --yes is required on top.
	errNeedsConfirmation = errors.New("clean --apply needs a terminal to confirm on; pass --yes to proceed unattended")
)

func (d *deps) runClean(ctx context.Context, args []string) int {
	opts, err := parseOptions(args)

	if err != nil {
		return d.usageError(err)
	}

	categories, err := config.ParseCategories(opts.categories)

	if err != nil {
		return d.usageError(err)
	}

	session, err := d.session(opts, categories)

	if err != nil {
		return d.reportError(err)
	}

	found, work, err := session.scanFor(ctx)

	if err != nil {
		return d.reportError(err)
	}

	// Without --apply this is where the command stops: the plan that was just
	// built is exactly the plan --apply would execute.
	if !opts.apply {
		out := session.projection(d.version, report.ModeDry).build(found, work, nil)

		return d.render(session, out, opts.format)
	}

	if err := d.confirm(session, work, opts); err != nil {
		return d.reportError(err)
	}

	sweeper := sweep.Sweeper{
		Guard:  session.guard(),
		Runner: session.scanner.Runner,
		Force:  opts.force,
	}

	result := sweeper.Run(ctx, work)

	out := session.projection(d.version, report.ModeApply).build(found, work, &result)

	return d.render(session, out, opts.format)
}

// confirm decides whether the sweep may proceed.
//
// Three postures are supported deliberately: an interactive run prompts, an
// unattended run needs --yes, and a run with nothing to do stops early rather
// than asking a pointless question.
func (d *deps) confirm(session *session, work plan.Plan, opts options) error {
	if len(work.Actions) == 0 {
		return nil
	}

	if opts.yes {
		return nil
	}

	if !console.IsTerminal(d.stdin) {
		return errNeedsConfirmation
	}

	printer := console.NewPrinter(d.stdout, console.DetectColor(d.stdout))

	d.summarise(printer, work)

	question := fmt.Sprintf(
		"Remove %d item(s), freeing %s?",
		len(work.Actions),
		report.Bytes(work.Bytes()),
	)

	if !printer.Prompt(d.stdin, question) {
		return errAborted
	}

	return nil
}

func (d *deps) summarise(printer *console.Printer, work plan.Plan) {
	printer.Section("about to remove")

	shown := 0

	for _, action := range work.Actions {
		if action.Kind == plan.KindPruneRegistry {
			continue
		}

		if shown >= 10 {
			printer.Dim(fmt.Sprintf("... and %d more", len(work.Actions)-shown))

			break
		}

		printer.Line(fmt.Sprintf("%10s  %s", report.Bytes(action.Bytes), action.Label))

		shown++
	}
}

var errAborted = errors.New("aborted")

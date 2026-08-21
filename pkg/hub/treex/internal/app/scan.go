package app

import (
	"context"

	"hara.sh/alloy/treex/config"
	"hara.sh/alloy/treex/report"
)

func (d *deps) runScan(ctx context.Context, args []string) int {
	opts, err := parseOptions(args)

	if err != nil {
		return d.usageError(err)
	}

	if opts.apply {
		return d.usageError(errApplyOnScan)
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

	out := session.projection(d.version, report.ModeScan).build(found, work, nil)

	return d.render(session, out, opts.format)
}

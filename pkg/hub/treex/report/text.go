package report

import (
	"fmt"
	"io"
	"strings"
)

const defaultLimit = 8

func (r Renderer) renderText(w io.Writer, report Report) error {
	r.head(w, report)

	for _, provider := range report.Providers {
		r.provider(w, provider)
	}

	r.repairs(w, report)
	r.warnings(w, report)
	r.summary(w, report)

	return nil
}

func (r Renderer) head(w io.Writer, report Report) {
	write(w, "\n==> treex %s\n", report.Mode)

	if report.Config != "" {
		write(w, "    %-14s %s\n", "config", report.Config)
	} else {
		write(w, "    %-14s %s\n", "config", "(built-in defaults)")
	}

	names := make([]string, 0, len(report.Providers))

	for _, provider := range report.Providers {
		names = append(names, provider.Name)
	}

	if len(names) > 0 {
		write(w, "    %-14s %s\n", "providers", strings.Join(names, ", "))
	}
}

func (r Renderer) provider(w io.Writer, provider Provider) {
	if len(provider.Entries) == 0 {
		return
	}

	write(w, "\n==> %-44s %14s\n", provider.Name, provider.Bytes)

	for _, category := range provider.Categories {
		if category.Count == 0 {
			continue
		}

		write(w, "    %-12s %12s  %4d entries   reclaimable %s\n",
			category.Name, category.Bytes, category.Count, category.Reclaimable)
	}

	entries := provider.Sorted()
	limit := r.Limit

	if limit <= 0 {
		limit = defaultLimit
	}

	if r.Verbose || len(entries) < limit {
		limit = len(entries)
	}

	if limit == 0 {
		return
	}

	write(w, "\n")

	for _, entry := range entries[:limit] {
		r.entry(w, entry)
	}

	if remaining := len(entries) - limit; remaining > 0 {
		write(w, "      %s\n", fmt.Sprintf("... and %d more (--verbose to list)", remaining))
	}
}

func (r Renderer) entry(w io.Writer, entry Entry) {
	marker := " "

	switch {
	case entry.Error != "":
		marker = "!"
	case entry.Removed:
		marker = "-"
	case entry.Risk == RiskBlocked:
		marker = "x"
	}

	write(w, "  %s %10s  %-52s %s\n", marker, entry.Bytes, truncate(entry.Label, 52), detail(entry))

	if entry.Error != "" {
		write(w, "      %s\n", entry.Error)
	}

	if r.Explain && entry.Risk == RiskBlocked && len(entry.Reasons) > 0 {
		for _, reason := range entry.Reasons {
			write(w, "      %s\n", reason)
		}
	}
}

func detail(entry Entry) string {
	parts := make([]string, 0, 3)

	if entry.Kind != "" {
		parts = append(parts, entry.Kind)
	}

	if entry.Branch != "" {
		parts = append(parts, entry.Branch)
	}

	if entry.Risk == RiskBlocked && len(entry.Reasons) > 0 {
		parts = append(parts, entry.Reasons[0])
	}

	return strings.Join(parts, "  ")
}

func (r Renderer) repairs(w io.Writer, report Report) {
	if len(report.Repairs) == 0 {
		return
	}

	write(w, "\n==> registry repairs\n")

	for _, repair := range report.Repairs {
		status := "needs pruning"

		if repair.Fixed {
			status = "pruned"
		}

		write(w, "    %-52s %d stale entr(y/ies), %s\n", truncate(repair.Repo, 52), len(repair.Entries), status)
	}
}

func (r Renderer) warnings(w io.Writer, report Report) {
	if len(report.Warnings) == 0 {
		return
	}

	write(w, "\n==> warnings\n")

	for _, warning := range report.Warnings {
		write(w, "    %s\n", warning)
	}
}

func (r Renderer) summary(w io.Writer, report Report) {
	write(w, "\n==> summary\n")
	write(w, "    %-14s %12s   %d entries, %d files\n", "scanned", report.Totals.Scanned, report.Totals.Entries, report.Totals.Files)
	write(w, "    %-14s %12s\n", "reclaimable", report.Totals.Reclaimable)

	if report.Totals.BlockedRows > 0 {
		write(w, "    %-14s %12s   %d entries%s\n", "blocked", report.Totals.Blocked, report.Totals.BlockedRows, explainHint(report))
	}

	if report.Mode == ModeApply {
		write(w, "    %-14s %12s\n", "removed", report.Totals.Removed)
	}

	if report.Elapsed != "" {
		write(w, "    %-14s %12s\n", "elapsed", report.Elapsed)
	}

	if report.Mode == ModeDry {
		write(w, "\n    Nothing was removed. Re-run with --apply to act on this plan.\n")
	}
}

func explainHint(report Report) string {
	if report.Mode == ModeApply {
		return ""
	}

	return "  (--explain for why)"
}

func truncate(value string, width int) string {
	if len(value) <= width {
		return value
	}

	if width <= 3 {
		return value[:width]
	}

	// Keep the tail: the distinguishing part of a worktree path is its name,
	// not the directory prefix every entry shares.
	return "..." + value[len(value)-(width-3):]
}

func write(w io.Writer, format string, args ...any) {
	_, _ = fmt.Fprintf(w, format, args...)
}

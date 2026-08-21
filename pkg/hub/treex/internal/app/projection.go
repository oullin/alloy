package app

import (
	"sort"

	"hara.sh/alloy/treex/config"
	"hara.sh/alloy/treex/internal/gitwork"
	"hara.sh/alloy/treex/internal/inventory"
	"hara.sh/alloy/treex/internal/plan"
	"hara.sh/alloy/treex/internal/sweep"
	"hara.sh/alloy/treex/report"
)

// projection converts internal state into the public report shape.
//
// It lives here rather than in report so that report can stay free of internal
// dependencies: the JSON other tools parse is a contract, and it should not be
// possible to change it by refactoring an internal package.
type projection struct {
	version string
	config  string
	mode    report.Mode
	roots   map[string]string
}

func (p projection) build(found inventory.Inventory, work plan.Plan, result *sweep.Result) report.Report {
	out := report.Report{
		Mode:      p.mode,
		Version:   p.version,
		Config:    p.config,
		Elapsed:   found.Duration.Round(elapsedPrecision).String(),
		Providers: p.providers(found, work, result),
	}

	out.Totals = p.totals(found, work, result)
	out.Repairs = p.repairs(work, result)

	return out
}

func (p projection) providers(found inventory.Inventory, work plan.Plan, result *sweep.Result) []report.Provider {
	selected := selectedPaths(work)
	outcomes := outcomeIndex(result)

	grouped := make(map[string]*report.Provider)
	order := make([]string, 0, len(p.roots))

	for _, candidate := range found.Candidates {
		provider, ok := grouped[candidate.Provider]

		if !ok {
			provider = &report.Provider{Name: candidate.Provider, Root: p.roots[candidate.Provider]}
			grouped[candidate.Provider] = provider

			order = append(order, candidate.Provider)
		}

		entry := p.entry(candidate, selected, outcomes)

		provider.Entries = append(provider.Entries, entry)

		// A provider's size is what it occupies on disk. Artifact entries are
		// a breakdown of space already counted by the worktree containing them.
		if candidate.Category != config.CategoryArtifact {
			provider.Bytes += entry.Bytes
		}
	}

	out := make([]report.Provider, 0, len(order))

	for _, name := range order {
		provider := grouped[name]
		provider.Categories = categorise(provider.Entries)

		out = append(out, *provider)
	}

	sort.SliceStable(out, func(i, j int) bool {
		return out[i].Bytes > out[j].Bytes
	})

	return out
}

func (p projection) entry(
	candidate inventory.Candidate,
	selected map[string]struct{},
	outcomes map[string]sweep.Outcome,
) report.Entry {
	entry := report.Entry{
		Path:     candidate.Path,
		Label:    candidate.Label,
		Category: string(candidate.Category),
		Bytes:    report.Bytes(candidate.Bytes),
		Risk:     riskOf(candidate.Risk),
		Reasons:  candidate.Reasons,
		Repo:     candidate.Repo,
		Newest:   candidate.Newest,
	}

	if candidate.Worktree != nil {
		entry.Kind = string(candidate.Worktree.Kind)
		entry.Branch = candidate.Worktree.Branch
	}

	if _, chosen := selected[candidate.Path]; !chosen {
		return entry
	}

	outcome, acted := outcomes[candidate.Path]

	if !acted {
		return entry
	}

	if outcome.Err != nil {
		entry.Error = outcome.Err.Error()

		return entry
	}

	entry.Removed = true

	return entry
}

func (p projection) totals(found inventory.Inventory, work plan.Plan, result *sweep.Result) report.Totals {
	totals := report.Totals{
		Scanned:     report.Bytes(found.Scanned),
		Reclaimable: report.Bytes(work.Bytes()),
		Blocked:     report.Bytes(work.BlockedBytes()),
		Files:       found.Files,
		Entries:     len(found.Candidates),
		BlockedRows: len(work.Blocked),
	}

	if result != nil {
		totals.Removed = report.Bytes(result.Removed)
	}

	return totals
}

func (p projection) repairs(work plan.Plan, result *sweep.Result) []report.Repair {
	if len(work.Repairs) == 0 {
		return nil
	}

	repos := make([]string, 0, len(work.Repairs))

	for repo := range work.Repairs {
		repos = append(repos, repo)
	}

	sort.Strings(repos)

	out := make([]report.Repair, 0, len(repos))

	for _, repo := range repos {
		entries := make([]string, 0, len(work.Repairs[repo]))

		for _, entry := range work.Repairs[repo] {
			entries = append(entries, entry.Path)
		}

		out = append(out, report.Repair{Repo: repo, Entries: entries, Fixed: pruned(result, repo)})
	}

	return out
}

func pruned(result *sweep.Result, repo string) bool {
	if result == nil {
		return false
	}

	for _, outcome := range result.Outcomes {
		if outcome.Action.Kind == plan.KindPruneRegistry && outcome.Action.Repo == repo {
			return outcome.Done()
		}
	}

	return false
}

func categorise(entries []report.Entry) []report.Category {
	totals := make(map[string]*report.Category)

	for _, entry := range entries {
		category, ok := totals[entry.Category]

		if !ok {
			category = &report.Category{Name: entry.Category}
			totals[entry.Category] = category
		}

		category.Count++
		category.Bytes += entry.Bytes

		if entry.Risk == report.RiskBlocked {
			category.Blocked += entry.Bytes

			continue
		}

		category.Reclaimable += entry.Bytes
	}

	out := make([]report.Category, 0, len(totals))

	// Report categories in their declared order rather than map order, so two
	// runs of the same scan render identically.
	for _, name := range config.Categories {
		if category, ok := totals[string(name)]; ok {
			out = append(out, *category)
		}
	}

	return out
}

func selectedPaths(work plan.Plan) map[string]struct{} {
	out := make(map[string]struct{}, len(work.Actions))

	for _, action := range work.Actions {
		out[action.Path] = struct{}{}
	}

	return out
}

func outcomeIndex(result *sweep.Result) map[string]sweep.Outcome {
	if result == nil {
		return nil
	}

	out := make(map[string]sweep.Outcome, len(result.Outcomes))

	for _, outcome := range result.Outcomes {
		out[outcome.Action.Path] = outcome
	}

	return out
}

func riskOf(risk gitwork.Risk) report.Risk {
	if risk == gitwork.RiskBlocked {
		return report.RiskBlocked
	}

	return report.RiskSafe
}

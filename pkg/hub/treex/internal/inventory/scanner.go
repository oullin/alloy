package inventory

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"time"

	"hara.sh/alloy/treex/config"
	"hara.sh/alloy/treex/internal/diskwalk"
	"hara.sh/alloy/treex/internal/gitwork"
	"hara.sh/alloy/treex/internal/proc"
)

// Scanner enumerates and measures everything the selected providers own.
type Scanner struct {
	Config     config.Config
	Providers  []config.Resolved
	Categories []config.Category
	Pool       *diskwalk.Pool
	Inspector  gitwork.Inspector
	Policy     gitwork.Policy
	Runner     gitwork.Runner

	// Now is injected so staleness filtering is deterministic under test.
	Now func() time.Time

	// Errors receives non-fatal read failures.
	Errors func(path string, err error)
}

// Inventory is the finished scan.
type Inventory struct {
	Candidates []Candidate

	// Prunable lists registry entries whose working tree is already gone,
	// keyed by the repository that still carries them. These are damage from
	// earlier manual deletes, and repairing them is what doctor does.
	Prunable map[string][]gitwork.Entry

	Scanned  int64
	Files    int64
	Duration time.Duration
}

// Scan enumerates, inspects, and measures. The context is honoured throughout,
// so a cancelled scan returns whatever it had rather than nothing.
func (s Scanner) Scan(ctx context.Context) (Inventory, error) {
	started := time.Now()

	found := s.enumerate(ctx)

	if err := ctx.Err(); err != nil {
		return Inventory{}, err
	}

	prunable := s.inspect(ctx, found)

	if err := ctx.Err(); err != nil {
		return Inventory{}, err
	}

	measured, err := s.measure(ctx, found)

	if err != nil && !isCancelled(err) {
		return Inventory{}, err
	}

	candidates := s.expand(measured)

	s.judge(candidates)

	filtered := s.filter(candidates)

	sort.SliceStable(filtered, func(i, j int) bool {
		return filtered[i].Bytes > filtered[j].Bytes
	})

	inventory := Inventory{
		Candidates: filtered,
		Prunable:   prunable,
		Duration:   time.Since(started),
	}

	for _, candidate := range filtered {
		// Artifacts are nested inside a worktree that has already contributed
		// its full size, so adding them again would report a tree half as large
		// again as it is on disk.
		if candidate.Category == config.CategoryArtifact {
			continue
		}

		inventory.Scanned += candidate.Bytes
		inventory.Files += candidate.Files
	}

	return inventory, nil
}

// enumerate walks the provider roots without measuring anything, which keeps
// this pass cheap enough to run over every configured provider.
func (s Scanner) enumerate(ctx context.Context) []Candidate {
	out := make([]Candidate, 0, 64)

	for _, provider := range s.Providers {
		if ctx.Err() != nil {
			return out
		}

		if !dirExists(provider.Root) {
			continue
		}

		if s.wants(config.CategoryWorktree) || s.wants(config.CategoryArtifact) {
			out = append(out, s.worktrees(provider)...)
		}

		if s.wants(config.CategoryCache) {
			out = append(out, s.named(provider, provider.Caches, config.CategoryCache)...)
		}

		if s.wants(config.CategorySession) {
			out = append(out, s.named(provider, provider.Sessions, config.CategorySession)...)
		}

		if s.wants(config.CategoryOrphan) {
			out = append(out, s.orphans(provider)...)
		}
	}

	return out
}

// worktrees descends each configured source until it meets a directory holding
// a .git, which it takes as a worktree and does not enter. Depth is bounded per
// source because the layouts differ: some agents keep worktrees directly under
// one directory, others nest a whole second tier by repository.
func (s Scanner) worktrees(provider config.Resolved) []Candidate {
	out := make([]Candidate, 0, 16)

	for _, source := range provider.Worktrees {
		root := filepath.Join(provider.Root, source.Path)

		if !dirExists(root) {
			continue
		}

		s.descend(provider, root, root, source.Depth, &out)
	}

	return out
}

// descend walks a worktree source looking for directories that hold a .git,
// which it takes as worktrees and does not enter.
//
// A directory with no .git at the depth limit is still recorded, as a plain
// candidate. Agents nest their scratch in ways that vary and change: some group
// worktrees by repository, some leave loose build directories beside them.
// Recording the leaf rather than silently dropping it is what stops a scan from
// under-reporting a tree by tens of gigabytes.
func (s Scanner) descend(provider config.Resolved, root, dir string, depth int, out *[]Candidate) {
	if depth <= 0 {
		return
	}

	entries, err := os.ReadDir(dir)

	if err != nil {
		s.report(dir, err)

		return
	}

	for _, entry := range entries {
		if !entry.IsDir() || entry.Type()&os.ModeSymlink != 0 {
			continue
		}

		path := filepath.Join(dir, entry.Name())

		if _, err := os.Lstat(filepath.Join(path, ".git")); err == nil {
			*out = append(*out, Candidate{
				Provider: provider.Name,
				Category: config.CategoryWorktree,
				Path:     path,
				Label:    label(provider.Root, path),
			})

			continue
		}

		if depth == 1 {
			*out = append(*out, Candidate{
				Provider: provider.Name,
				Category: config.CategoryWorktree,
				Path:     path,
				Label:    label(provider.Root, path),
			})

			continue
		}

		s.descend(provider, root, path, depth-1, out)
	}
}

func (s Scanner) named(provider config.Resolved, names []string, category config.Category) []Candidate {
	out := make([]Candidate, 0, len(names))

	for _, name := range names {
		path := filepath.Clean(filepath.Join(provider.Root, name))

		// A "." entry means the provider root itself is the cache, which is
		// true of the single-purpose toolchain directories.
		if !strings.HasPrefix(path, provider.Root) {
			continue
		}

		if _, err := os.Lstat(path); err != nil {
			continue
		}

		out = append(out, Candidate{
			Provider: provider.Name,
			Category: category,
			Path:     path,
			Label:    label(provider.Root, path),
		})
	}

	return out
}

// orphans matches leftover files in the provider root against the configured
// rules, keeping only those whose owning process is gone.
func (s Scanner) orphans(provider config.Resolved) []Candidate {
	if len(provider.Orphans) == 0 {
		return nil
	}

	entries, err := os.ReadDir(provider.Root)

	if err != nil {
		s.report(provider.Root, err)

		return nil
	}

	out := make([]Candidate, 0, 16)

	for _, entry := range entries {
		for _, rule := range provider.Orphans {
			pattern := rule.Pattern()

			if pattern == nil || !pattern.MatchString(entry.Name()) {
				continue
			}

			if rule.Liveness == config.LivenessPID {
				pid, ok := rule.PID(entry.Name())

				if !ok || proc.Alive(pid) {
					continue
				}
			}

			path := filepath.Join(provider.Root, entry.Name())

			out = append(out, Candidate{
				Provider: provider.Name,
				Category: config.CategoryOrphan,
				Path:     path,
				Label:    label(provider.Root, path),
			})

			break
		}
	}

	return out
}

// inspect gathers git facts, reading each parent repository's registry once
// rather than once per worktree.
func (s Scanner) inspect(ctx context.Context, candidates []Candidate) map[string][]gitwork.Entry {
	registries := make(map[string]gitwork.Registry)
	prunable := make(map[string][]gitwork.Entry)

	for index := range candidates {
		if ctx.Err() != nil {
			break
		}

		candidate := &candidates[index]

		if candidate.Category != config.CategoryWorktree {
			continue
		}

		tree := s.Inspector.Inspect(ctx, candidate.Path)

		candidate.Worktree = &tree
		candidate.Repo = tree.MainRepo

		if tree.MainRepo == "" {
			continue
		}

		registry, ok := registries[tree.MainRepo]

		if !ok {
			loaded, err := gitwork.LoadRegistry(ctx, s.Runner, tree.MainRepo)

			if err != nil {
				s.report(tree.MainRepo, err)

				registries[tree.MainRepo] = gitwork.Registry{}

				continue
			}

			registry = loaded
			registries[tree.MainRepo] = loaded

			if stale := loaded.Prunable(); len(stale) > 0 {
				prunable[tree.MainRepo] = stale
			}
		}

		if entry, found := registry.Lookup(candidate.Path); found {
			tree.Apply(entry)
			candidate.Worktree = &tree
		}
	}

	return prunable
}

// measure sizes every candidate in a single pass, marking build artifacts as it
// goes so the per-artifact breakdown comes free with the totals.
func (s Scanner) measure(ctx context.Context, candidates []Candidate) ([]Candidate, error) {
	if len(candidates) == 0 {
		return candidates, nil
	}

	roots := make([]string, 0, len(candidates))

	for _, candidate := range candidates {
		roots = append(roots, candidate.Path)
	}

	artifacts := s.Config.Artifacts

	walker := diskwalk.Walker{
		Apparent:       s.Config.Defaults.ApparentSize,
		FollowSymlinks: s.Config.Defaults.FollowSymlinks,
		Dedupe:         true,
		Errors:         s.Errors,
		Pruner: diskwalk.PrunerFunc(func(_ string, entry fs.DirEntry, depth int) diskwalk.Decision {
			if depth <= artifacts.MaxDepth && artifacts.IsArtifact(entry.Name()) {
				return diskwalk.DecisionMark
			}

			return diskwalk.DecisionDescend
		}),
	}

	results, err := s.Pool.Run(ctx, walker, roots)

	for index := range results {
		if index >= len(candidates) {
			break
		}

		candidates[index].Bytes = results[index].Bytes
		candidates[index].Files = results[index].Files
		candidates[index].Newest = results[index].Newest
		candidates[index].Leaves = results[index].Leaves
		candidates[index].Key = keyOf(candidates[index].Path)
	}

	return candidates, err
}

// expand turns the build artifacts found inside worktrees into candidates of
// their own. This is what makes it possible to reclaim tens of gigabytes from a
// worktree treex has quite correctly refused to delete.
func (s Scanner) expand(candidates []Candidate) []Candidate {
	if !s.wants(config.CategoryArtifact) {
		return candidates
	}

	out := make([]Candidate, 0, len(candidates)*2)

	for _, candidate := range candidates {
		out = append(out, candidate)

		if candidate.Category != config.CategoryWorktree {
			continue
		}

		for _, leaf := range candidate.Leaves {
			out = append(out, Candidate{
				Provider: candidate.Provider,
				Category: config.CategoryArtifact,
				Path:     leaf.Path,
				Label:    leaf.Path,
				Bytes:    leaf.Bytes,
				Files:    leaf.Files,
				Newest:   candidate.Newest,
				Repo:     candidate.Repo,
				// Build output is regenerable by definition, so it carries no
				// git risk even when the worktree around it is blocked.
				Risk: gitwork.RiskSafe,
				Key:  keyOf(leaf.Path),
			})
		}
	}

	return out
}

func (s Scanner) judge(candidates []Candidate) {
	for index := range candidates {
		candidate := &candidates[index]

		if candidate.Category != config.CategoryWorktree {
			if candidate.Risk == "" {
				candidate.Risk = s.categoryRisk(*candidate)
			}

			continue
		}

		if candidate.Worktree == nil {
			candidate.Risk = gitwork.RiskBlocked
			candidate.Reasons = []string{"git state could not be determined"}

			continue
		}

		candidate.Risk, candidate.Reasons = s.Policy.Verdict(*candidate.Worktree)
	}
}

// categoryRisk applies the structural guards to the non-git categories, which
// have no commits to lose but are still subject to protected paths.
func (s Scanner) categoryRisk(candidate Candidate) gitwork.Risk {
	risk, _ := s.Policy.Verdict(gitwork.Worktree{Path: candidate.Path, Kind: gitwork.KindPlain})

	return risk
}

// filter drops candidates the user asked not to see: wrong category, too
// small, or too recently touched.
func (s Scanner) filter(candidates []Candidate) []Candidate {
	now := time.Now()

	if s.Now != nil {
		now = s.Now()
	}

	minimum := s.Config.Defaults.MinSize.Bytes()
	stale := s.Config.Defaults.OlderThan.Std()

	out := make([]Candidate, 0, len(candidates))

	for _, candidate := range candidates {
		if !s.wants(candidate.Category) {
			continue
		}

		if minimum > 0 && candidate.Bytes < minimum {
			continue
		}

		// A tree touched minutes ago is live work whatever its git state says,
		// so staleness is measured from the newest mtime anywhere inside it.
		if stale > 0 && !candidate.Newest.IsZero() && now.Sub(candidate.Newest) < stale {
			continue
		}

		out = append(out, candidate)
	}

	return out
}

func (s Scanner) wants(category config.Category) bool {
	return slices.Contains(s.Categories, category)
}

func (s Scanner) report(path string, err error) {
	if s.Errors == nil {
		return
	}

	s.Errors(path, err)
}

func label(root, path string) string {
	relative, err := filepath.Rel(root, path)

	if err != nil {
		return path
	}

	return filepath.Join(filepath.Base(root), relative)
}

func dirExists(path string) bool {
	info, err := os.Stat(path)

	return err == nil && info.IsDir()
}

func keyOf(path string) diskwalk.FileKey {
	info, err := os.Lstat(path)

	if err != nil {
		return diskwalk.FileKey{}
	}

	key, _, _ := diskwalk.FileKeyOf(info)

	return key
}

func isCancelled(err error) bool {
	return errors.Is(err, diskwalk.ErrCancelled)
}

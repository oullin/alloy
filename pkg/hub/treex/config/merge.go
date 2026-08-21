package config

import "slices"

// Merge layers a parsed file onto the built-in defaults and returns the result.
//
// The rule is patch, not replace: a provider in the file whose name matches a
// built-in updates only the fields it actually sets, so a user who wants to add
// one cache directory does not have to restate the provider's roots, sessions,
// and orphan rules. Setting Override on a provider opts out of that and
// replaces the built-in wholesale. Names with no built-in counterpart are
// appended in file order.
//
// The zero value is genuinely absent here rather than meaningfully false, with
// one exception: Enabled must be settable to false, which is why the file's
// raw key set is threaded through as seen.
func Merge(base Config, file Config, seen KeySet) Config {
	out := base

	if file.Version != 0 {
		out.Version = file.Version
	}

	out.Defaults = mergeDefaults(base.Defaults, file.Defaults, seen)
	out.Safety = mergeSafety(base.Safety, file.Safety, seen)

	if len(file.Artifacts.Names) > 0 {
		out.Artifacts.Names = file.Artifacts.Names
	}

	if file.Artifacts.MaxDepth != 0 {
		out.Artifacts.MaxDepth = file.Artifacts.MaxDepth
	}

	out.Providers = mergeProviders(base.Providers, file.Providers, seen)

	return out
}

func mergeDefaults(base Defaults, file Defaults, seen KeySet) Defaults {
	out := base

	if file.Jobs != 0 {
		out.Jobs = file.Jobs
	}

	if file.MaxDepth != 0 {
		out.MaxDepth = file.MaxDepth
	}

	if file.OlderThan != 0 {
		out.OlderThan = file.OlderThan
	}

	if file.MinSize != 0 {
		out.MinSize = file.MinSize
	}

	if seen.Has("defaults.apparent-size") {
		out.ApparentSize = file.ApparentSize
	}

	if seen.Has("defaults.follow-symlinks") {
		out.FollowSymlinks = file.FollowSymlinks
	}

	return out
}

func mergeSafety(base Safety, file Safety, seen KeySet) Safety {
	out := base

	// These three default to true, so an absent key and an explicit false are
	// the same Go value. Only the key set can tell them apart, and getting it
	// wrong would silently disable a safety gate.
	if seen.Has("safety.require-clean-worktree") {
		out.RequireCleanWorktree = file.RequireCleanWorktree
	}

	if seen.Has("safety.require-pushed-commits") {
		out.RequirePushedCommits = file.RequirePushedCommits
	}

	if seen.Has("safety.require-no-stash") {
		out.RequireNoStash = file.RequireNoStash
	}

	if file.StatusTimeout != 0 {
		out.StatusTimeout = file.StatusTimeout
	}

	if len(file.ProtectedBranches) > 0 {
		out.ProtectedBranches = file.ProtectedBranches
	}

	// Protected paths are additive. Replacing them wholesale would let a file
	// that names one extra directory quietly drop the built-in guards on
	// ~/Sites and ~/Documents.
	for _, path := range file.ProtectPaths {
		if !slices.Contains(out.ProtectPaths, path) {
			out.ProtectPaths = append(out.ProtectPaths, path)
		}
	}

	return out
}

func mergeProviders(base []Provider, file []Provider, seen KeySet) []Provider {
	out := slices.Clone(base)

	for index, provider := range file {
		position := slices.IndexFunc(out, func(candidate Provider) bool {
			return candidate.Name == provider.Name
		})

		if position < 0 {
			// An unknown name is a new provider, and a user who writes one out
			// in full means to use it.
			if !seen.Has(providerKey(index, "enabled")) {
				provider.Enabled = true
			}

			if provider.Kind == "" {
				provider.Kind = KindAgent
			}

			out = append(out, provider)

			continue
		}

		if provider.Override {
			provider.Kind = defaulted(provider.Kind, out[position].Kind)
			out[position] = provider

			continue
		}

		out[position] = patchProvider(out[position], provider, index, seen)
	}

	return out
}

func patchProvider(base Provider, file Provider, index int, seen KeySet) Provider {
	out := base

	if seen.Has(providerKey(index, "enabled")) {
		out.Enabled = file.Enabled
	}

	if file.Root != "" {
		out.Root = file.Root
	}

	if file.Kind != "" {
		out.Kind = file.Kind
	}

	if len(file.Worktrees) > 0 {
		out.Worktrees = file.Worktrees
	}

	if len(file.Caches) > 0 {
		out.Caches = file.Caches
	}

	if len(file.Sessions) > 0 {
		out.Sessions = file.Sessions
	}

	if len(file.Orphans) > 0 {
		out.Orphans = file.Orphans
	}

	return out
}

func defaulted(value Kind, fallback Kind) Kind {
	if value == "" {
		return fallback
	}

	return value
}

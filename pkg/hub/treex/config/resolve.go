package config

import (
	"path/filepath"
	"strings"
)

// Resolved is a provider with every path expanded to an absolute, cleaned
// location. Nothing downstream of Resolve ever handles a "~" or a relative
// path, which is what lets the sweep guard compare paths by simple prefix.
type Resolved struct {
	Provider

	// Root is the provider root, absolute and cleaned.
	Root string

	// Jail is the directory destructive operations are confined to. It equals
	// Root normally; under a mounted --root it is the mount-relative root, so a
	// bug cannot escape the mount even if a stale plan names a host path.
	Jail string
}

// Resolve expands a provider's root against home and an optional mount prefix.
//
// The mount is how the Docker image works: the container sees the host home at
// /host, so a provider rooted at ~/.claude must be swept at /host/.claude while
// still being reported under its familiar name.
func (p Provider) Resolve(home, mount string) Resolved {
	root := Expand(p.Root, home)

	if mount != "" {
		root = filepath.Join(mount, strings.TrimPrefix(root, string(filepath.Separator)))
	}

	return Resolved{Provider: p, Root: filepath.Clean(root), Jail: filepath.Clean(root)}
}

// ResolveAll expands every provider in a selection.
func ResolveAll(providers []Provider, home, mount string) []Resolved {
	out := make([]Resolved, 0, len(providers))

	for _, provider := range providers {
		out = append(out, provider.Resolve(home, mount))
	}

	return out
}

// ProtectedPaths expands the configured protect-paths against home and an
// optional mount. A path here is never deleted and never descended into.
func (s Safety) ProtectedPaths(home, mount string) []string {
	out := make([]string, 0, len(s.ProtectPaths))

	for _, path := range s.ProtectPaths {
		expanded := Expand(path, home)

		if mount != "" {
			expanded = filepath.Join(mount, strings.TrimPrefix(expanded, string(filepath.Separator)))
		}

		out = append(out, filepath.Clean(expanded))
	}

	return out
}

// Expand resolves a leading ~ against home and cleans the result. Only a
// leading tilde is a home reference, so a directory whose name merely contains
// one survives intact.
func Expand(path, home string) string {
	trimmed := strings.TrimSpace(path)

	if trimmed == "~" {
		return filepath.Clean(home)
	}

	if strings.HasPrefix(trimmed, "~/") {
		return filepath.Clean(filepath.Join(home, trimmed[2:]))
	}

	return filepath.Clean(trimmed)
}

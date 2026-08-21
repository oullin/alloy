package gitwork

import (
	"os"
	"path/filepath"
	"strings"
)

// Kind is how a directory relates to the repository it belongs to.
type Kind string

// Classification is the result of the cheap, subprocess-free inspection.
type Classification struct {
	Kind Kind

	// AdminDir is the parent repository's per-worktree administrative
	// directory, set for linked and broken worktrees.
	AdminDir string

	// Reason explains a broken classification.
	Reason string
}

const (
	// KindLinked is a worktree created by git worktree add. Its .git is a file
	// pointing into the parent repository, which holds a registry entry for it.
	// Removing the directory without telling git leaves that entry dangling.
	KindLinked Kind = "linked"

	// KindClone is a self-contained repository with a real .git directory.
	KindClone Kind = "clone"

	// KindBroken is a linked worktree whose pointer no longer resolves. The
	// registry entry is already gone or already stale, so nothing can be
	// corrupted by removing what is left.
	KindBroken Kind = "broken"

	// KindPlain is a directory with no git state.
	KindPlain Kind = "plain"
)

// Classify determines how path relates to git using a single lstat and, for a
// linked worktree, two small file reads. No git process is started, which is
// what makes it affordable to run over several hundred candidates.
func Classify(path string) Classification {
	marker := filepath.Join(path, ".git")

	info, err := os.Lstat(marker)

	if err != nil {
		return Classification{Kind: KindPlain}
	}

	if info.IsDir() {
		return Classification{Kind: KindClone}
	}

	if !info.Mode().IsRegular() {
		return Classification{Kind: KindBroken, Reason: "'.git' is neither a file nor a directory"}
	}

	contents, err := os.ReadFile(marker)

	if err != nil {
		return Classification{Kind: KindBroken, Reason: "'.git' could not be read"}
	}

	admin := parseGitdir(string(contents))

	if admin == "" {
		return Classification{Kind: KindBroken, Reason: "'.git' has no gitdir pointer"}
	}

	if !filepath.IsAbs(admin) {
		admin = filepath.Join(path, admin)
	}

	admin = filepath.Clean(admin)

	if _, err := os.Stat(admin); err != nil {
		// The parent repository has already forgotten this worktree, or was
		// itself deleted. There is no registry left to damage.
		return Classification{
			Kind:     KindBroken,
			AdminDir: admin,
			Reason:   "the parent repository's worktree entry is gone",
		}
	}

	// The pointer must round-trip. A gitdir file naming a different working
	// tree means two directories disagree about who owns this administrative
	// entry, and removing either could damage the other.
	//
	// Both sides are resolved before comparing because git always writes the
	// fully-resolved path while the caller may well be holding a symlinked one:
	// on macOS every path under /tmp or /var reaches the same directory by two
	// different names, and a textual comparison would call every worktree there
	// broken.
	if back, err := os.ReadFile(filepath.Join(admin, "gitdir")); err == nil {
		pointed := strings.TrimSpace(string(back))

		if !samePath(pointed, marker) && !samePath(filepath.Dir(pointed), path) {
			return Classification{
				Kind:     KindBroken,
				AdminDir: admin,
				Reason:   "the gitdir pointer names a different working tree",
			}
		}
	}

	return Classification{Kind: KindLinked, AdminDir: admin}
}

func parseGitdir(contents string) string {
	for line := range strings.SplitSeq(strings.TrimSpace(contents), "\n") {
		trimmed := strings.TrimSpace(line)

		if after, found := strings.CutPrefix(trimmed, "gitdir:"); found {
			return strings.TrimSpace(after)
		}
	}

	return ""
}

// samePath reports whether two paths reach the same location, resolving
// symlinks where it can and falling back to a textual comparison where a path
// no longer exists.
func samePath(left, right string) bool {
	if filepath.Clean(left) == filepath.Clean(right) {
		return true
	}

	return resolvePath(left) == resolvePath(right)
}

func resolvePath(path string) string {
	cleaned := filepath.Clean(path)

	resolved, err := filepath.EvalSymlinks(cleaned)

	if err != nil {
		// A missing leaf is normal here: the ".git" entry of a worktree that
		// has already been deleted cannot be resolved, but its parent can.
		parent, err := filepath.EvalSymlinks(filepath.Dir(cleaned))

		if err != nil {
			return cleaned
		}

		return filepath.Join(parent, filepath.Base(cleaned))
	}

	return resolved
}

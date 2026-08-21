package sweep

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"hara.sh/alloy/treex/internal/diskwalk"
	"hara.sh/alloy/treex/internal/plan"
)

// Guard re-checks a path immediately before it is deleted.
//
// It exists because a plan is a snapshot: minutes can pass between measuring a
// tree and acting on it, during which a directory can be replaced by a symlink,
// a provider can be reconfigured, or the user can start working in a worktree
// that was idle. Re-deriving the answer costs two syscalls and removes a whole
// class of failure.
type Guard struct {
	// Roots are the currently enabled provider roots, recomputed rather than
	// carried over from the scan.
	Roots []string

	// Protected are paths that are never touched, nor descended into.
	Protected []string
}

// minSegments is the shallowest path treex will ever unlink. Three segments
// rules out "/", "/Users", and "/Users/someone" while still permitting the
// provider roots, which sit at four.
const minSegments = 3

// Check validates one action. A nil return is the only thing that authorises a
// delete.
func (g Guard) Check(action plan.Action) error {
	path := action.Path

	if !filepath.IsAbs(path) || path != filepath.Clean(path) {
		return fmt.Errorf("%w: %s is not an absolute, cleaned path", ErrGuardRejected, path)
	}

	if segments(path) < minSegments {
		return fmt.Errorf("%w: %s is too close to the filesystem root", ErrGuardRejected, path)
	}

	if guard := g.protects(path); guard != "" {
		return fmt.Errorf("%w: %s is inside %s", ErrProtected, path, guard)
	}

	if !g.inRoot(path) {
		return fmt.Errorf("%w: %s", ErrOutsideRoot, path)
	}

	info, err := os.Lstat(path)

	if err != nil {
		return fmt.Errorf("stat %s: %w", path, err)
	}

	// Never delete through a symlink: the link is cheap to remove, but
	// following one would take the target with it.
	if info.Mode()&os.ModeSymlink != 0 {
		return fmt.Errorf("%w: %s is a symlink", ErrGuardRejected, path)
	}

	if action.Kind != plan.KindRemoveFile && !info.IsDir() {
		return fmt.Errorf("%w: %s", ErrNotADirectory, path)
	}

	return g.identical(path, info, action.Key)
}

// identical confirms the path still refers to the same inode the scan
// measured, catching a directory that was swapped out in between.
//
// It is a cheap best-effort check rather than a guarantee. Filesystems differ
// in whether they reuse a freed inode number immediately — ext4 does, APFS does
// not — so a directory deleted and recreated between the scan and the sweep can
// still present the same identity. The real protections are the ones above it:
// the path must be absolute, deep enough, inside a provider root, outside every
// protected path, and not a symlink.
func (g Guard) identical(path string, info os.FileInfo, want diskwalk.FileKey) error {
	if want.Zero() {
		return nil
	}

	got, _, ok := diskwalk.FileKeyOf(info)

	if !ok || got.Zero() {
		return nil
	}

	if got != want {
		return fmt.Errorf("%w: %s", ErrChanged, path)
	}

	return nil
}

func (g Guard) inRoot(path string) bool {
	for _, root := range g.Roots {
		cleaned := filepath.Clean(root)

		if path == cleaned || strings.HasPrefix(path, cleaned+string(filepath.Separator)) {
			return true
		}
	}

	return false
}

func (g Guard) protects(path string) string {
	for _, protected := range g.Protected {
		cleaned := filepath.Clean(protected)

		if path == cleaned || strings.HasPrefix(path, cleaned+string(filepath.Separator)) {
			return cleaned
		}
	}

	return ""
}

func segments(path string) int {
	trimmed := strings.Trim(filepath.ToSlash(path), "/")

	if trimmed == "" {
		return 0
	}

	return len(strings.Split(trimmed, "/"))
}

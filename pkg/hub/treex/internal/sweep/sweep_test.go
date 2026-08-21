package sweep_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"hara.sh/alloy/treex/internal/diskwalk"
	"hara.sh/alloy/treex/internal/gitwork"
	"hara.sh/alloy/treex/internal/plan"
	"hara.sh/alloy/treex/internal/sweep"
)

type repo struct {
	t   *testing.T
	dir string
}

// The test this whole package exists to pass. Removing a linked worktree by
// unlinking its directory leaves the parent repository holding a registry entry
// that points at nothing, which is the damage treex was written to clean up.
// Doing that itself would be the worst possible bug.
func TestSweeperLeavesNoDanglingRegistryEntries(t *testing.T) {
	t.Parallel()

	origin := newRepo(t)

	first := origin.addWorktree("one")

	origin.addWorktree("two")
	origin.addWorktree("three")

	if got := countPrunable(t, origin.dir); got != 0 {
		t.Fatalf("prunable before sweep = %d, want 0", got)
	}

	sweeper := sweep.Sweeper{
		Guard: sweep.Guard{Roots: []string{filepath.Dir(first)}},
		Force: true,
	}

	result := sweeper.Run(t.Context(), plan.Plan{
		Actions: []plan.Action{
			removeWorktree(t, first, origin.dir),
			{Kind: plan.KindPruneRegistry, Path: origin.dir, Repo: origin.dir},
		},
	})

	if result.Failed != 0 {
		t.Fatalf("failed = %d, want 0: %v", result.Failed, firstError(result))
	}

	if _, err := os.Stat(first); !os.IsNotExist(err) {
		t.Fatalf("worktree still present at %s", first)
	}

	if got := countPrunable(t, origin.dir); got != 0 {
		t.Fatalf("prunable after sweep = %d, want 0 (treex must not create the corruption it cleans)", got)
	}

	// The other two must survive untouched.
	if got, want := countWorktrees(t, origin.dir), 3; got != want {
		t.Fatalf("worktrees = %d, want %d", got, want)
	}
}

func TestSweeperRemovesACloneOutright(t *testing.T) {
	t.Parallel()

	origin := newRepo(t)
	clone := origin.cloneTo(t)

	sweeper := sweep.Sweeper{Guard: sweep.Guard{Roots: []string{filepath.Dir(clone)}}}

	result := sweeper.Run(t.Context(), plan.Plan{
		Actions: []plan.Action{{
			Kind:     plan.KindRemoveWorktree,
			Path:     clone,
			Worktree: &gitwork.Worktree{Path: clone, Kind: gitwork.KindClone},
			Key:      keyOf(t, clone),
		}},
	})

	if result.Failed != 0 {
		t.Fatalf("failed = %d, want 0: %v", result.Failed, firstError(result))
	}

	if _, err := os.Stat(clone); !os.IsNotExist(err) {
		t.Fatal("clone still present")
	}
}

func TestSweeperDryRunTouchesNothing(t *testing.T) {
	t.Parallel()

	origin := newRepo(t)
	worktree := origin.addWorktree("kept")

	sweeper := sweep.Sweeper{
		Guard:  sweep.Guard{Roots: []string{filepath.Dir(worktree)}},
		DryRun: true,
		Force:  true,
	}

	result := sweeper.Run(t.Context(), plan.Plan{
		Actions: []plan.Action{removeWorktree(t, worktree, origin.dir)},
	})

	if result.Failed != 0 {
		t.Fatalf("failed = %d, want 0", result.Failed)
	}

	if _, err := os.Stat(worktree); err != nil {
		t.Fatalf("dry run removed %s: %v", worktree, err)
	}
}

func TestSweeperPrunesAnAlreadyDanglingEntry(t *testing.T) {
	t.Parallel()

	origin := newRepo(t)
	abandoned := origin.addWorktree("abandoned")

	// Simulate the damage an earlier manual delete left behind.
	if err := os.RemoveAll(abandoned); err != nil {
		t.Fatalf("remove: %v", err)
	}

	if got := countPrunable(t, origin.dir); got != 1 {
		t.Fatalf("prunable = %d, want 1", got)
	}

	sweeper := sweep.Sweeper{Guard: sweep.Guard{Roots: []string{origin.dir}}}

	result := sweeper.Run(t.Context(), plan.Plan{
		Actions: []plan.Action{{Kind: plan.KindPruneRegistry, Path: origin.dir, Repo: origin.dir}},
	})

	if result.Failed != 0 {
		t.Fatalf("failed = %d, want 0: %v", result.Failed, firstError(result))
	}

	if got := countPrunable(t, origin.dir); got != 0 {
		t.Fatalf("prunable after prune = %d, want 0", got)
	}
}

func TestGuardRefusesAPathOutsideEveryRoot(t *testing.T) {
	t.Parallel()

	outside := t.TempDir()
	guard := sweep.Guard{Roots: []string{t.TempDir()}}

	err := guard.Check(plan.Action{Kind: plan.KindRemoveTree, Path: outside})

	if err == nil {
		t.Fatal("Check err = nil, want a refusal")
	}
}

func TestGuardRefusesAProtectedPath(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "keep")

	if err := os.MkdirAll(target, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	guard := sweep.Guard{Roots: []string{root}, Protected: []string{root}}

	if err := guard.Check(plan.Action{Kind: plan.KindRemoveTree, Path: target}); err == nil {
		t.Fatal("Check err = nil, want the protected path refused")
	}
}

func TestGuardRefusesAShallowPath(t *testing.T) {
	t.Parallel()

	guard := sweep.Guard{Roots: []string{"/"}}

	for _, path := range []string{"/", "/Users"} {
		if err := guard.Check(plan.Action{Kind: plan.KindRemoveTree, Path: path}); err == nil {
			t.Fatalf("Check(%q) err = nil, want a refusal", path)
		}
	}
}

func TestGuardRefusesASymlink(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	real := filepath.Join(root, "real")
	link := filepath.Join(root, "link")

	if err := os.MkdirAll(real, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.Symlink(real, link); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	guard := sweep.Guard{Roots: []string{root}}

	// Following the link would take the target with it.
	if err := guard.Check(plan.Action{Kind: plan.KindRemoveTree, Path: link}); err == nil {
		t.Fatal("Check err = nil, want the symlink refused")
	}
}

func TestGuardRefusesAPathWhoseIdentityChanged(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "tree")
	other := filepath.Join(root, "other")

	for _, dir := range []string{target, other} {
		if err := os.MkdirAll(dir, 0o700); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	// Recreating a directory is not a reliable way to change its inode: ext4
	// hands the freed number straight back, where APFS does not. Comparing
	// against a genuinely different directory tests the guard rather than the
	// allocator.
	guard := sweep.Guard{Roots: []string{root}}

	err := guard.Check(plan.Action{Kind: plan.KindRemoveTree, Path: target, Key: keyOf(t, other)})

	if err == nil {
		t.Fatal("Check err = nil, want the mismatched identity refused")
	}
}

func TestGuardAcceptsAPathThatIsUnchanged(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	target := filepath.Join(root, "tree")

	if err := os.MkdirAll(target, 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	guard := sweep.Guard{Roots: []string{root}}

	if err := guard.Check(plan.Action{Kind: plan.KindRemoveTree, Path: target, Key: keyOf(t, target)}); err != nil {
		t.Fatalf("Check err = %v, want nil", err)
	}
}

func removeWorktree(t *testing.T, path, repo string) plan.Action {
	t.Helper()

	return plan.Action{
		Kind: plan.KindRemoveWorktree,
		Path: path,
		Repo: repo,
		Worktree: &gitwork.Worktree{
			Path:     path,
			Kind:     gitwork.KindLinked,
			MainRepo: repo,
		},
		Key: keyOf(t, path),
	}
}

func keyOf(t *testing.T, path string) diskwalk.FileKey {
	t.Helper()

	info, err := os.Lstat(path)

	if err != nil {
		t.Fatalf("lstat %s: %v", path, err)
	}

	key, _, _ := diskwalk.FileKeyOf(info)

	return key
}

func countPrunable(t *testing.T, repo string) int {
	t.Helper()

	registry, err := gitwork.LoadRegistry(t.Context(), gitwork.Runner{}, repo)

	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}

	return len(registry.Prunable())
}

func countWorktrees(t *testing.T, repo string) int {
	t.Helper()

	registry, err := gitwork.LoadRegistry(t.Context(), gitwork.Runner{}, repo)

	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}

	return len(registry.Entries)
}

func firstError(result sweep.Result) error {
	for _, outcome := range result.Outcomes {
		if outcome.Err != nil {
			return outcome.Err
		}
	}

	return nil
}

func newRepo(t *testing.T) *repo {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}

	r := &repo{t: t, dir: t.TempDir()}

	r.git("init", "--initial-branch=work")
	r.git("config", "user.email", "treex@example.test")
	r.git("config", "user.name", "treex tests")

	if err := os.WriteFile(filepath.Join(r.dir, "README.md"), []byte("hello"), 0o600); err != nil {
		t.Fatalf("write: %v", err)
	}

	r.git("add", "README.md")
	r.git("commit", "-m", "init")

	return r
}

func (r *repo) git(args ...string) string {
	r.t.Helper()

	cmd := exec.Command("git", append([]string{"-C", r.dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0", "HOME="+r.dir)

	out, err := cmd.CombinedOutput()

	if err != nil {
		r.t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}

	return strings.TrimSpace(string(out))
}

func (r *repo) addWorktree(name string) string {
	r.t.Helper()

	path := filepath.Join(r.t.TempDir(), name)

	r.git("worktree", "add", "-b", name, path)

	return path
}

func (r *repo) cloneTo(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "clone")

	cmd := exec.Command("git", "clone", r.dir, path)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git clone: %v\n%s", err, out)
	}

	return path
}

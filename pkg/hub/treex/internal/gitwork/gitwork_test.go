package gitwork_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hara.sh/alloy/treex/internal/gitwork"
)

func TestClassifyRecognisesALinkedWorktree(t *testing.T) {
	t.Parallel()

	origin := newRepo(t)
	worktree := origin.addWorktree("feature")

	classification := gitwork.Classify(worktree)

	if got, want := classification.Kind, gitwork.KindLinked; got != want {
		t.Fatalf("Kind = %q, want %q", got, want)
	}

	if classification.AdminDir == "" {
		t.Fatal("AdminDir is empty, want the parent repository's worktree entry")
	}

	// The distinction the whole package rests on: a linked worktree's .git is a
	// file, not a directory.
	info, err := os.Lstat(filepath.Join(worktree, ".git"))

	if err != nil {
		t.Fatalf("lstat .git: %v", err)
	}

	if info.IsDir() {
		t.Fatal(".git is a directory, want a gitdir pointer file")
	}
}

func TestClassifyRecognisesAClone(t *testing.T) {
	t.Parallel()

	origin := newRepo(t)
	clone := origin.cloneTo(t)

	if got, want := gitwork.Classify(clone).Kind, gitwork.KindClone; got != want {
		t.Fatalf("Kind = %q, want %q", got, want)
	}
}

func TestClassifyRecognisesAPlainDirectory(t *testing.T) {
	t.Parallel()

	if got, want := gitwork.Classify(t.TempDir()).Kind, gitwork.KindPlain; got != want {
		t.Fatalf("Kind = %q, want %q", got, want)
	}
}

func TestClassifyReportsAWorktreeWhoseParentIsGone(t *testing.T) {
	t.Parallel()

	origin := newRepo(t)
	worktree := origin.addWorktree("orphan")

	// Deleting the parent repository is what leaves a worktree pointing at an
	// administrative directory that no longer exists.
	if err := os.RemoveAll(origin.dir); err != nil {
		t.Fatalf("remove parent: %v", err)
	}

	classification := gitwork.Classify(worktree)

	if got, want := classification.Kind, gitwork.KindBroken; got != want {
		t.Fatalf("Kind = %q, want %q", got, want)
	}

	if classification.Reason == "" {
		t.Fatal("Reason is empty, want an explanation")
	}
}

func TestInspectFindsUnpushedCommits(t *testing.T) {
	t.Parallel()

	origin := newRepo(t)
	clone := origin.cloneTo(t)

	write(t, filepath.Join(clone, "local.txt"), "only here")

	run(t, clone, "add", "local.txt")
	run(t, clone, "commit", "-m", "local only")

	tree := inspector().Inspect(t.Context(), clone)

	if got, want := tree.Unpushed, 1; got != want {
		t.Fatalf("Unpushed = %d, want %d", got, want)
	}

	if tree.Dirty != 0 {
		t.Fatalf("Dirty = %d, want 0", tree.Dirty)
	}
}

func TestInspectReportsAPushedCloneAsHavingNothingToLose(t *testing.T) {
	t.Parallel()

	origin := newRepo(t)
	clone := origin.cloneTo(t)

	tree := inspector().Inspect(t.Context(), clone)

	if tree.HasWork() {
		t.Fatalf("HasWork = true, want false (dirty=%d untracked=%d unpushed=%d)", tree.Dirty, tree.Untracked, tree.Unpushed)
	}
}

func TestInspectCountsUncommittedChanges(t *testing.T) {
	t.Parallel()

	origin := newRepo(t)
	clone := origin.cloneTo(t)

	write(t, filepath.Join(clone, "README.md"), "modified")

	tree := inspector().Inspect(t.Context(), clone)

	if got, want := tree.Dirty, 1; got != want {
		t.Fatalf("Dirty = %d, want %d", got, want)
	}
}

// This is the filter that makes treex useful at all. Every agent worktree has an
// untracked node_modules; if that counts as uncommitted work then every single
// candidate is blocked and the tool reclaims nothing.
func TestInspectIgnoresUntrackedBuildArtifacts(t *testing.T) {
	t.Parallel()

	origin := newRepo(t)
	clone := origin.cloneTo(t)

	write(t, filepath.Join(clone, "node_modules", "left-pad", "index.js"), "module.exports = 1")
	write(t, filepath.Join(clone, "dist", "bundle.js"), "console.log(1)")

	tree := inspector().Inspect(t.Context(), clone)

	if got, want := tree.Untracked, 0; got != want {
		t.Fatalf("Untracked = %d, want %d (build artifacts must not read as work)", got, want)
	}

	if tree.HasWork() {
		t.Fatal("HasWork = true, want false for a tree dirty only with build output")
	}
}

func TestInspectStillCountsRealUntrackedFiles(t *testing.T) {
	t.Parallel()

	origin := newRepo(t)
	clone := origin.cloneTo(t)

	write(t, filepath.Join(clone, "node_modules", "pkg", "index.js"), "1")
	write(t, filepath.Join(clone, "notes.md"), "a real untracked file")

	tree := inspector().Inspect(t.Context(), clone)

	if got, want := tree.Untracked, 1; got != want {
		t.Fatalf("Untracked = %d, want %d", got, want)
	}
}

func TestInspectResolvesTheParentOfALinkedWorktree(t *testing.T) {
	t.Parallel()

	origin := newRepo(t)
	worktree := origin.addWorktree("feature")

	tree := inspector().Inspect(t.Context(), worktree)

	if got, want := tree.Kind, gitwork.KindLinked; got != want {
		t.Fatalf("Kind = %q, want %q", got, want)
	}

	if got, want := resolve(t, tree.MainRepo), resolve(t, origin.dir); got != want {
		t.Fatalf("MainRepo = %q, want %q", got, want)
	}

	if got, want := tree.Branch, "feature"; got != want {
		t.Fatalf("Branch = %q, want %q", got, want)
	}
}

func TestLoadRegistryListsEveryWorktree(t *testing.T) {
	t.Parallel()

	origin := newRepo(t)

	first := origin.addWorktree("one")
	origin.addWorktree("two")

	registry, err := gitwork.LoadRegistry(t.Context(), gitwork.Runner{}, origin.dir)

	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}

	// The main working tree plus the two linked ones.
	if got, want := len(registry.Entries), 3; got != want {
		t.Fatalf("entries = %d, want %d", got, want)
	}

	if _, ok := registry.Lookup(resolve(t, first)); !ok {
		if _, ok := registry.Lookup(first); !ok {
			t.Fatalf("worktree %s missing from the registry", first)
		}
	}
}

// The damage treex exists to clean up: deleting a linked worktree's directory
// without telling git leaves a dangling registry entry behind.
func TestLoadRegistryReportsADeletedWorktreeAsPrunable(t *testing.T) {
	t.Parallel()

	origin := newRepo(t)
	worktree := origin.addWorktree("abandoned")

	if err := os.RemoveAll(worktree); err != nil {
		t.Fatalf("remove worktree: %v", err)
	}

	registry, err := gitwork.LoadRegistry(t.Context(), gitwork.Runner{}, origin.dir)

	if err != nil {
		t.Fatalf("LoadRegistry: %v", err)
	}

	prunable := registry.Prunable()

	if got, want := len(prunable), 1; got != want {
		t.Fatalf("prunable = %d, want %d", got, want)
	}

	if prunable[0].PrunableReason == "" {
		t.Fatal("PrunableReason is empty, want git's explanation")
	}
}

func TestRunnerReportsAFailureRatherThanHanging(t *testing.T) {
	t.Parallel()

	_, err := gitwork.Runner{}.Run(t.Context(), t.TempDir(), "rev-parse", "--git-common-dir")

	if err == nil {
		t.Fatal("Run err = nil, want a failure outside a repository")
	}
}

func inspector() gitwork.Inspector {
	return gitwork.Inspector{
		ArtifactRoot: func(relative string) bool {
			segment, _, _ := strings.Cut(relative, "/")

			switch segment {
			case "node_modules", "dist", "build", ".turbo":
				return true
			default:
				return false
			}
		},
	}
}

func resolve(t *testing.T, path string) string {
	t.Helper()

	resolved, err := filepath.EvalSymlinks(path)

	if err != nil {
		return filepath.Clean(path)
	}

	return resolved
}

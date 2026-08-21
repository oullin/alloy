package diskwalk_test

import (
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"
	"time"

	"hara.sh/alloy/treex/internal/diskwalk"
)

func TestRunMeasuresApparentSizeOfAFlatTree(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	writeFile(t, filepath.Join(root, "a.txt"), 100)
	writeFile(t, filepath.Join(root, "b.txt"), 200)

	makeDir(t, filepath.Join(root, "nested"))
	writeFile(t, filepath.Join(root, "nested", "c.txt"), 300)

	results := run(t, diskwalk.Walker{Apparent: true}, root)

	if got, want := results[0].Bytes, int64(600); got != want {
		t.Fatalf("Bytes = %d, want %d", got, want)
	}

	if got, want := results[0].Files, int64(3); got != want {
		t.Fatalf("Files = %d, want %d", got, want)
	}

	if got, want := results[0].Dirs, int64(1); got != want {
		t.Fatalf("Dirs = %d, want %d", got, want)
	}
}

func TestRunReportsAllocatedBlocksByDefault(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	writeFile(t, filepath.Join(root, "tiny.txt"), 1)

	results := run(t, diskwalk.Walker{}, root)

	// A one-byte file still occupies at least one allocation unit. The exact
	// figure is filesystem-dependent, so assert the property that matters:
	// block accounting never under-reports what deleting the file will free.
	if results[0].Bytes < 1 {
		t.Fatalf("Bytes = %d, want at least the apparent size", results[0].Bytes)
	}

	apparent := run(t, diskwalk.Walker{Apparent: true}, root)

	if results[0].Bytes < apparent[0].Bytes {
		t.Fatalf("block size %d is below apparent size %d", results[0].Bytes, apparent[0].Bytes)
	}
}

func TestRunAttributesMarkedSubtreesToLeaves(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	writeFile(t, filepath.Join(root, "src.ts"), 50)

	makeDir(t, filepath.Join(root, "node_modules", "pkg"))
	writeFile(t, filepath.Join(root, "node_modules", "pkg", "index.js"), 400)

	results := run(t, diskwalk.Walker{Apparent: true, Pruner: markNamed("node_modules")}, root)

	// The root total includes the leaf: a worktree's size is what it occupies,
	// artifacts and all.
	if got, want := results[0].Bytes, int64(450); got != want {
		t.Fatalf("Bytes = %d, want %d", got, want)
	}

	if got, want := len(results[0].Leaves), 1; got != want {
		t.Fatalf("Leaves = %d, want %d", got, want)
	}

	leaf := results[0].Leaves[0]

	if got, want := leaf.Bytes, int64(400); got != want {
		t.Fatalf("leaf Bytes = %d, want %d", got, want)
	}

	if got, want := leaf.Name, "node_modules"; got != want {
		t.Fatalf("leaf Name = %q, want %q", got, want)
	}
}

func TestRunRollsNestedMarksIntoTheOuterLeaf(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	inner := filepath.Join(root, "node_modules", "a", "node_modules", "b")

	makeDir(t, inner)
	writeFile(t, filepath.Join(inner, "index.js"), 128)

	results := run(t, diskwalk.Walker{Apparent: true, Pruner: markNamed("node_modules")}, root)

	if got, want := len(results[0].Leaves), 1; got != want {
		t.Fatalf("Leaves = %d, want %d (nested marks must roll up)", got, want)
	}

	if got, want := results[0].Leaves[0].Bytes, int64(128); got != want {
		t.Fatalf("leaf Bytes = %d, want %d", got, want)
	}
}

func TestRunSkipsDirectoriesThePrunerRejects(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	writeFile(t, filepath.Join(root, "keep.txt"), 10)

	makeDir(t, filepath.Join(root, "ignore"))
	writeFile(t, filepath.Join(root, "ignore", "big.bin"), 9000)

	pruner := diskwalk.PrunerFunc(func(_ string, entry fs.DirEntry, _ int) diskwalk.Decision {
		if entry.Name() == "ignore" {
			return diskwalk.DecisionSkip
		}

		return diskwalk.DecisionDescend
	})

	results := run(t, diskwalk.Walker{Apparent: true, Pruner: pruner}, root)

	if got, want := results[0].Bytes, int64(10); got != want {
		t.Fatalf("Bytes = %d, want %d", got, want)
	}
}

func TestRunCountsAHardLinkOnceWhenDeduping(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	original := filepath.Join(root, "original.bin")

	writeFile(t, original, 500)

	if err := os.Link(original, filepath.Join(root, "linked.bin")); err != nil {
		t.Skipf("hard links unavailable: %v", err)
	}

	linked := run(t, diskwalk.Walker{Apparent: true, Dedupe: true}, root)

	if got, want := linked[0].Bytes, int64(500); got != want {
		t.Fatalf("deduped Bytes = %d, want %d", got, want)
	}

	// Without dedupe the same bytes are counted twice, which is what makes the
	// flag worth having.
	doubled := run(t, diskwalk.Walker{Apparent: true}, root)

	if got, want := doubled[0].Bytes, int64(1000); got != want {
		t.Fatalf("undeduped Bytes = %d, want %d", got, want)
	}
}

func TestRunDoesNotFollowSymlinksOutOfTheTree(t *testing.T) {
	t.Parallel()

	outside := t.TempDir()

	writeFile(t, filepath.Join(outside, "secret.bin"), 4096)

	root := t.TempDir()

	writeFile(t, filepath.Join(root, "own.txt"), 32)

	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Skipf("symlinks unavailable: %v", err)
	}

	results := run(t, diskwalk.Walker{Apparent: true}, root)

	if got, want := results[0].Bytes, int64(32); got != want {
		t.Fatalf("Bytes = %d, want %d (the walk must not leave the tree)", got, want)
	}
}

func TestRunSurvivesAnUnreadableDirectory(t *testing.T) {
	t.Parallel()

	if os.Geteuid() == 0 {
		t.Skip("root ignores directory permissions")
	}

	root := t.TempDir()

	writeFile(t, filepath.Join(root, "readable.txt"), 64)

	locked := filepath.Join(root, "locked")

	makeDir(t, locked)
	writeFile(t, filepath.Join(locked, "hidden.txt"), 128)

	if err := os.Chmod(locked, 0o000); err != nil {
		t.Fatalf("chmod: %v", err)
	}

	t.Cleanup(func() {
		_ = os.Chmod(locked, 0o700)
	})

	var failures int

	walker := diskwalk.Walker{
		Apparent: true,
		Errors: func(_ string, _ error) {
			failures++
		},
	}

	results := run(t, walker, root)

	// One unreadable directory must not cost the rest of the report.
	if got, want := results[0].Bytes, int64(64); got != want {
		t.Fatalf("Bytes = %d, want %d", got, want)
	}

	if results[0].Errors == 0 {
		t.Fatal("Errors = 0, want the unreadable directory to be counted")
	}
}

func TestRunTracksTheNewestModificationTime(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	recent := filepath.Join(root, "recent.txt")

	writeFile(t, filepath.Join(root, "old.txt"), 10)
	writeFile(t, recent, 10)

	stamp := time.Now().Add(-time.Hour)

	if err := os.Chtimes(filepath.Join(root, "old.txt"), stamp, stamp); err != nil {
		t.Fatalf("chtimes: %v", err)
	}

	results := run(t, diskwalk.Walker{Apparent: true}, root)

	if !results[0].Newest.After(stamp) {
		t.Fatalf("Newest = %v, want something after %v", results[0].Newest, stamp)
	}
}

func TestRunMeasuresEveryRootInOrder(t *testing.T) {
	t.Parallel()

	first := t.TempDir()
	second := t.TempDir()

	writeFile(t, filepath.Join(first, "a.bin"), 111)
	writeFile(t, filepath.Join(second, "b.bin"), 222)

	results := run(t, diskwalk.Walker{Apparent: true}, first, second)

	if got, want := len(results), 2; got != want {
		t.Fatalf("results = %d, want %d", got, want)
	}

	if got, want := results[0].Bytes, int64(111); got != want {
		t.Fatalf("first Bytes = %d, want %d", got, want)
	}

	if got, want := results[1].Bytes, int64(222); got != want {
		t.Fatalf("second Bytes = %d, want %d", got, want)
	}
}

func TestRunReturnsPromptlyWhenCancelled(t *testing.T) {
	t.Parallel()

	root := t.TempDir()

	// Wide and deep enough that a walk which ignored cancellation would keep
	// going well past the deadline asserted below.
	for i := range 40 {
		branch := filepath.Join(root, "branch", string(rune('a'+i%26)), string(rune('a'+i)))

		makeDir(t, branch)

		for f := range 40 {
			writeFile(t, filepath.Join(branch, string(rune('a'+f%26))+".bin"), 64)
		}
	}

	ctx, cancel := context.WithCancel(t.Context())

	cancel()

	started := time.Now()

	_, err := diskwalk.NewPool(4).Run(ctx, diskwalk.Walker{}, []string{root})

	if err == nil {
		t.Fatal("Run err = nil, want ErrCancelled")
	}

	if elapsed := time.Since(started); elapsed > 2*time.Second {
		t.Fatalf("cancelled Run took %v, want it to unwind promptly", elapsed)
	}
}

func TestRunWithNoRootsIsANoOp(t *testing.T) {
	t.Parallel()

	results, err := diskwalk.NewPool(2).Run(t.Context(), diskwalk.Walker{}, nil)

	if err != nil {
		t.Fatalf("Run err = %v, want nil", err)
	}

	if len(results) != 0 {
		t.Fatalf("results = %d, want 0", len(results))
	}
}

func TestNewPoolClampsTheWorkerCount(t *testing.T) {
	t.Parallel()

	if got := diskwalk.NewPool(1_000).Jobs(); got > 64 {
		t.Fatalf("Jobs = %d, want it capped at 64", got)
	}

	if got := diskwalk.NewPool(0).Jobs(); got <= 0 {
		t.Fatalf("Jobs = %d, want a derived positive count", got)
	}
}

func run(t *testing.T, walker diskwalk.Walker, roots ...string) []diskwalk.Result {
	t.Helper()

	results, err := diskwalk.NewPool(4).Run(t.Context(), walker, roots)

	if err != nil {
		t.Fatalf("Run: %v", err)
	}

	return results
}

func markNamed(name string) diskwalk.Pruner {
	return diskwalk.PrunerFunc(func(_ string, entry fs.DirEntry, _ int) diskwalk.Decision {
		if entry.Name() == name {
			return diskwalk.DecisionMark
		}

		return diskwalk.DecisionDescend
	})
}

func writeFile(t *testing.T, path string, size int) {
	t.Helper()

	if err := os.WriteFile(path, make([]byte, size), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func makeDir(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o700); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

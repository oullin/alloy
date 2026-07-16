package filesystem_test

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oullin/alloy/pkg/hub/filesystem"
)

// rootedFixture builds a root directory with a secret sitting outside it, so
// every escape test has something real to try to reach.
func rootedFixture(t *testing.T) (root string, outside string) {
	t.Helper()

	base := t.TempDir()
	outside = filepath.Join(base, "secret.txt")
	root = filepath.Join(base, "root")

	writeFile(t, outside, "top secret")
	makeDir(t, root)

	return root, outside
}

func openRooted(t *testing.T, root string) *filesystem.Rooted {
	t.Helper()

	r, err := filesystem.At(root)

	if err != nil {
		t.Fatal(err)
	}

	t.Cleanup(func() {
		_ = r.Close()
	})

	return r
}

// TestRootedRefusesEscape is the reason this type exists. Each name here is a
// real way out of a directory that a naive path check misses.
func TestRootedRefusesEscape(t *testing.T) {
	root, outside := rootedFixture(t)
	writeFile(t, filepath.Join(root, "inside.txt"), "fine")

	r := openRooted(t, root)
	ctx := context.Background()

	names := map[string]string{
		"parent traversal":     filepath.Join("..", "secret.txt"),
		"repeated traversal":   filepath.Join("..", "..", "..", "etc", "hosts"),
		"absolute path":        outside,
		"traversal mid-path":   filepath.Join("sub", "..", "..", "secret.txt"),
		"absolute unix system": "/etc/hosts",
	}

	for label, name := range names {
		t.Run(label, func(t *testing.T) {
			if _, err := r.Get(ctx, name); err == nil {
				t.Fatalf("Get(%q) succeeded; it must not escape the root", name)
			}

			if err := r.Put(ctx, name, []byte("owned")); err == nil {
				t.Fatalf("Put(%q) succeeded; it must not escape the root", name)
			}
		})
	}

	// The escape must not have damaged what it was aiming at.
	content, err := os.ReadFile(outside)

	if err != nil {
		t.Fatal(err)
	}

	if string(content) != "top secret" {
		t.Errorf("a refused write still modified the file outside the root: %q", content)
	}
}

// TestRootedRefusesSymlinkEscape covers the case the consumer's hand-rolled
// guard existed for: a link that lives inside the root but points out of it.
// Lexical containment checks pass this happily.
func TestRootedRefusesSymlinkEscape(t *testing.T) {
	requireSymlinks(t)

	root, outside := rootedFixture(t)

	if err := os.Symlink(outside, filepath.Join(root, "escape")); err != nil {
		t.Fatal(err)
	}

	r := openRooted(t, root)

	if _, err := r.Get(context.Background(), "escape"); err == nil {
		t.Fatal("Get through a symlink pointing outside the root succeeded")
	}
}

// TestRootedFollowsInternalSymlink pins the deliberate other half: Rooted
// prevents escape, it does not ban links.
func TestRootedFollowsInternalSymlink(t *testing.T) {
	requireSymlinks(t)

	root, _ := rootedFixture(t)
	writeFile(t, filepath.Join(root, "target.txt"), "reachable")

	if err := os.Symlink("target.txt", filepath.Join(root, "link.txt")); err != nil {
		t.Fatal(err)
	}

	r := openRooted(t, root)

	got, err := r.Get(context.Background(), "link.txt")

	if err != nil {
		t.Fatalf("Get through a symlink staying inside the root failed: %v", err)
	}

	if string(got) != "reachable" {
		t.Errorf("Get = %q, want %q", got, "reachable")
	}

	// ...and callers who need to reject links as policy can still detect one.
	if !r.IsLink("link.txt") {
		t.Error("IsLink did not report the link, leaving policy callers no way to refuse it")
	}
}

func TestRootedReadWrite(t *testing.T) {
	root, _ := rootedFixture(t)
	r := openRooted(t, root)
	ctx := context.Background()

	t.Run("Put then Get round-trips and creates parents", func(t *testing.T) {
		if err := r.Put(ctx, filepath.Join("nested", "deep", "file.txt"), []byte("payload")); err != nil {
			t.Fatal(err)
		}

		got, err := r.Get(ctx, filepath.Join("nested", "deep", "file.txt"))

		if err != nil {
			t.Fatal(err)
		}

		if string(got) != "payload" {
			t.Errorf("Get = %q, want %q", got, "payload")
		}
	})

	t.Run("PutStream round-trips", func(t *testing.T) {
		if err := r.PutStream(ctx, "streamed.txt", strings.NewReader("via reader")); err != nil {
			t.Fatal(err)
		}

		got, err := r.Get(ctx, "streamed.txt")

		if err != nil {
			t.Fatal(err)
		}

		if string(got) != "via reader" {
			t.Errorf("Get = %q, want %q", got, "via reader")
		}
	})

	// The error model has to match Local's, or the uniformity fixed in the
	// sibling change would not hold for rooted callers.
	t.Run("missing name reports fs.ErrNotExist", func(t *testing.T) {
		_, err := r.Get(ctx, "absent.txt")

		if !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("Get on a missing name: errors.Is(err, fs.ErrNotExist) = false (err = %v)", err)
		}

		if !errors.Is(err, filesystem.ErrNotFound) {
			t.Errorf("Get on a missing name lost the ErrNotFound sentinel (err = %v)", err)
		}
	})

	t.Run("an escape error is not a missing-file error", func(t *testing.T) {
		_, err := r.Get(ctx, filepath.Join("..", "secret.txt"))

		if err == nil {
			t.Fatal("escape succeeded")
		}

		if errors.Is(err, fs.ErrNotExist) {
			t.Errorf("an escape was reported as a missing file, hiding the refusal: %v", err)
		}
	})
}

func TestRootedDelete(t *testing.T) {
	root, outside := rootedFixture(t)
	r := openRooted(t, root)
	ctx := context.Background()

	t.Run("removes a file", func(t *testing.T) {
		writeFile(t, filepath.Join(root, "gone.txt"), "x")

		if err := r.Delete("gone.txt"); err != nil {
			t.Fatal(err)
		}

		if r.Exists("gone.txt") {
			t.Error("Delete left the file behind")
		}
	})

	t.Run("ignores a missing file", func(t *testing.T) {
		if err := r.Delete("never-existed.txt"); err != nil {
			t.Errorf("Delete on a missing name = %v, want nil", err)
		}
	})

	t.Run("DeleteAll removes a tree", func(t *testing.T) {
		writeFile(t, filepath.Join(root, "tree", "child.txt"), "x")

		if err := r.DeleteAll(ctx, "tree"); err != nil {
			t.Fatal(err)
		}

		if r.Exists("tree") {
			t.Error("DeleteAll left the tree behind")
		}
	})

	t.Run("cannot delete outside the root", func(t *testing.T) {
		if err := r.DeleteAll(ctx, filepath.Join("..", "secret.txt")); err == nil {
			t.Error("DeleteAll escaped the root")
		}

		if _, err := os.Stat(outside); err != nil {
			t.Errorf("the file outside the root was deleted: %v", err)
		}
	})
}

func TestRootedListing(t *testing.T) {
	root, _ := rootedFixture(t)

	writeFile(t, filepath.Join(root, "a.txt"), "a")
	writeFile(t, filepath.Join(root, ".hidden"), "h")
	makeDir(t, filepath.Join(root, "sub"))

	r := openRooted(t, root)
	ctx := context.Background()

	files, err := r.Files(ctx, ".")

	if err != nil {
		t.Fatal(err)
	}

	if len(files) != 1 || filepath.Base(files[0]) != "a.txt" {
		t.Errorf("Files = %v, want just a.txt (hidden excluded, dirs excluded)", files)
	}

	withHidden, err := r.Files(ctx, ".", true)

	if err != nil {
		t.Fatal(err)
	}

	if len(withHidden) != 2 {
		t.Errorf("Files(hidden=true) = %v, want 2 entries", withHidden)
	}

	dirs, err := r.Directories(ctx, ".")

	if err != nil {
		t.Fatal(err)
	}

	if len(dirs) != 1 || filepath.Base(dirs[0]) != "sub" {
		t.Errorf("Directories = %v, want just sub", dirs)
	}
}

func TestRootedMakeDirectory(t *testing.T) {
	root, _ := rootedFixture(t)
	r := openRooted(t, root)

	if err := r.MakeDirectory(filepath.Join("a", "b", "c")); err != nil {
		t.Fatal(err)
	}

	if !r.IsDirectory(filepath.Join("a", "b", "c")) {
		t.Error("MakeDirectory did not create the nested directory")
	}

	if err := r.MakeExclusiveDirectory(filepath.Join("a", "b", "c")); !errors.Is(err, fs.ErrExist) {
		t.Errorf("MakeExclusiveDirectory on an existing name = %v, want fs.ErrExist", err)
	}

	if err := r.MakeDirectory(filepath.Join("..", "escaped")); err == nil {
		t.Error("MakeDirectory escaped the root")
	}
}

func TestAtRejectsMissingRoot(t *testing.T) {
	_, err := filesystem.At(filepath.Join(t.TempDir(), "no-such-dir"))

	if !errors.Is(err, fs.ErrNotExist) {
		t.Errorf("At on a missing root = %v, want fs.ErrNotExist", err)
	}
}

func TestRootedPathAndFS(t *testing.T) {
	root, _ := rootedFixture(t)
	writeFile(t, filepath.Join(root, "walk-me.txt"), "x")

	r := openRooted(t, root)

	if r.Path() != root {
		t.Errorf("Path = %q, want %q", r.Path(), root)
	}

	matches, err := fs.Glob(r.FS(), "*.txt")

	if err != nil {
		t.Fatal(err)
	}

	if len(matches) != 1 || matches[0] != "walk-me.txt" {
		t.Errorf("fs.Glob over FS() = %v, want [walk-me.txt]", matches)
	}
}

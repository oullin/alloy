package filesystem_test

import (
	"bytes"
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"hara.sh/alloy/filesystem"
)

func TestIsLinkAndReadLink(t *testing.T) {
	requireSymlinks(t)

	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	link := filepath.Join(dir, "link.txt")

	writeFile(t, target, "payload")

	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	f := newFilesystem()

	if !f.IsLink(link) {
		t.Error("IsLink(link) = false, want true")
	}

	if f.IsLink(target) {
		t.Error("IsLink(target) = true, want false for a regular file")
	}

	got, err := f.ReadLink(link)

	if err != nil {
		t.Fatal(err)
	}

	if got != target {
		t.Errorf("ReadLink = %q, want %q", got, target)
	}
}

// TestInfoFollowsLinksAndLinkInfoDoesNot is the reason both methods exist: the
// pre-existing IsFile/IsDirectory both stat through links and so can never
// report one.
func TestInfoFollowsLinksAndLinkInfoDoesNot(t *testing.T) {
	requireSymlinks(t)

	dir := t.TempDir()
	target := filepath.Join(dir, "target.txt")
	link := filepath.Join(dir, "link.txt")

	writeFile(t, target, "payload")

	if err := os.Symlink(target, link); err != nil {
		t.Fatal(err)
	}

	f := newFilesystem()

	info, err := f.Info(link)

	if err != nil {
		t.Fatal(err)
	}

	if info.Mode()&fs.ModeSymlink != 0 {
		t.Error("Info(link) reported a symlink; it should follow the link")
	}

	if info.Size() != int64(len("payload")) {
		t.Errorf("Info(link).Size() = %d, want the target's size %d", info.Size(), len("payload"))
	}

	linkInfo, err := f.LinkInfo(link)

	if err != nil {
		t.Fatal(err)
	}

	if linkInfo.Mode()&fs.ModeSymlink == 0 {
		t.Error("LinkInfo(link) did not report a symlink; it should not follow the link")
	}
}

func TestDeleteAll(t *testing.T) {
	f := newFilesystem()
	ctx := context.Background()

	t.Run("removes a non-empty directory", func(t *testing.T) {
		dir := t.TempDir()
		nested := filepath.Join(dir, "nested")

		writeFile(t, filepath.Join(nested, "child.txt"), "x")

		if err := f.DeleteAll(ctx, nested); err != nil {
			t.Fatal(err)
		}

		if f.Exists(nested) {
			t.Error("DeleteAll left the directory behind")
		}
	})

	t.Run("removes a regular file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "file.txt")
		writeFile(t, path, "x")

		if err := f.DeleteAll(ctx, path); err != nil {
			t.Fatal(err)
		}

		if f.Exists(path) {
			t.Error("DeleteAll left the file behind")
		}
	})

	// The idempotence that Delete has and DeleteDirectory lacks: a missing path
	// is not an error, and a repeat call is a no-op.
	t.Run("is idempotent on a missing path", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "never-existed")

		if err := f.DeleteAll(ctx, path); err != nil {
			t.Errorf("DeleteAll on a missing path = %v, want nil", err)
		}

		if err := f.DeleteAll(ctx, path); err != nil {
			t.Errorf("repeated DeleteAll = %v, want nil", err)
		}
	})

	t.Run("stops on a cancelled context", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "file.txt")
		writeFile(t, path, "x")

		cancelled, cancel := context.WithCancel(context.Background())
		cancel()

		if err := f.DeleteAll(cancelled, path); !errors.Is(err, context.Canceled) {
			t.Errorf("DeleteAll with a cancelled ctx = %v, want context.Canceled", err)
		}

		if !f.Exists(path) {
			t.Error("DeleteAll removed the file despite a cancelled context")
		}
	})
}

func TestMakeExclusiveDirectory(t *testing.T) {
	f := newFilesystem()

	t.Run("creates a directory", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "fresh")

		if err := f.MakeExclusiveDirectory(path); err != nil {
			t.Fatal(err)
		}

		if !f.IsDirectory(path) {
			t.Error("MakeExclusiveDirectory did not create the directory")
		}
	})

	// This is the property that makes it usable as an atomic claim, and the one
	// MakeDirectory cannot offer.
	t.Run("fails when the path already exists", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "taken")
		makeDir(t, path)

		err := f.MakeExclusiveDirectory(path)

		if !errors.Is(err, fs.ErrExist) {
			t.Errorf("MakeExclusiveDirectory on an existing path = %v, want fs.ErrExist", err)
		}
	})

	t.Run("does not create parents", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "missing-parent", "child")

		if err := f.MakeExclusiveDirectory(path); !errors.Is(err, fs.ErrNotExist) {
			t.Errorf("MakeExclusiveDirectory with a missing parent = %v, want fs.ErrNotExist", err)
		}
	})
}

// TestMakeDirectoryStaysRecursive pins a deliberate decision: Laravel's
// makeDirectory is non-recursive, but flipping this one would silently break
// downstream callers that rely on it creating parents (a nested path would fail
// at runtime with ENOENT while every other test still passed).
// MakeExclusiveDirectory covers the non-recursive case instead.
func TestMakeDirectoryStaysRecursive(t *testing.T) {
	f := newFilesystem()
	path := filepath.Join(t.TempDir(), "a", "b", "c")

	if err := f.MakeDirectory(path); err != nil {
		t.Fatalf("MakeDirectory on a nested path = %v, want nil (it must create parents)", err)
	}

	if !f.IsDirectory(path) {
		t.Error("MakeDirectory did not create the nested directory")
	}

	if err := f.MakeDirectory(path); err != nil {
		t.Errorf("MakeDirectory on an existing path = %v, want nil", err)
	}
}

func TestMakeTempDirectory(t *testing.T) {
	f := newFilesystem()
	parent := t.TempDir()

	first, err := f.MakeTempDirectory(parent, "alloy-*")

	if err != nil {
		t.Fatal(err)
	}

	second, err := f.MakeTempDirectory(parent, "alloy-*")

	if err != nil {
		t.Fatal(err)
	}

	if first == second {
		t.Errorf("MakeTempDirectory returned the same path twice: %q", first)
	}

	for _, path := range []string{first, second} {
		if !f.IsDirectory(path) {
			t.Errorf("MakeTempDirectory(%q) did not create a directory", path)
		}

		if filepath.Dir(path) != parent {
			t.Errorf("MakeTempDirectory created %q outside %q", path, parent)
		}

		if !strings.HasPrefix(filepath.Base(path), "alloy-") {
			t.Errorf("MakeTempDirectory ignored the pattern: %q", filepath.Base(path))
		}
	}
}

func TestMakeTempFile(t *testing.T) {
	f := newFilesystem()
	parent := t.TempDir()

	file, err := f.MakeTempFile(parent, "alloy-*.txt")

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = file.Close()
	}()

	if _, err := file.WriteString("written through the handle"); err != nil {
		t.Fatal(err)
	}

	if filepath.Dir(file.Name()) != parent {
		t.Errorf("MakeTempFile created %q outside %q", file.Name(), parent)
	}

	if !f.IsFile(file.Name()) {
		t.Error("MakeTempFile did not create a regular file")
	}
}

func TestPutStream(t *testing.T) {
	f := newFilesystem()
	ctx := context.Background()

	t.Run("writes the reader and creates parents", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "nested", "out.txt")
		content := "streamed payload"

		if err := f.PutStream(ctx, path, strings.NewReader(content)); err != nil {
			t.Fatal(err)
		}

		got, err := f.Get(ctx, path)

		if err != nil {
			t.Fatal(err)
		}

		if string(got) != content {
			t.Errorf("PutStream wrote %q, want %q", got, content)
		}
	})

	t.Run("rejects a nil reader without creating a file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "nested", "out.txt")

		err := f.PutStream(ctx, path, nil)

		if !errors.Is(err, filesystem.ErrNilReader) {
			t.Errorf("PutStream with a nil reader = %v, want ErrNilReader", err)
		}

		if f.Exists(path) {
			t.Error("PutStream with a nil reader created a file")
		}
	})

	t.Run("truncates an existing file", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "out.txt")
		writeFile(t, path, "a much longer previous body")

		if err := f.PutStream(ctx, path, strings.NewReader("short")); err != nil {
			t.Fatal(err)
		}

		got, err := f.Get(ctx, path)

		if err != nil {
			t.Fatal(err)
		}

		if string(got) != "short" {
			t.Errorf("PutStream left stale bytes: got %q", got)
		}
	})

	t.Run("honours an explicit mode", func(t *testing.T) {
		requirePermissionBits(t)

		path := filepath.Join(t.TempDir(), "out.txt")

		if err := f.PutStream(ctx, path, strings.NewReader("x"), 0o600); err != nil {
			t.Fatal(err)
		}

		info, err := f.Info(path)

		if err != nil {
			t.Fatal(err)
		}

		if info.Mode().Perm() != 0o600 {
			t.Errorf("PutStream mode = %v, want 0600", info.Mode().Perm())
		}
	})

	t.Run("handles a large multi-chunk reader", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "big.bin")
		content := bytes.Repeat([]byte("alloy"), 500_000)

		if err := f.PutStream(ctx, path, bytes.NewReader(content)); err != nil {
			t.Fatal(err)
		}

		got, err := f.Get(ctx, path)

		if err != nil {
			t.Fatal(err)
		}

		if !bytes.Equal(got, content) {
			t.Errorf("PutStream corrupted a %d-byte payload (got %d bytes)", len(content), len(got))
		}
	})

	t.Run("stops on a cancelled context", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "out.txt")

		cancelled, cancel := context.WithCancel(context.Background())
		cancel()

		if err := f.PutStream(cancelled, path, strings.NewReader("x")); !errors.Is(err, context.Canceled) {
			t.Errorf("PutStream with a cancelled ctx = %v, want context.Canceled", err)
		}
	})
}

// TestDeleteRemovesEmptyDirectories pins what Delete actually does. The doc
// comment used to claim it would not touch directories, which was false: it
// unlinks empty ones, and only a non-empty one is an error.
func TestDeleteRemovesEmptyDirectories(t *testing.T) {
	f := newFilesystem()

	t.Run("removes an empty directory", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "empty")
		makeDir(t, path)

		if err := f.Delete(path); err != nil {
			t.Fatalf("Delete on an empty directory = %v, want nil", err)
		}

		if f.Exists(path) {
			t.Error("Delete left the empty directory behind")
		}
	})

	t.Run("errors on a non-empty directory", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "full")
		writeFile(t, filepath.Join(path, "child.txt"), "x")

		if err := f.Delete(path); err == nil {
			t.Error("Delete on a non-empty directory = nil, want an error")
		}

		if !f.Exists(path) {
			t.Error("Delete removed a non-empty directory")
		}
	})
}

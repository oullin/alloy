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

// TestDocsExamplesCompile exercises the snippets published in
// web/docs/packages/filesystem.md. The previous docs described a
// service-provider API that never existed, so the examples are pinned here to
// keep the page honest.
func TestDocsExamplesCompile(t *testing.T) {
	dir := t.TempDir()
	data := []byte("avatar bytes")

	// --- Basic Usage ---
	fsys := filesystem.New()
	ctx := context.Background()

	if err := fsys.Put(ctx, filepath.Join(dir, "uploads/avatar-42.png"), data); err != nil {
		t.Fatal(err)
	}

	if _, err := fsys.Get(ctx, filepath.Join(dir, "uploads/avatar-42.png")); err != nil {
		t.Fatal(err)
	}

	if err := fsys.PutStream(ctx, filepath.Join(dir, "uploads/big.iso"), strings.NewReader("x")); err != nil {
		t.Fatal(err)
	}

	if err := fsys.Copy(ctx, filepath.Join(dir, "uploads/avatar-42.png"), filepath.Join(dir, "backup/avatar-42.png")); err != nil {
		t.Fatal(err)
	}

	if err := fsys.Delete(filepath.Join(dir, "backup/avatar-42.png")); err != nil {
		t.Fatal(err)
	}

	// Move does not create parents — the doc says so because of this.
	if err := fsys.MakeDirectory(filepath.Join(dir, "archive")); err != nil {
		t.Fatal(err)
	}

	if err := fsys.Move(filepath.Join(dir, "uploads/avatar-42.png"), filepath.Join(dir, "archive/avatar-42.png")); err != nil {
		t.Fatal(err)
	}

	// --- Untrusted Paths ---
	root := filepath.Join(dir, "srv")

	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}

	uploads, err := filesystem.At(root)

	if err != nil {
		t.Fatal(err)
	}

	defer uploads.Close()

	if _, err = uploads.Get(ctx, "../escape"); err == nil {
		t.Fatal("docs claim escapes are refused, but this one succeeded")
	}

	// --- Errors ---
	if _, err := fsys.Get(ctx, filepath.Join(dir, "nope")); !errors.Is(err, fs.ErrNotExist) {
		t.Fatal("docs claim errors.Is(err, fs.ErrNotExist) works; it did not")
	}

	// --- MakeExclusiveDirectory claim ---
	claimed := filepath.Join(dir, "claim")

	if err := fsys.MakeExclusiveDirectory(claimed); err != nil {
		t.Fatal(err)
	}

	if err := fsys.MakeExclusiveDirectory(claimed); !errors.Is(err, fs.ErrExist) {
		t.Fatal("docs claim MakeExclusiveDirectory fails with fs.ErrExist when taken")
	}
}

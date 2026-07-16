package foundation_test

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oullin/alloy/pkg/hub/filesystem"
	"github.com/oullin/alloy/pkg/hub/httpx/foundation"
)

// FileStore previously declared Put(path, io.Reader), which neither filesystem
// type could satisfy — its only implementation was the test double in this
// package, so uploads had no production backend at all. These assertions are
// the guard against that regressing.
var (
	_ foundation.FileStore = (*filesystem.Local)(nil)
	_ foundation.FileStore = (*filesystem.Rooted)(nil)
)

func TestStoreAsWritesThroughLocal(t *testing.T) {
	t.Parallel()

	dir := t.TempDir()
	file := createTestUploadedFile(t, "doc", "test.txt", "real disk content")

	path, err := file.StoreAs(context.Background(), dir, "upload.txt", filesystem.New())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("upload was not written to disk: %v", err)
	}

	if string(got) != "real disk content" {
		t.Fatalf("stored content = %q, want %q", got, "real disk content")
	}
}

func TestStoreAsWritesThroughRooted(t *testing.T) {
	t.Parallel()

	root := t.TempDir()
	store, err := filesystem.At(root)

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = store.Close()
	}()

	file := createTestUploadedFile(t, "doc", "test.txt", "sandboxed content")

	path, err := file.StoreAs(context.Background(), "uploads", "upload.txt", store)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	got, err := os.ReadFile(filepath.Join(root, path))

	if err != nil {
		t.Fatalf("upload was not written inside the root: %v", err)
	}

	if string(got) != "sandboxed content" {
		t.Fatalf("stored content = %q, want %q", got, "sandboxed content")
	}
}

// TestStoreAsThroughRootedRefusesTraversal is why Rooted is the recommended
// upload backend: the name comes from the client, so it is hostile input.
func TestStoreAsThroughRootedRefusesTraversal(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	root := filepath.Join(base, "uploads")
	victim := filepath.Join(base, "authorized_keys")

	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(victim, []byte("original"), 0o600); err != nil {
		t.Fatal(err)
	}

	store, err := filesystem.At(root)

	if err != nil {
		t.Fatal(err)
	}

	defer func() {
		_ = store.Close()
	}()

	file := createTestUploadedFile(t, "doc", "evil.txt", "pwned")

	_, err = file.StoreAs(context.Background(), ".", filepath.Join("..", "authorized_keys"), store)

	if err == nil {
		t.Fatal("StoreAs wrote outside the root via a traversing name")
	}

	got, err := os.ReadFile(victim)

	if err != nil {
		t.Fatal(err)
	}

	if string(got) != "original" {
		t.Fatalf("the file outside the root was overwritten: %q", got)
	}
}

// The same traversal through a plain Local store lands wherever it points,
// which is the contrast that makes Rooted worth reaching for.
func TestStoreAsThroughLocalDoesNotSandbox(t *testing.T) {
	t.Parallel()

	base := t.TempDir()
	root := filepath.Join(base, "uploads")

	if err := os.MkdirAll(root, 0o755); err != nil {
		t.Fatal(err)
	}

	file := createTestUploadedFile(t, "doc", "evil.txt", "escaped")

	path, err := file.StoreAs(context.Background(), root, filepath.Join("..", "escaped.txt"), filesystem.New())

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.HasPrefix(filepath.Clean(path), filepath.Clean(base)) {
		t.Fatalf("test setup is wrong: %q", path)
	}

	if _, err := os.Stat(filepath.Join(base, "escaped.txt")); err != nil {
		t.Fatalf("expected Local to write outside the nominal directory, documenting why Rooted exists: %v", err)
	}
}

func TestStorePubliclyUsesPublicMode(t *testing.T) {
	t.Parallel()

	file := createTestUploadedFile(t, "doc", "test.txt", "public")
	store := newMemoryFileStore()

	path, err := file.StorePublicly(context.Background(), "uploads", store)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if store.modes[path] != 0o644 {
		t.Fatalf("StorePublicly mode = %v, want 0644", store.modes[path])
	}
}

package filesystem_test

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/oullin/alloy/pkg/hub/filesystem"
)

func newFilesystem() *filesystem.Local {
	return filesystem.New()
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}

	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
}

func makeDir(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatal(err)
	}
}

// requireSymlinks skips the test when the platform will not let us create
// symbolic links. Windows only permits this under Developer Mode or with
// SeCreateSymbolicLinkPrivilege, so an unprivileged run would otherwise fail on
// setup rather than on the behaviour under test.
func requireSymlinks(t *testing.T) {
	t.Helper()

	dir := t.TempDir()

	if err := os.Symlink(filepath.Join(dir, "target"), filepath.Join(dir, "link")); err != nil {
		if errors.Is(err, fs.ErrPermission) {
			t.Skip("symlinks are not permitted on this platform")
		}

		t.Fatal(err)
	}
}

// requirePermissionBits skips the test on Windows, whose file mode does not
// carry POSIX permission bits.
func requirePermissionBits(t *testing.T) {
	t.Helper()

	if runtime.GOOS == "windows" {
		t.Skip("permission bits are not modelled on Windows")
	}
}

package gitwork_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// repo is a real git repository built in a temporary directory. These tests
// drive git itself rather than a fake, because the behaviour under test is
// precisely how git represents linked worktrees on disk.
type repo struct {
	t   *testing.T
	dir string
}

func newRepo(t *testing.T) *repo {
	t.Helper()

	requireGit(t)

	r := &repo{t: t, dir: t.TempDir()}

	r.git("init", "--initial-branch=work")
	r.git("config", "user.email", "treex@example.test")
	r.git("config", "user.name", "treex tests")
	r.commit("README.md", "hello")

	return r
}

func (r *repo) git(args ...string) string {
	r.t.Helper()

	cmd := exec.Command("git", append([]string{"-C", r.dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_CONFIG_NOSYSTEM=1", "GIT_TERMINAL_PROMPT=0", "HOME="+r.dir)

	out, err := cmd.CombinedOutput()

	if err != nil {
		r.t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}

	return strings.TrimSpace(string(out))
}

func (r *repo) commit(name, contents string) {
	r.t.Helper()

	write(r.t, filepath.Join(r.dir, name), contents)

	r.git("add", name)
	r.git("commit", "-m", "add "+name)
}

// addWorktree creates a linked worktree, the shape whose removal needs care.
func (r *repo) addWorktree(name string) string {
	r.t.Helper()

	path := filepath.Join(r.t.TempDir(), name)

	r.git("worktree", "add", "-b", name, path)

	return path
}

// cloneTo produces a self-contained clone with a remote, the shape that can be
// removed outright.
func (r *repo) cloneTo(t *testing.T) string {
	t.Helper()

	path := filepath.Join(t.TempDir(), "clone")

	cmd := exec.Command("git", "clone", r.dir, path)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git clone: %v\n%s", err, out)
	}

	run(t, path, "config", "user.email", "treex@example.test")
	run(t, path, "config", "user.name", "treex tests")

	return path
}

func run(t *testing.T, dir string, args ...string) string {
	t.Helper()

	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	cmd.Env = append(os.Environ(), "GIT_TERMINAL_PROMPT=0")

	out, err := cmd.CombinedOutput()

	if err != nil {
		t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
	}

	return strings.TrimSpace(string(out))
}

func write(t *testing.T, path, contents string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatalf("mkdir: %v", err)
	}

	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func requireGit(t *testing.T) {
	t.Helper()

	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git is not installed")
	}
}

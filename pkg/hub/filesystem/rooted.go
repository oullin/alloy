package filesystem

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Rooted provides filesystem operations confined to a root directory.
//
// Names passed to its methods are interpreted relative to that root and cannot
// escape it — not via "..", not via an absolute path, and not via a symbolic
// link pointing outside. The confinement is enforced by the operating system
// against an open directory handle rather than by inspecting paths, so it holds
// even if the tree is rewritten mid-operation. That is the part callers cannot
// build for themselves: checking a path and then opening it are two steps, and
// anything can change in between.
//
// Use it whenever the name comes from outside the program — an upload filename,
// a request parameter, an entry in a config file.
//
// Symbolic links that stay inside the root are followed normally. Rooted
// prevents escape; it does not forbid links. Callers who need to reject them as
// a matter of policy can check LinkInfo on the final component.
//
// Its guarantee stops at the filesystem: it does not prohibit crossing mount
// points or bind mounts below the root, nor reading /proc or device files that
// live inside it, and on Unix Chmod remains racy against a file being swapped
// for a symlink. Each of those needs an attacker who already controls the root.
//
// A Rooted holds an open file descriptor and must be closed.
type Rooted struct {
	root *os.Root
}

// At opens root as a sandbox for subsequent operations. The directory must
// already exist. The caller must Close the result.
func At(root string) (*Rooted, error) {
	handle, err := os.OpenRoot(root)

	if err != nil {
		return nil, err
	}

	return &Rooted{root: handle}, nil
}

// Close releases the root's file descriptor.
func (r *Rooted) Close() error {
	return r.root.Close()
}

// Path returns the name of the root directory.
func (r *Rooted) Path() string {
	return r.root.Name()
}

// FS returns a read-only fs.FS backed by the root, for use with fs.WalkDir,
// fs.Glob, and template parsing.
func (r *Rooted) FS() fs.FS {
	return r.root.FS()
}

// Exists determines if a file or directory exists at the given name.
func (r *Rooted) Exists(name string) bool {
	_, err := r.root.Stat(name)

	return err == nil
}

// Missing determines if a file or directory is missing at the given name.
func (r *Rooted) Missing(name string) bool {
	return !r.Exists(name)
}

// IsFile determines if the given name is a regular file.
func (r *Rooted) IsFile(name string) bool {
	info, err := r.root.Stat(name)

	if err != nil {
		return false
	}

	return info.Mode().IsRegular()
}

// IsDirectory determines if the given name is a directory.
func (r *Rooted) IsDirectory(name string) bool {
	info, err := r.root.Stat(name)

	if err != nil {
		return false
	}

	return info.IsDir()
}

// IsLink determines if the given name is a symbolic link.
func (r *Rooted) IsLink(name string) bool {
	info, err := r.root.Lstat(name)

	if err != nil {
		return false
	}

	return info.Mode()&fs.ModeSymlink != 0
}

// Info returns the metadata for the given name, following symbolic links.
func (r *Rooted) Info(name string) (fs.FileInfo, error) {
	return r.root.Stat(name)
}

// LinkInfo returns the metadata for the given name without following symbolic
// links.
func (r *Rooted) LinkInfo(name string) (fs.FileInfo, error) {
	return r.root.Lstat(name)
}

// ReadLink returns the target of the symbolic link at the given name. The
// target is returned as recorded and is not itself resolved, so it may point
// outside the root; reading through the link would still be refused.
func (r *Rooted) ReadLink(name string) (string, error) {
	return r.root.Readlink(name)
}

// Get reads the entire contents of a file.
func (r *Rooted) Get(ctx context.Context, name string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	data, err := r.root.ReadFile(name)

	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
		}

		return nil, err
	}

	return data, nil
}

// Put writes contents to a file, creating it and any missing parents. An
// optional file mode can be provided; defaults to 0644.
func (r *Rooted) Put(ctx context.Context, name string, contents []byte, mode ...fs.FileMode) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	perm := fs.FileMode(0o644)

	if len(mode) > 0 {
		perm = mode[0]
	}

	if err := r.makeParents(name); err != nil {
		return err
	}

	return r.root.WriteFile(name, contents, perm)
}

// PutStream writes everything readable from contents to a file, creating it and
// any missing parents. An optional file mode can be provided; defaults to 0644.
// The file is streamed in chunks rather than buffered whole, and the write stops
// early when ctx is cancelled — leaving a partially written file behind.
func (r *Rooted) PutStream(ctx context.Context, name string, contents io.Reader, mode ...fs.FileMode) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if contents == nil {
		return ErrNilReader
	}

	perm := fs.FileMode(0o644)

	if len(mode) > 0 {
		perm = mode[0]
	}

	if err := r.makeParents(name); err != nil {
		return err
	}

	file, err := r.root.OpenFile(name, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)

	if err != nil {
		return err
	}

	if _, err := copyContext(ctx, file, contents); err != nil {
		_ = file.Close()

		return err
	}

	return file.Close()
}

// Delete removes one or more files or empty directories. Names that do not
// exist are ignored. A non-empty directory returns an error; use DeleteAll to
// remove one along with its contents.
func (r *Rooted) Delete(names ...string) error {
	for _, name := range names {
		if err := r.root.Remove(name); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}

	return nil
}

// DeleteAll removes one or more names along with any children they contain.
// Names that do not exist are ignored.
func (r *Rooted) DeleteAll(ctx context.Context, names ...string) error {
	for _, name := range names {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := r.root.RemoveAll(name); err != nil {
			return err
		}
	}

	return nil
}

// MakeDirectory creates a directory along with any missing parents. The default
// mode is 0755. It succeeds when the directory already exists.
func (r *Rooted) MakeDirectory(name string, mode ...fs.FileMode) error {
	perm := fs.FileMode(0o755)

	if len(mode) > 0 {
		perm = mode[0]
	}

	return r.root.MkdirAll(name, perm)
}

// MakeExclusiveDirectory creates a single directory. Parent directories are not
// created, and an error wrapping fs.ErrExist is returned when the name already
// exists.
func (r *Rooted) MakeExclusiveDirectory(name string, mode ...fs.FileMode) error {
	perm := fs.FileMode(0o755)

	if len(mode) > 0 {
		perm = mode[0]
	}

	return r.root.Mkdir(name, perm)
}

// Files returns the files in the given directory (non-recursive).
// Hidden files (starting with ".") are excluded unless hidden is true.
func (r *Rooted) Files(ctx context.Context, directory string, hidden ...bool) ([]string, error) {
	entries, err := r.readDir(ctx, directory)

	if err != nil {
		return nil, err
	}

	includeHidden := false

	if len(hidden) > 0 {
		includeHidden = hidden[0]
	}

	var files []string

	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		if !includeHidden && strings.HasPrefix(entry.Name(), ".") {
			continue
		}

		files = append(files, filepath.Join(directory, entry.Name()))
	}

	return files, nil
}

// Directories returns the directories in the given directory (non-recursive).
func (r *Rooted) Directories(ctx context.Context, directory string) ([]string, error) {
	entries, err := r.readDir(ctx, directory)

	if err != nil {
		return nil, err
	}

	var dirs []string

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(directory, entry.Name()))
		}
	}

	return dirs, nil
}

// makeParents creates the parent directories of name. os.Root has no ReadDir or
// MkdirAll-for-a-file helper, so the "." case has to be skipped explicitly.
func (r *Rooted) makeParents(name string) error {
	parent := filepath.Dir(name)

	if parent == "." || parent == "" || parent == string(filepath.Separator) {
		return nil
	}

	return r.root.MkdirAll(parent, 0o755)
}

// readDir lists a directory inside the root. os.Root exposes no ReadDir, so the
// directory is opened and read through its handle.
func (r *Rooted) readDir(ctx context.Context, directory string) ([]os.DirEntry, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	file, err := r.root.Open(directory)

	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: %w", ErrNotFound, err)
		}

		return nil, err
	}

	defer func() {
		_ = file.Close()
	}()

	return file.ReadDir(-1)
}

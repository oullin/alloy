package filesystem

import (
	"context"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Files returns the files in the given directory (non-recursive).
// Hidden files (starting with ".") are excluded unless hidden is true.
func (f *Local) Files(ctx context.Context, directory string, hidden ...bool) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	includeHidden := false

	if len(hidden) > 0 {
		includeHidden = hidden[0]
	}

	entries, err := os.ReadDir(directory)

	if err != nil {
		return nil, err
	}

	var files []string

	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

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

// AllFiles returns all files in the directory tree recursively.
// Hidden files and directories (starting with ".") are excluded unless hidden
// is true. The walk stops early when ctx is cancelled.
func (f *Local) AllFiles(ctx context.Context, directory string, hidden ...bool) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	includeHidden := false

	if len(hidden) > 0 {
		includeHidden = hidden[0]
	}

	var files []string

	err := filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if d.IsDir() {
			if !includeHidden && path != directory && strings.HasPrefix(d.Name(), ".") {
				return filepath.SkipDir
			}

			return nil
		}

		if !includeHidden && strings.HasPrefix(d.Name(), ".") {
			return nil
		}

		files = append(files, path)

		return nil
	})

	return files, err
}

// Directories returns the directories in the given directory (non-recursive).
func (f *Local) Directories(ctx context.Context, directory string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	entries, err := os.ReadDir(directory)

	if err != nil {
		return nil, err
	}

	var dirs []string

	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return nil, err
		}

		if entry.IsDir() {
			dirs = append(dirs, filepath.Join(directory, entry.Name()))
		}
	}

	return dirs, nil
}

// AllDirectories returns all directories in the directory tree recursively.
// The walk stops early when ctx is cancelled.
func (f *Local) AllDirectories(ctx context.Context, directory string) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	var dirs []string

	err := filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		if d.IsDir() && path != directory {
			dirs = append(dirs, path)
		}

		return nil
	})

	return dirs, err
}

// EnsureDirectoryExists creates the directory if it does not already exist.
// The default mode is 0755.
func (f *Local) EnsureDirectoryExists(path string, mode ...fs.FileMode) error {
	perm := fs.FileMode(0o755)

	if len(mode) > 0 {
		perm = mode[0]
	}

	return os.MkdirAll(path, perm)
}

// MakeDirectory creates a directory along with any missing parents. The
// default mode is 0755. It succeeds when the directory already exists; use
// MakeExclusiveDirectory when creation must be exclusive.
func (f *Local) MakeDirectory(path string, mode ...fs.FileMode) error {
	perm := fs.FileMode(0o755)

	if len(mode) > 0 {
		perm = mode[0]
	}

	return os.MkdirAll(path, perm)
}

// MakeExclusiveDirectory creates a single directory. The default mode is 0755.
// Parent directories are not created, and an error wrapping fs.ErrExist is
// returned when the path already exists — which makes it usable as an atomic
// claim on a name.
func (f *Local) MakeExclusiveDirectory(path string, mode ...fs.FileMode) error {
	perm := fs.FileMode(0o755)

	if len(mode) > 0 {
		perm = mode[0]
	}

	return os.Mkdir(path, perm)
}

// MakeTempDirectory creates a new directory with a name built from pattern
// inside dir and returns its path. When dir is empty the default directory for
// temporary files is used. The caller is responsible for removing it.
func (f *Local) MakeTempDirectory(dir, pattern string) (string, error) {
	return os.MkdirTemp(dir, pattern)
}

// MoveDirectory moves a directory from one location to another.
// When overwrite is true, any existing destination directory is removed first.
func (f *Local) MoveDirectory(ctx context.Context, from, to string, overwrite ...bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	shouldOverwrite := false

	if len(overwrite) > 0 {
		shouldOverwrite = overwrite[0]
	}

	if _, err := os.Stat(to); err == nil {
		if !shouldOverwrite {
			return fs.ErrExist
		}

		if err := os.RemoveAll(to); err != nil {
			return err
		}
	} else if !errors.Is(err, fs.ErrNotExist) {
		return err
	}

	// Try atomic rename first.
	if err := os.Rename(from, to); err == nil {
		return nil
	}

	// Fallback: copy + delete (handles cross-device moves).
	if err := f.CopyDirectory(ctx, from, to); err != nil {
		return err
	}

	return os.RemoveAll(from)
}

// CopyDirectory recursively copies a directory and its contents.
// The walk stops early when ctx is cancelled.
func (f *Local) CopyDirectory(ctx context.Context, directory, destination string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	info, err := os.Stat(directory)

	if err != nil {
		return err
	}

	if !info.IsDir() {
		return ErrNotDirectory
	}

	return filepath.WalkDir(directory, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if err := ctx.Err(); err != nil {
			return err
		}

		rel, err := filepath.Rel(directory, path)

		if err != nil {
			return err
		}

		target := filepath.Join(destination, rel)

		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		return f.Copy(ctx, path, target)
	})
}

// DeleteDirectory removes a directory and all of its contents.
// When preserve is true, the directory itself is kept but its contents are removed.
func (f *Local) DeleteDirectory(ctx context.Context, directory string, preserve ...bool) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	info, err := os.Stat(directory)

	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}

		return err
	}

	if !info.IsDir() {
		return ErrNotDirectory
	}

	shouldPreserve := false

	if len(preserve) > 0 {
		shouldPreserve = preserve[0]
	}

	if shouldPreserve {
		return f.cleanDir(ctx, directory)
	}

	return os.RemoveAll(directory)
}

// DeleteDirectories removes all subdirectories within the given directory,
// leaving files intact.
func (f *Local) DeleteDirectories(ctx context.Context, directory string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	entries, err := os.ReadDir(directory)

	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}

		if entry.IsDir() {
			if err := os.RemoveAll(filepath.Join(directory, entry.Name())); err != nil {
				return err
			}
		}
	}

	return nil
}

// CleanDirectory removes all contents of a directory but keeps the directory.
func (f *Local) CleanDirectory(ctx context.Context, directory string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	return f.cleanDir(ctx, directory)
}

func (f *Local) cleanDir(ctx context.Context, directory string) error {
	entries, err := os.ReadDir(directory)

	if err != nil {
		return err
	}

	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}

		path := filepath.Join(directory, entry.Name())

		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}

	return nil
}

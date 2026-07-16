package filesystem

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// Put writes contents to a file, creating it if necessary.
// An optional file mode can be provided; defaults to 0644.
func (f *Local) Put(ctx context.Context, path string, contents []byte, mode ...fs.FileMode) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	perm := fs.FileMode(0o644)

	if len(mode) > 0 {
		perm = mode[0]
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	return os.WriteFile(path, contents, perm)
}

// PutStream writes everything readable from contents to a file, creating it if
// necessary. An optional file mode can be provided; defaults to 0644. The file
// is streamed in chunks rather than buffered whole, so it suits uploads and
// other sources of unknown size, and the write stops early when ctx is
// cancelled — leaving a partially written file behind.
func (f *Local) PutStream(ctx context.Context, path string, contents io.Reader, mode ...fs.FileMode) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	perm := fs.FileMode(0o644)

	if len(mode) > 0 {
		perm = mode[0]
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, perm)

	if err != nil {
		return err
	}

	if _, err := copyContext(ctx, file, contents); err != nil {
		_ = file.Close()

		return err
	}

	return file.Close()
}

// MakeTempFile creates a new file with a name built from pattern inside dir and
// returns the open file. When dir is empty the default directory for temporary
// files is used. The caller is responsible for closing and removing it. The
// open file is returned rather than its path so callers need not re-open by
// name, which would reintroduce the race this avoids.
func (f *Local) MakeTempFile(dir, pattern string) (*os.File, error) {
	return os.CreateTemp(dir, pattern)
}

// Replace atomically writes content to a file using a temporary file
// and rename. An optional file mode can be provided.
func (f *Local) Replace(ctx context.Context, path string, content []byte, mode ...fs.FileMode) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	perm := fs.FileMode(0o644)

	if len(mode) > 0 {
		perm = mode[0]
	}

	dir := filepath.Dir(path)

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	tmp, err := os.CreateTemp(dir, ".tmp_*")

	if err != nil {
		return err
	}

	tmpName := tmp.Name()

	if _, err := tmp.Write(content); err != nil {
		tmp.Close()
		os.Remove(tmpName)

		return err
	}

	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)

		return err
	}

	if err := os.Chmod(tmpName, perm); err != nil {
		os.Remove(tmpName)

		return err
	}

	return os.Rename(tmpName, path)
}

// ReplaceInFile replaces all occurrences of search with replace in the file.
func (f *Local) ReplaceInFile(ctx context.Context, search, replace, path string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	data, err := os.ReadFile(path)

	if err != nil {
		return err
	}

	content := strings.ReplaceAll(string(data), search, replace)

	return f.Replace(ctx, path, []byte(content))
}

// Prepend prepends data to the beginning of a file. If the file does not
// exist, it is created with only the given data.
func (f *Local) Prepend(ctx context.Context, path string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	existing, err := os.ReadFile(path)

	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return f.Put(ctx, path, data)
		}

		return err
	}

	combined := make([]byte, 0, len(data)+len(existing))
	combined = append(combined, data...)
	combined = append(combined, existing...)

	return f.Replace(ctx, path, combined)
}

// Append appends data to the end of a file.
func (f *Local) Append(ctx context.Context, path string, data []byte) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)

	if err != nil {
		return err
	}

	_, writeErr := file.Write(data)
	closeErr := file.Close()

	if writeErr != nil {
		return writeErr
	}

	return closeErr
}

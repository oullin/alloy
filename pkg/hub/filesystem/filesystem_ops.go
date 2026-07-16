package filesystem

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// Delete removes one or more files or empty directories. Paths that do not
// exist are ignored. A non-empty directory returns an error; use DeleteAll to
// remove one along with its contents.
func (f *Local) Delete(paths ...string) error {
	for _, path := range paths {
		if err := os.Remove(path); err != nil && !errors.Is(err, fs.ErrNotExist) {
			return err
		}
	}

	return nil
}

// DeleteAll removes one or more paths along with any children they contain.
// Paths that do not exist are ignored. Unlike Delete, it removes directories
// recursively. Cancelling ctx stops the loop between paths; the removal of an
// individual path is not itself interruptible.
func (f *Local) DeleteAll(ctx context.Context, paths ...string) error {
	for _, path := range paths {
		if err := ctx.Err(); err != nil {
			return err
		}

		if err := os.RemoveAll(path); err != nil {
			return err
		}
	}

	return nil
}

// Move moves a file from path to target.
func (f *Local) Move(path, target string) error {
	return os.Rename(path, target)
}

// ReadLink returns the target of the symbolic link at the given path.
func (f *Local) ReadLink(path string) (string, error) {
	return os.Readlink(path)
}

// Copy copies a file from path to target. The copy stops early when ctx
// is cancelled.
func (f *Local) Copy(ctx context.Context, path, target string) error {
	if err := ctx.Err(); err != nil {
		return err
	}

	src, err := os.Open(path)

	if err != nil {
		return err
	}

	info, err := src.Stat()

	if err != nil {
		_ = src.Close()

		return err
	}

	if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
		_ = src.Close()

		return err
	}

	dst, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, info.Mode())

	if err != nil {
		_ = src.Close()

		return err
	}

	if _, err := copyContext(ctx, dst, src); err != nil {
		_ = dst.Close()
		_ = src.Close()

		return err
	}

	if err := dst.Close(); err != nil {
		_ = src.Close()

		return err
	}

	return src.Close()
}

// Link creates a symbolic link.
func (f *Local) Link(target, link string) error {
	return os.Symlink(target, link)
}

// RelativeLink creates a relative symbolic link.
func (f *Local) RelativeLink(target, link string) error {
	linkDir := filepath.Dir(link)

	rel, err := filepath.Rel(linkDir, target)

	if err != nil {
		return err
	}

	return os.Symlink(rel, link)
}

// copyContext copies from src to dst in 1 MiB chunks, checking for context
// cancellation between chunks.
func copyContext(ctx context.Context, dst io.Writer, src io.Reader) (int64, error) {
	buf := make([]byte, 1<<20)

	var written int64

	for {
		if err := ctx.Err(); err != nil {
			return written, err
		}

		n, readErr := src.Read(buf)

		if n > 0 {
			w, writeErr := dst.Write(buf[:n])
			written += int64(w)

			if writeErr != nil {
				return written, writeErr
			}

			if w < n {
				return written, io.ErrShortWrite
			}
		}

		if readErr == io.EOF {
			return written, nil
		}

		if readErr != nil {
			return written, readErr
		}
	}
}

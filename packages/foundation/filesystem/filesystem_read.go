package filesystem

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"iter"
	"os"
	"strings"
)

// Get reads the entire contents of a file.
func (f *Local) Get(ctx context.Context, path string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	data, err := os.ReadFile(path)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return data, nil
}

// JSON reads a file and unmarshals its JSON contents into v.
func (f *Local) JSON(ctx context.Context, path string, v any) error {
	data, err := f.Get(ctx, path)

	if err != nil {
		return err
	}

	return json.Unmarshal(data, v)
}

// SharedGet reads a file's contents while holding a shared (read) lock.
func (f *Local) SharedGet(ctx context.Context, path string) ([]byte, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	file, err := os.Open(path)

	if err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	defer file.Close()

	if err := lockShared(file); err != nil {
		return nil, err
	}

	defer func() {
		_ = unlockFile(file)
	}()

	return os.ReadFile(path)
}

// Lines returns an iterator that yields each line of the file.
// The iterator stops yielding when ctx is cancelled.
func (f *Local) Lines(ctx context.Context, path string) (iter.Seq[string], error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			return nil, ErrNotFound
		}

		return nil, err
	}

	return func(yield func(string) bool) {
		file, err := os.Open(path)

		if err != nil {
			return
		}

		defer file.Close()

		reader := bufio.NewReader(file)

		for {
			if ctx.Err() != nil {
				return
			}

			line, err := reader.ReadString('\n')

			if len(line) > 0 {
				line = strings.TrimSuffix(line, "\n")
				line = strings.TrimSuffix(line, "\r")

				if !yield(line) {
					return
				}
			}

			if err == io.EOF {
				return
			}

			if err != nil {
				return
			}
		}
	}, nil
}

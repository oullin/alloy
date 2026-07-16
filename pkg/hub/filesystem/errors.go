package filesystem

import (
	"errors"
	"fmt"
	"io/fs"
)

var (
	// ErrNotFound wraps fs.ErrNotExist so callers can test for a missing path
	// with errors.Is(err, fs.ErrNotExist) regardless of which method produced
	// the error.
	ErrNotFound      = fmt.Errorf("filesystem: file not found (%w)", fs.ErrNotExist)
	ErrNotDirectory  = errors.New("filesystem: not a directory")
	ErrHashAlgorithm = errors.New("filesystem: unsupported hash algorithm")
	ErrLockFailed    = errors.New("filesystem: failed to acquire lock")
	ErrLocked        = errors.New("filesystem: lock is held elsewhere")
	ErrEscapesRoot   = errors.New("filesystem: path escapes the root directory")
)

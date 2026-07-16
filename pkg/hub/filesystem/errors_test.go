package filesystem_test

import (
	"context"
	"errors"
	"io/fs"
	"path/filepath"
	"strings"
	"testing"

	"github.com/oullin/alloy/pkg/hub/filesystem"
)

// TestMissingPathErrorsAreErrNotExist pins the invariant that every method
// which can be handed a missing path reports it as fs.ErrNotExist. Before this
// was enforced, Get/SharedGet/Lines returned a bare sentinel that discarded the
// *fs.PathError, so errors.Is(err, fs.ErrNotExist) was false for those three
// and true for every other method — with no signal to the caller.
func TestMissingPathErrorsAreErrNotExist(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "does-not-exist")
	ctx := context.Background()

	cases := map[string]func(f *filesystem.Local) error{
		"Get": func(f *filesystem.Local) error {
			_, err := f.Get(ctx, missing)

			return err
		},
		"JSON": func(f *filesystem.Local) error {
			var v any

			return f.JSON(ctx, missing, &v)
		},
		"SharedGet": func(f *filesystem.Local) error {
			_, err := f.SharedGet(ctx, missing)

			return err
		},
		"Lines": func(f *filesystem.Local) error {
			_, err := f.Lines(ctx, missing)

			return err
		},
		"Size": func(f *filesystem.Local) error {
			_, err := f.Size(missing)

			return err
		},
		"LastModified": func(f *filesystem.Local) error {
			_, err := f.LastModified(missing)

			return err
		},
		"Type": func(f *filesystem.Local) error {
			_, err := f.Type(missing)

			return err
		},
		"MimeType": func(f *filesystem.Local) error {
			_, err := f.MimeType(missing)

			return err
		},
		"Hash": func(f *filesystem.Local) error {
			_, err := f.Hash(ctx, missing)

			return err
		},
		"Chmod": func(f *filesystem.Local) error {
			return f.Chmod(missing, 0o644)
		},
		"Info": func(f *filesystem.Local) error {
			_, err := f.Info(missing)

			return err
		},
		"LinkInfo": func(f *filesystem.Local) error {
			_, err := f.LinkInfo(missing)

			return err
		},
		"ReadLink": func(f *filesystem.Local) error {
			_, err := f.ReadLink(missing)

			return err
		},
	}

	f := filesystem.New()

	for name, call := range cases {
		t.Run(name, func(t *testing.T) {
			err := call(f)

			if err == nil {
				t.Fatalf("%s on a missing path returned nil error", name)
			}

			if !errors.Is(err, fs.ErrNotExist) {
				t.Errorf("%s: errors.Is(err, fs.ErrNotExist) = false, want true (err = %v)", name, err)
			}
		})
	}
}

func TestErrNotFoundIsErrNotExist(t *testing.T) {
	if !errors.Is(filesystem.ErrNotFound, fs.ErrNotExist) {
		t.Error("ErrNotFound does not satisfy errors.Is(_, fs.ErrNotExist)")
	}
}

// TestReadErrorsKeepSentinelAndPath guards both halves of the fix: the sentinel
// stays matchable and the wrapped *fs.PathError puts the offending path back
// into the message.
func TestReadErrorsKeepSentinelAndPath(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "absent.txt")
	f := filesystem.New()

	_, err := f.Get(context.Background(), missing)

	if !errors.Is(err, filesystem.ErrNotFound) {
		t.Errorf("Get error lost the ErrNotFound sentinel: %v", err)
	}

	if !strings.Contains(err.Error(), missing) {
		t.Errorf("Get error dropped the path from its message: %v", err)
	}
}

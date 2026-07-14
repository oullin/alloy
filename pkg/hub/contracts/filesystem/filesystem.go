package filesystem

import (
	"context"
	"io/fs"
	"iter"
)

// Filesystem defines the contract for local filesystem operations.
type Filesystem interface {
	Exists(path string) bool
	Missing(path string) bool
	IsFile(path string) bool
	IsDirectory(path string) bool
	IsEmptyDirectory(directory string, ignoreDotFiles bool) (bool, error)
	IsReadable(path string) bool
	IsWritable(path string) bool
	Get(ctx context.Context, path string) ([]byte, error)
	JSON(ctx context.Context, path string, v any) error
	SharedGet(ctx context.Context, path string) ([]byte, error)
	Lines(ctx context.Context, path string) (iter.Seq[string], error)
	Put(ctx context.Context, path string, contents []byte, mode ...fs.FileMode) error
	Replace(ctx context.Context, path string, content []byte, mode ...fs.FileMode) error
	ReplaceInFile(ctx context.Context, search, replace, path string) error
	Prepend(ctx context.Context, path string, data []byte) error
	Append(ctx context.Context, path string, data []byte) error
	Hash(ctx context.Context, path string, algorithm ...string) (string, error)
	HasSameHash(ctx context.Context, firstFile, secondFile string) (bool, error)
	Type(path string) (string, error)
	MimeType(path string) (string, error)
	GuessExtension(path string) (string, error)
	Size(path string) (int64, error)
	LastModified(path string) (int64, error)
	Chmod(path string, mode fs.FileMode) error
	Name(path string) string
	Basename(path string) string
	Dirname(path string) string
	Extension(path string) string
	Glob(pattern string) ([]string, error)
	Delete(paths ...string) error
	Move(path, target string) error
	Copy(ctx context.Context, path, target string) error
	Link(target, link string) error
	RelativeLink(target, link string) error
	Files(ctx context.Context, directory string, hidden ...bool) ([]string, error)
	AllFiles(ctx context.Context, directory string, hidden ...bool) ([]string, error)
	Directories(ctx context.Context, directory string) ([]string, error)
	AllDirectories(ctx context.Context, directory string) ([]string, error)
	EnsureDirectoryExists(path string, mode ...fs.FileMode) error
	MakeDirectory(path string, mode ...fs.FileMode) error
	MoveDirectory(ctx context.Context, from, to string, overwrite ...bool) error
	CopyDirectory(ctx context.Context, directory, destination string) error
	DeleteDirectory(ctx context.Context, directory string, preserve ...bool) error
	DeleteDirectories(ctx context.Context, directory string) error
	CleanDirectory(ctx context.Context, directory string) error
}

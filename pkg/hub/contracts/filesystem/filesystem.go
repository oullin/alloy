package filesystem

import (
	"context"
	"io"
	"io/fs"
	"iter"
	"os"
)

// Filesystem defines the contract for local filesystem operations.
//
// MakeTempFile returns an *os.File rather than something from io/fs because
// io/fs models only read-only filesystems and has no writable-handle type. This
// is the local-disk contract, so the dependency is honest.
type Filesystem interface {
	Exists(path string) bool
	Missing(path string) bool
	IsFile(path string) bool
	IsDirectory(path string) bool
	IsLink(path string) bool
	IsEmptyDirectory(directory string, ignoreDotFiles bool) (bool, error)
	IsReadable(path string) bool
	IsWritable(path string) bool
	Info(path string) (fs.FileInfo, error)
	LinkInfo(path string) (fs.FileInfo, error)
	Get(ctx context.Context, path string) ([]byte, error)
	JSON(ctx context.Context, path string, v any) error
	SharedGet(ctx context.Context, path string) ([]byte, error)
	Lines(ctx context.Context, path string) (iter.Seq[string], error)
	Put(ctx context.Context, path string, contents []byte, mode ...fs.FileMode) error
	PutStream(ctx context.Context, path string, contents io.Reader, mode ...fs.FileMode) error
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
	DeleteAll(ctx context.Context, paths ...string) error
	Move(path, target string) error
	Copy(ctx context.Context, path, target string) error
	Link(target, link string) error
	RelativeLink(target, link string) error
	ReadLink(path string) (string, error)
	MakeTempFile(dir, pattern string) (*os.File, error)
	Files(ctx context.Context, directory string, hidden ...bool) ([]string, error)
	AllFiles(ctx context.Context, directory string, hidden ...bool) ([]string, error)
	Directories(ctx context.Context, directory string) ([]string, error)
	AllDirectories(ctx context.Context, directory string) ([]string, error)
	EnsureDirectoryExists(path string, mode ...fs.FileMode) error
	MakeDirectory(path string, mode ...fs.FileMode) error
	MakeExclusiveDirectory(path string, mode ...fs.FileMode) error
	MakeTempDirectory(dir, pattern string) (string, error)
	MoveDirectory(ctx context.Context, from, to string, overwrite ...bool) error
	CopyDirectory(ctx context.Context, directory, destination string) error
	DeleteDirectory(ctx context.Context, directory string, preserve ...bool) error
	DeleteDirectories(ctx context.Context, directory string) error
	CleanDirectory(ctx context.Context, directory string) error
}

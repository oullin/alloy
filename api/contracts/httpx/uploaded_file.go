package httpx

import "io"

// FileStore abstracts file storage for uploaded files.
type FileStore interface {
	Put(path string, contents io.Reader) error
}

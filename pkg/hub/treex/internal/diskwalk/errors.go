package diskwalk

import "errors"

var (
	// ErrClosed reports use of a pool that has already been closed.
	ErrClosed = errors.New("diskwalk: pool is closed")

	// ErrCancelled reports that a walk stopped because its context was done.
	// The partial results are still returned alongside it, because a scan
	// interrupted halfway is still worth showing.
	ErrCancelled = errors.New("diskwalk: walk cancelled")
)

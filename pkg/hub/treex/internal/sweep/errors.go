package sweep

import "errors"

var (
	// ErrGuardRejected reports an action the guard refused. It is not a
	// failure of the sweep so much as the guard doing its job.
	ErrGuardRejected = errors.New("sweep: guard rejected the path")

	// ErrOutsideRoot reports a path that is not under any enabled provider
	// root, which should be impossible and is treated as a bug rather than a
	// user error.
	ErrOutsideRoot = errors.New("sweep: path is outside every provider root")

	// ErrProtected reports a path covered by a protect-path rule.
	ErrProtected = errors.New("sweep: path is protected")

	// ErrChanged reports a path whose identity no longer matches what the scan
	// recorded, meaning it was replaced between measuring and deleting.
	ErrChanged = errors.New("sweep: path changed since it was scanned")

	// ErrNotADirectory reports a path that is no longer the kind of thing the
	// plan described.
	ErrNotADirectory = errors.New("sweep: path is not a directory")
)

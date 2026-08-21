package report

import "errors"

var (
	// ErrUnknownFormat reports a --format value treex cannot render.
	ErrUnknownFormat = errors.New("report: unknown format")
)

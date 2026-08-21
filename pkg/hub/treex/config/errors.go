package config

import "errors"

var (
	// ErrNotFound reports that an explicitly requested config file is absent.
	// Discovery treats a missing default location as "use the built-ins", but
	// an explicit --config that does not exist is a mistake worth surfacing.
	ErrNotFound = errors.New("config: file not found")

	// ErrUnsupportedVersion reports a schema version treex cannot read.
	ErrUnsupportedVersion = errors.New("config: unsupported schema version")

	// ErrInvalidProvider reports a provider entry that cannot be used, such as
	// one with no name or no root.
	ErrInvalidProvider = errors.New("config: invalid provider")

	// ErrInvalidOrphanRule reports an orphan rule whose pattern will not
	// compile, or which asks for pid liveness without capturing a pid.
	ErrInvalidOrphanRule = errors.New("config: invalid orphan rule")

	// ErrInvalidSize reports a size value that is not a byte count with an
	// optional unit suffix.
	ErrInvalidSize = errors.New("config: invalid size")

	// ErrInvalidDuration reports a duration value that is neither a Go
	// duration nor a day/week suffixed count.
	ErrInvalidDuration = errors.New("config: invalid duration")

	// ErrUnknownCategory reports a --categories value outside the known set.
	ErrUnknownCategory = errors.New("config: unknown category")
)

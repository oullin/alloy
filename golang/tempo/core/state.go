// Package core defines the shared value primitives every temporal type in
// tempo is built from: the State snapshot of a moment in time and the
// Bearer interface that both immutable and mutable variants implement so
// feature packages can stay generic.
package core

import "time"

// State is the minimal snapshot a temporal value carries: an instant in UTC
// plus the location used to project it. Runtime concerns (locale, translator,
// settings) live on the higher-level type, not here, so kernel-level
// operations stay pure.
type State struct {
	Value    time.Time
	Location *time.Location
}

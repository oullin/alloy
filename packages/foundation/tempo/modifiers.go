package tempo

import "regexp"

var (
	modifierPattern = regexp.MustCompile(`^([+-]?\d+(?:\.\d+)?)\s*(milliseconds?|seconds?|minutes?|hours?|days?|weeks?|months?|quarters?|years?|decades?|centuries?|millenniums?)$`)
	movePattern     = regexp.MustCompile(`^(next|last|previous)\s+(milliseconds?|seconds?|minutes?|hours?|days?|weeks?|months?|quarters?|years?|decades?|centuries?|millenniums?)$`)
)

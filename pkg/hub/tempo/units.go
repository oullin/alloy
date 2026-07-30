package tempo

import (
	"time"

	"hara.sh/alloy/tempo/duration"
)

func normalizeUnit(unit Unit) Unit {
	return duration.NormalizeUnit(unit)
}

func fixedUnitDuration(unit Unit) (time.Duration, bool) {
	return duration.FixedUnitDuration(unit)
}

func unitDuration(unit Unit) time.Duration {
	return duration.UnitDuration(unit)
}

func bestRelativeUnit(milliseconds int64) Unit {
	return duration.BestRelativeUnit(milliseconds)
}

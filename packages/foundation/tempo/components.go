package tempo

import (
	"time"

	"github.com/oullin/alloy/packages/foundation/tempo/factory"
)

func timeFromComponents(components Components, location *time.Location) time.Time {
	return factory.TimeFromComponents(components, location)
}

func componentsMatchTime(components Components, value time.Time, location *time.Location) bool {
	return factory.ComponentsMatchTime(components, value, location)
}

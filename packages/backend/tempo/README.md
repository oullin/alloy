# Time Go

Time exposes a root package (`alloy.dev/backend/tempo`) that wires
methods onto `Time` / `*MutableTime` and a set of feature packages
that hold reusable mechanics behind a generic `core.Bearer[T]` contract.
Every feature package — `arithmetic`, `boundaries`, `comparison`,
`setters`, `diff`, `formatting` — exports plain functions you can call
on any bearer:

```go
package examples

import (
	"time"

	"alloy.dev/backend/tempo/arithmetic"
	"alloy.dev/backend/tempo/boundaries"
	"alloy.dev/backend/tempo/formatting"
	tempo "alloy.dev/backend/tempo"
)

func DateOnly(value tempo.Time) string {
	return formatting.DateString(value)
}

func NextBusinessDay(value tempo.Time) tempo.Time {
	return boundaries.NextWeekday(value, []time.Weekday{time.Saturday, time.Sunday})
}

func ShiftDays(value *tempo.MutableTime, days int) *tempo.MutableTime {
	return arithmetic.AddDays(value, days)
}
```

Locale and translation behavior is composed with runtimes and factories:

```go
import (
	"alloy.dev/backend/tempo/runtime"
	tempo "alloy.dev/backend/tempo"
)

rt := runtime.New(
	runtime.Locale("en-US"),
	runtime.WithTranslator(myTranslator),
)

factory, err := tempo.NewFactory(
	tempo.WithRuntime(rt),
	tempo.WithTimezone("UTC"),
)
if err != nil {
	return err
}

value, err := factory.Parse("2024-05-15")
```

## Package layout

- `core/` – the `State` snapshot and `Bearer[T]` contract feature
  packages target.
- `internal/kernel/` – pure `time.Time` math the feature packages
  build on.
- `duration/`, `calendar/` – calendar primitives and `Unit`
  definitions.
- `arithmetic/`, `boundaries/`, `comparison/`, `setters/`, `diff/`,
  `formatting/` – generic Bearer-based feature implementations.
- `interval/`, `period/` – the time-span helpers used by
  `tempo.Interval` and `tempo.Period`.
- `runtime/` – `Context`, `Translator`, locale/fallback/translator
  options.
- `factory/`, `parser/`, `config/` – lower-level construction and
  configuration mechanics.
- module root – the application entrypoint that wires methods onto
  `Time` and `*MutableTime`, exposes `Parse`, `NewFactory`,
  `NewRuntime`, and the public value/configuration surface.

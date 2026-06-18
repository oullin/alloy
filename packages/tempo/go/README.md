# Tempo Go

Tempo exposes a root package (`github.com/oullin/alloy/tempo`) that wires
methods onto `Tempo` / `*MutableTempo` and a set of feature packages
that hold reusable mechanics behind a generic `core.Bearer[T]` contract.
Every feature package — `arithmetic`, `boundaries`, `comparison`,
`setters`, `diff`, `formatting` — exports plain functions you can call
on any bearer:

```go
package examples

import (
	"time"

	"github.com/oullin/alloy/tempo/arithmetic"
	"github.com/oullin/alloy/tempo/boundaries"
	"github.com/oullin/alloy/tempo/formatting"
	tempo "github.com/oullin/alloy/tempo"
)

func DateOnly(value tempo.Tempo) string {
	return formatting.DateString(value)
}

func NextBusinessDay(value tempo.Tempo) tempo.Tempo {
	return boundaries.NextWeekday(value, []time.Weekday{time.Saturday, time.Sunday})
}

func ShiftDays(value *tempo.MutableTempo, days int) *tempo.MutableTempo {
	return arithmetic.AddDays(value, days)
}
```

Locale and translation behavior is composed with runtimes and factories:

```go
import (
	"github.com/oullin/alloy/tempo/runtime"
	tempo "github.com/oullin/alloy/tempo"
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
- `runtime/` – `Runtime`, `Translator`, locale/fallback/translator
  options.
- `factory/`, `parser/`, `config/` – lower-level construction and
  configuration mechanics.
- module root – the application entrypoint that wires methods onto
  `Tempo` and `*MutableTempo`, exposes `Parse`, `NewFactory`,
  `NewRuntime`, and the public value/configuration surface.

# Tempo Go

Tempo extension points are plain Go functions or user-owned wrapper types. Keep
reusable behavior outside the package surface:

```go
package examples

import tempo "github.com/oullin/alloy/tempo/tempo"
import "github.com/oullin/alloy/tempo/arithmetic"
import "github.com/oullin/alloy/tempo/formatting"

func DateOnly(value tempo.Tempo) string {
	return formatting.From(value).DateString()
}

func NextBusinessDay(value tempo.Tempo) tempo.Tempo {
	next := arithmetic.From(value).AddDays(1).Tempo()
	for next.IsWeekend() {
		next = arithmetic.From(next).AddDays(1).Tempo()
	}

	return next
}
```

Locale and translation behavior is composed with runtimes and factories:

```go
runtime := tempo.NewRuntime(
	tempo.RuntimeLocale("en-US"),
	tempo.RuntimeTranslator(myTranslator),
)

factory, err := tempo.NewFactory(
	tempo.WithRuntime(runtime),
	tempo.WithTimezone("UTC"),
)
if err != nil {
	return err
}

value, err := factory.Parse("2024-05-15")
```

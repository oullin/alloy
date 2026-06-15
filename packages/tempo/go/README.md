# Tempo Go

Tempo extension points are plain Go functions or user-owned wrapper types. Keep
reusable behavior outside the package surface:

```go
package examples

import tempo "github.com/oullin/alloy/tempo/tempo"

func DateOnly(value tempo.Tempo) string {
	return value.DateString()
}

func NextBusinessDay(value tempo.Tempo) tempo.Tempo {
	next := value.AddDays(1)
	for next.IsWeekend() {
		next = next.AddDays(1)
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

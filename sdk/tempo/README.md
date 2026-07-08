# Tempo TypeScript

The TypeScript package lives in `sdk/tempo` and publishes as
`@alloy/sdk/tempo`. Acceptance tests for this package live in
`sdk/tempo/tests` and expose the `test:tempo` script.

Tempo extension points are plain composables. Keep reusable behavior in normal
functions that accept and return Tempo values:

```ts
import { Tempo, type TempoImmutable } from "@alloy/sdk/tempo";

export const dateOnly = (value: TempoImmutable): string => value.toDateString();

export const nextBusinessDay = (value: TempoImmutable): TempoImmutable => {
  let next = value.addDays(1);

  while (next.isWeekend()) {
    next = next.addDays(1);
  }

  return next;
};

dateOnly(Tempo.parse("2024-05-15"));
```

Parsing and component construction are strict by default:

```ts
Tempo.create({ year: 2024, month: 2, day: 29 });
Tempo.createNormalized({ year: 2024, month: 2, day: 31 });
```

Locale and translation behavior is composed with runtimes and factories:

```ts
import { TempoFactory, createTempoRuntime } from "@alloy/sdk/tempo";

const runtime = createTempoRuntime({
  locale: "en-US",
  translator: {
    getMessage: (key) => (key === "greeting" ? "Hello :name" : null),
  },
});

const factory = TempoFactory.create({ runtime, timeZone: "UTC" });

factory.parse("2024-05-15").translate("greeting", { name: "Tempo" });
```

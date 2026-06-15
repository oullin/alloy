# Tempo TypeScript

Tempo extension points are plain composables. Keep reusable behavior in normal
functions that accept and return Tempo values:

```ts
import { Tempo, type TempoImmutable } from "@alloy/tempo";

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

Locale and translation behavior is composed with runtimes and factories:

```ts
import { TempoFactory, createTempoRuntime } from "@alloy/tempo";

const runtime = createTempoRuntime({
  locale: "en-US",
  translator: {
    getMessage: (key) => (key === "greeting" ? "Hello :name" : null),
  },
});

const factory = TempoFactory.create({ runtime, timeZone: "UTC" });

factory.parse("2024-05-15").translate("greeting", { name: "Tempo" });
```

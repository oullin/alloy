# Cross-runtime conformance fixtures

This directory is the **single source of cross-runtime truth** for the money and
tempo twins. `pkg/hub/money` (Go) and `@hara/sdk-money` (TS), and
`pkg/hub/tempo` (Go) and `@hara/sdk-tempo` (TS), are behavioral mirrors. The
JSON files here encode language-neutral golden vectors that **both** runtimes
execute, so any divergence fails CI instead of surfacing as a production
discrepancy between backend and frontend. This is the guard that raises the twins
from parity level **L1 (asserted)** to **L2 (enforced)** — see
`web/docs/architecture/parity.md`.

## Files

| Fixture      | Go loader                                 | TS loader                                 |
| ------------ | ----------------------------------------- | ----------------------------------------- |
| `money.json` | `pkg/hub/money/money/conformance_test.go` | `sdk/money/tests/src/conformance.test.ts` |
| `tempo.json` | `pkg/hub/tempo/conformance_test.go`       | `sdk/tempo/tests/src/conformance.test.ts` |

Both loaders read the JSON here by a path relative to their own source file, so
the fixtures live outside every package and neither runtime owns them.

## Format

Every numeric input and output is a **string** so JSON float precision never
enters the comparison. Error cases are matched by **error identity** (the Go
`exception` sentinel via `errors.Is`; the TS `MoneyError.code`), never by message
text.

### `money.json`

```json
{ "op": "round",  "args": ["1250", "2"],       "expected": "1300",  "note": "..." }
{ "op": "add",    "args": ["9223372036854775807", "1"], "error": "ERR_OVERFLOW", "note": "..." }
```

- `op`: `round` | `absolute` | `add` | `subtract` | `multiply` |
  `createFromFloat` | `convertWithRate` | `avg` | `unmarshalAmount` |
  `unmarshalCurrency` | `resolveWithDefault`.
- `args`: operands as decimal strings. The two `unmarshal*` ops are the
  exception: each takes a whole JSON payload as its single arg, decoded by each
  runtime's money unmarshaller (Go `json.Unmarshal` into `Value` / TS
  `MoneyJson.unmarshal`).
  - `unmarshalAmount` expects the resulting minor-unit amount as a decimal
    string. It pins that both runtimes accept a **quoted** amount as readily as
    a bare one, which matters because Go types the field as `json.Number` and TS
    `toJSON()` emits the quoted form so an int64 survives a JavaScript
    consumer's `JSON.parse`.
  - `unmarshalCurrency` expects the resulting ISO code. It pins the **default
    currency**, which is shared state neither runtime can change alone: an
    absent or empty `currency` falls back to the provider default (Go
    `currency.DefaultProvider` / TS `DefaultCurrencyProvider`, both **SGD**),
    while an unknown code is an error rather than a silent fallback.

- `resolveWithDefault` takes **two** args — a currency code and a code to look
  up — and expects the resulting ISO code. Each loader builds a fresh manager,
  changes its default (Go `(*currency.Manager).SetDefault` / TS
  `CurrencyManager.setDefault`), and resolves the second code. It pins that the
  setter normalises identically, that a known code still resolves to itself, and
  that an unknown code is rejected rather than adopted. The default is per
  manager, so nothing leaks between cases.

  For `createFromFloat` and
  `convertWithRate`, float-valued args (a major-unit amount, a rate) are decimal
  strings parsed to IEEE-754 doubles identically in both runtimes. For `avg`,
  the first arg is the currency code and the rest are minor-unit amounts; the
  loaders build currency-carrying values (Go `Manager.Create` /
  TS `MoneyManager.create`) and run them through the aggregator
  (Go `Aggregator.Avg` / TS `MoneyAggregator.avg`).
- Exactly one of `expected` (decimal-string result) or `error` (error identity)
  is present.

`add`/`subtract`/`multiply` map to the **error-propagating** variants in each
runtime (Go `Engine.SafeAdd`/`SafeSubtract`/`SafeMultiply`; TS
`MoneyCalculator.add`/`subtract`/`multiply`), so an overflow surfaces as
`ERR_OVERFLOW` in both.

### `tempo.json`

```json
{ "op": "addDays", "base": { "year": 2024, "month": 3, "day": 10, "hour": 0, "timeZone": "America/New_York" }, "arg": 1, "render": "iso", "expected": "2024-03-11T04:00:00.000Z", "note": "..." }
```

- `op`: `addDays` | `addWeeks` | `addHours` | `addMonths` |
  `addMonthsNoOverflow` | `diffInMonths` | `diffInYears` | `parseFromPattern`.
- `base` / `other`: calendar components (`year`, `month`, `day`, optional
  `hour`/`minute`/`second`, `timeZone`) used to build an instant in a named zone.
- `arg`: the integer operand for the `add*` operations.
- `input` / `pattern` / `timeZone`: inputs for `parseFromPattern`.
- `render`: how the result is stringified — `iso` (UTC ISO-8601 with
  milliseconds), `date` (`YYYY-MM-DD`). Diff ops render their integer directly.

## Adding a case

Conformance is a **contract**, not a convenience:

1. Add the behavior to **both** runtimes.
2. Add a case to the relevant fixture here.
3. Confirm both loaders execute it (`pkg/hub` Go suite + the `sdk/*` TS suites).

A new shared behavior is not "done" until it has a fixture. A divergence found in
either runtime must be recorded in the divergence register in
`web/docs/architecture/parity.md` (dropping the twin to L0) and closed only when
both runtimes agree **and** a fixture pins the agreement.

## Deliberate exclusions

These shared-surface behaviors are **intentionally not** encoded because the two
runtimes do not currently converge on them. Each is a genuine divergence or a
capability gap, not an oversight; encoding a single expected value would just
force one runtime to fail. They belong in the parity divergence register, not
here, until resolved in both runtimes at once.

- **`tempo` fractional-day / fractional-week additions.** TS `addDays`/`addWeeks`
  accept fractional amounts (e.g. `addDays(1.25)` = one calendar day + 6 elapsed
  hours). Go `AddDays`/`AddWeeks` take an `int` and cannot express a fraction.
  Only whole-unit additions are a shared surface.
- **`tempo` non-English-locale `parseFromPattern`.** TS `Tempo.fromFormat`
  accepts a `locale` and parses localized month names (`14 juillet 2026`). Go
  `parser.ParseFromPattern` has no locale parameter and a hard-coded English
  month map (`pkg/hub/tempo/parser/month.go`). Only English month-name parsing is
  shared; the fixture covers that.
- **JSON amounts beyond the int64 range.** Go's unmarshaller rejects
  `{"amount":"9223372036854775808"}` with `ErrInvalidJSONUnmarshal`, because the
  decoded value must fit an `int64`. TS decodes it to a `bigint`, which has no
  such ceiling, so it succeeds and yields a value the calculator's own
  `MAX_INT64` guard would later reject. This is a genuine divergence: encoding a
  single expectation would just force one runtime to fail. It belongs in the
  parity register until TS bounds the decoded amount to int64.
- **`workflow` `RetryPolicy.MaxExceptions` (X13).** Declared in both runtimes,
  enforced in neither, documented differently. It is an unresolved L0 divergence
  (parity register X13); no conformance case exists until it is defined and
  implemented in both runtimes together.

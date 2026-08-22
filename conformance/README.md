# Cross-runtime conformance fixtures

This directory is the **single source of cross-runtime truth** for the money,
tempo, and synchronous container twins. `pkg/hub/money` (Go) and
`@hara/sdk-money` (TS), `pkg/hub/tempo` (Go) and `@hara/sdk-tempo` (TS), and
`pkg/hub/container` (Go) and `@hara/sdk-container` (TS) are behavioral mirrors.
The JSON files here encode language-neutral golden vectors that **both** runtimes
execute, so any divergence fails CI instead of surfacing as a production
discrepancy between backend and frontend. This is the guard that raises the twins
from parity level **L1 (asserted)** to **L2 (enforced)** — see
`web/docs/architecture/parity.md`.

## Files

| Fixture          | Go loader                                 | TS loader                                 |
| ---------------- | ----------------------------------------- | ----------------------------------------- |
| `money.json`     | `pkg/hub/money/money/conformance_test.go` | `sdk/money/tests/src/conformance.test.ts` |
| `tempo.json`     | `pkg/hub/tempo/conformance_test.go`       | `sdk/tempo/tests/src/conformance.test.ts` |
| `container.json` | `pkg/hub/container/conformance_test.go`   | `sdk/container/tests/conformance.test.ts` |

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
  `createFromFloat` | `convertWithRate` | `avg` | `isSafeAsNumber` |
  `unmarshalAmount` | `unmarshalCurrency` | `resolveWithDefault` |
  `displayCompact` | `formatWhole` | `formatCompactSignificant`.
  The two `format*` ops answer a display string and take the currency code as
  their second arg, since layout is per-currency; `formatCompactSignificant`
  takes the significant-digit count as its third, clamped to `[1, 6]` — the
  upper bound is what keeps the Go twin's int64 arithmetic provably safe.
  `isSafeAsNumber` answers `"true"` or `"false"` rather than an amount: it asks
  whether the minor-unit figure survives conversion to an IEEE-754 double, which
  is the one place the twins' number types meet.
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

- `displayCompact` takes a currency code and a minor-unit amount, and expects
  the abbreviated display string. It pins the scale choice, the half-away
  rounding that can roll a value into the next scale, whether a decimal is kept,
  and where the suffix lands relative to a leading or trailing grapheme — all of
  which are currency-driven and easy to drift between runtimes.
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

### `container.json`

```json
{
	"schemaVersion": 1,
	"cases": [
		{
			"id": "lifetimes-and-scoped-reset",
			"note": "...",
			"tokens": [{ "id": "service" }],
			"operations": [
				{ "kind": "bind", "token": "service", "lifetime": "singleton", "primitive": "constant", "value": "v" },
				{ "kind": "resolve", "token": "service", "observe": "value" }
			],
			"expected": ["value=v"]
		}
	]
}
```

- This is a restricted declarative DSL, not a scenario-name selector. Every
  case declares its tokens, optional providers, ordered operations, and ordered
  observations. `tokens[].kind` is optional runtime metadata; `string` is the
  only currently shared enforceable kind.
- Operations cover bind/`bind-if`/`singleton-if`/`scoped-if`, instance, resolve,
  resolve-with-parameters, get, forget-scoped/instance, flush, alias, contextual
  value/factory/tagged bindings (including multi-concrete `consumers`),
  tag/tagged, extend, callbacks, rebinding, method bind/call, call/wrap/
  factory-func, provider register/register-many/boot, and introspection
  observations (`bound`/`has`/`resolved`/`isShared`/`bindings`, providers/
  hasProvider/providerFor/booted, counters/events). Observations compare in
  execution order.
- Factories, method callables, and extenders select a documented finite
  primitive registry: `constant`, `increment-counter`, `resolve-token`,
  `read-parameter`, `append-suffix`, and `return-instance`. Shared values use
  the JSON scalar union `string | number | boolean | null` with matching
  renderings (`null` → `<nil>`). JSON never embeds language expressions,
  constructors, or executable source.
- Exactly one of `expected` (ordered observations) or `error` is present.
  Conformance error identities map to the Go sentinel and the TypeScript typed
  error code respectively: `ALIAS_CYCLE` → `ErrAliasCycle` / `AliasCycleError`,
  `SELF_ALIAS` → `ErrSelfAlias` / `SelfAliasError`, `MISSING_BINDING` →
  `ErrNotBound` / `MissingBindingError`, `MISSING_METHOD_BINDING` →
  `ErrMethodNotBound` / `MissingMethodBindingError`, `CIRCULAR_RESOLUTION` →
  `ErrCircularDependency` / `CircularResolutionError`, and `PROVIDER_CYCLE` →
  `ErrProviderCycle` / `ProviderCycleError`. Consumers must compare these
  stable identities (`ContainerError.code` in TypeScript), never error message
  text.
- Both loaders validate the version, IDs, note, required `tokens`/`operations`
  arrays, lifetime (`transient`/`singleton`/`scoped` or absent), providers/
  primitives, operation fields and references, error identity, and
  expected/error exclusivity before executing a case. Contextual operations
  reject ambiguous value/factory and consumer/consumers declarations.

The shared vectors cover the synchronous common core: transient/singleton and
Go-style scoped-cache reset; parameter reset; scalar values; aliases including
self-alias and cycles; contextual resolution including multi-concrete and
GiveTagged; tag ordering; extenders; lifecycle and rebinding callbacks; method
bindings; Call/Wrap/FactoryFunc; conditional bindings; introspection queries;
missing/circular/method/provider-cycle errors; flush reset of aliases, tags,
extenders, callbacks, methods, and resolved state; provider ordering,
deduplication, introspection (including ProviderFor first-match), deferred
make/get, deferred register reentrancy, and boot idempotency. TypeScript
`childScope`, async resolution/hooks, promise singleflight, raw
`App.Make` bypass of deferred providers, and GiveConfig (typed config token +
getter in TS vs magic `"config"` string in Go) are intentional exclusions.

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
- **Container asynchronous APIs.** `@hara/sdk-container` adds `makeAsync`,
  async hooks, and promise singleflight. Go is the synchronous parity source
  and has no equivalent API, so async behavior stays in TypeScript tests rather
  than `container.json`.
- **Container child scopes.** `@hara/sdk-container` adds `childScope()` with
  isolated scoped values. The shared `scoped` vector instead pins the Go common
  behavior: `ForgetScopedInstances()` clears the scoped cache in the same
  container. Child scopes remain a documented TypeScript-only extension.
- **Container raw App deferred bypass.** Go `Application` embeds `*App`, so
  callers can invoke `App.Make` and skip deferred provider flush. TypeScript
  `Application` overrides `make`/`get` with no public raw-container escape
  hatch, so that boundary is not encoded as a shared vector.
- **Container GiveConfig.** Go resolves a magic `"config"` abstract with a
  `Get(key, fallback...)` duck type. TypeScript requires an explicit typed
  config token plus getter (`giveConfig(config, getter, fallback?)`) because
  token identity is structural/branded rather than a shared string key. No
  language-neutral GiveConfig vector exists until that boundary converges.

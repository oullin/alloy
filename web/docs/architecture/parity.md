# Cross-Runtime Parity Matrix & Twin Policy

Alloy's value proposition is *cross-runtime primitives*: a small set of packages
whose logic must produce identical results whether it runs on the Go backend or
in a TypeScript frontend. Most packages are server-only and have no frontend
meaning; a deliberate few are **twins** — a Go package and a `@alloy/sdk/*`
package that are maintained as behavioral mirrors of one another.

This page is the authoritative map of which primitive exists in which runtime,
how much parity each twin actually guarantees, where the runtimes currently
diverge, and the policy for what earns a twin in the first place. It exists so
consumers can tell — without reading source — whether the primitive they need is
available in their runtime, and so contributors know the rules before adding or
twinning a package.

> Surveyed against `origin/main`. Every classification and parity claim below is
> verified against source (files cited) or explicitly marked **pending PR merge**
> where a fix is in flight but not yet on `main`.

## The matrix

Every Go package under `pkg/hub` and every TypeScript package under `sdk` is
listed and classified **Go-only**, **TS-only**, or **Both (twin)**. Nested Go
modules (`auth/passkeys`, `queue/drivers/sqs`) are listed under their parents.

| Package | Go (`pkg/hub`) | TS (`sdk`) | Classification | Conformance coverage |
| --- | :---: | :---: | --- | --- |
| `money` | ✅ | ✅ `@alloy/sdk/money` | **Both (twin)** | None yet — see [plan 008](#conformance-the-guard-that-keeps-both-honest) |
| `tempo` | ✅ | ✅ `@alloy/sdk/tempo` | **Both (twin)** | None yet — plan 008 |
| `workflow` | ✅ | ✅ `@alloy/sdk/workflow` | **Both (twin)** | None yet — plan 008 |
| `console` | — | ✅ `@alloy/sdk/console` | **TS-only** | n/a (no backend meaning) |
| `navigator-routes` | — | ✅ `@alloy/sdk/navigator-routes` | **TS-only** | n/a (consumes a route manifest) |
| `auth` (+ `auth/passkeys`) | ✅ | — | **Go-only** | n/a |
| `bus` | ✅ | — | **Go-only** | n/a |
| `cache` | ✅ | — | **Go-only** | n/a |
| `collection` | ✅ | — | **Go-only** | n/a |
| `config` | ✅ | — | **Go-only** | n/a |
| `container` | ✅ | — | **Go-only** | n/a |
| `contracts` | ✅ | — | **Go-only** | n/a |
| `cookie` | ✅ | — | **Go-only** | n/a |
| `database` | ✅ | — | **Go-only** | n/a |
| `encryption` | ✅ | — | **Go-only** | n/a |
| `events` | ✅ | — | **Go-only** | n/a |
| `filesystem` | ✅ | — | **Go-only** | n/a |
| `hashing` | ✅ | — | **Go-only** | n/a |
| `httpx` | ✅ | — | **Go-only** | n/a |
| `inertia` | ✅ | — | **Go-only** | n/a |
| `queue` (+ `queue/drivers/sqs`) | ✅ | — | **Go-only** | n/a |
| `seo` | ✅ | — | **Go-only** | n/a |
| `session` | ✅ | — | **Go-only** | n/a |
| `str` | ✅ | — | **Go-only** | n/a |
| `validation` | ✅ | — | **Go-only** | n/a |

**Summary:** 3 twins (`money`, `tempo`, `workflow`), 20 Go-only packages (plus 2
nested Go modules), 2 TS-only packages. No twin currently has mechanical
conformance coverage; parity is asserted, not enforced (see the divergence
register and policy below for why that matters).

### Why the TS-only packages have no Go twin

- **`console`** is a terminal-UI toolkit (async prompts, spinners, tables) for
  Node CLIs (`sdk/console/README.md`). It is a build/DX tool, not an application
  primitive — there is no "identical result on both runtimes" to guarantee, so a
  Go twin would be meaningless.
- **`navigator-routes`** resolves named routes to URLs from a manifest of
  Laravel-style patterns (`sdk/navigator-routes/README.md`). It is the frontend
  *consumer* of route definitions the Go side produces (e.g. `httpx`/`inertia`
  route descriptors); the manifest is the contract, not a shared algorithm. It
  is TS-only by design, not by omission.

## Twin parity detail

Classification says a twin *exists*; it does not say the two implementations
*agree*. This section records, per twin, the guaranteed parity level and the
concrete divergences found on `origin/main`.

Parity levels are defined in the [policy](#parity-levels). In short: **L2** =
mechanically enforced by shared fixtures; **L1** = asserted parity with matching
tests but no shared oracle; **L0** = a shared surface with known, unresolved
divergence.

### `money` ↔ `@alloy/sdk/money` — currently **L1**, divergences in flight

Both runtimes model amounts as 64-bit minor units (Go `int64`; TS `bigint`
range-checked to int64) and expose matching calculator/currency/exchange/parser
surfaces. Two behaviors are actively being unified by in-flight PRs; **on
`origin/main` they are not yet unified**:

- **Rounding tie-direction.** Both `Engine.Round`
  (`pkg/hub/money/calculator/calculator.go:120`) and `Calculator.round`
  (`sdk/money/src/calculator.ts:106`) currently round with `module > reminder/2`,
  i.e. **half toward zero** — the Go comment says so explicitly ("ties rounded
  down (toward zero)", `calculator.go:118`). The two runtimes therefore *agree*
  numerically today, but the TS docstring already claims "half away from zero"
  (`calculator.ts:95`), so the TS doc is ahead of its own implementation.
  **Pending PR merge (#88):** unify to **half-away-from-zero** with shared
  vectors `1250 → 1300`, `-1250 → -1300`, `250 → 300`.
- **`Absolute(MinInt64)`.** Go's guard `amount < math.MinInt64`
  (`calculator.go:69`) is dead code — an `int64` can never be below `MinInt64` —
  so `Absolute(MinInt64)` overflows to `MinInt64` (still negative). TS
  `absolute` (`sdk/money/src/calculator.ts:79`) returns `2^63` via `bigint`,
  which is *outside* the int64 range. The runtimes **diverge here today.**
  **Pending PR merge (#88):** both return `0` for `Absolute(MinInt64)`.
- **Exchange conversion.** Both `Rates.ConvertAmount`
  (`pkg/hub/money/exchange/exchange.go:113`, `float64(amount)/pow … math.Round`)
  and `ExchangeRates.convertAmountWithRate`
  (`sdk/money/src/exchange/rates.ts:71`, `Number(amount)/… roundAwayFromZero`)
  convert through IEEE-754 doubles today, so both lose precision above `2^53`.
  **Pending PR merge (#90):** exact-integer conversion in both runtimes, with
  vector `(2^53 + 1) × 2.0`.

Once #88/#90 land, `money` reaches **L1** cleanly and is eligible for **L2** as
soon as plan 008's fixtures are added.

### `tempo` ↔ `@alloy/sdk/tempo` — currently **L0** for day/week, fix in flight

Tempo is described in the README as "the most complete cross-runtime package",
but calendar arithmetic diverges on `origin/main`:

- **DST-correct day/week arithmetic.** Go routes `AddDays`/`AddWeeks` through
  `kernel.Add`, which uses `addDate` → `time.AddDate` for `duration.Day`/`Week`
  (`pkg/hub/tempo/internal/kernel/arithmetic.go:25-28`, `:87`) — calendar-correct,
  so a day across a DST boundary is 23 or 25 wall-clock hours. TS treats `day`
  and `week` as **fixed milliseconds** (`fixedUnitMilliseconds` returns
  `millisecondsPerDay`/`millisecondsPerWeek`,
  `sdk/tempo/src/calendar/index.ts:314-318`, consumed by `add`,
  `sdk/tempo/src/core/index.ts:1059-1066`) — pure UTC-offset math that shifts the
  local hour across a DST transition. The runtimes **diverge here today.**
  **Pending PR merge (#87):** TS day/week become DST-correct, matching Go's
  `AddDate` semantics; the `addCalendarDays ↔ addDate` mappings are recorded so
  the equivalence is explicit.
- Fixed-duration units (`ms`/`s`/`min`/`hour`) already agree: both do offset math
  (`arithmetic.go:23-24`; `fixedUnitMilliseconds`), and month/quarter/year both
  route through calendar logic (Go `kernel.Add`; TS `addMonths`/`addYears`).

Once #87 lands, `tempo` day/week reaches **L1**; **L2** follows with plan 008.

### `workflow` ↔ `@alloy/sdk/workflow` — **L0**, unresolved divergence (X13)

The Petri-net engine, definition builder, marking, and multi-step runner exist
in both runtimes with matching vocabulary. One retry-policy field is a genuine,
**unresolved** divergence that no in-flight PR fixes — tracked as follow-up
**X13**:

- **`RetryPolicy.MaxExceptions` is declared in both runtimes and enforced in
  neither.** Go declares `MaxExceptions` with an aspirational doc — "caps the
  number of unique error types before giving up"
  (`pkg/hub/workflow/multisteps/retry.go:18-20`) — but `runWithRetry`
  (`retry.go:37-88`) never reads it; only `MaxTries`, `Backoff`, and `Timeout`
  drive the loop. TS mirrors the field
  (`sdk/workflow/src/multisteps/retry.ts:5,11`) and copies it through
  `withRetryPolicy` (`sdk/workflow/src/multisteps/jobs.ts:52`), but its
  `runWithRetry` (`retry.ts:23-58`) also never reads `maxExceptions`.
- **Net state:** the field is part of the public surface in both runtimes, is
  documented differently (Go: "unique error types"; the intended TS semantics
  per the campaign are "total exceptions per job"), and is a **no-op in both**.
  This is worse than a silent divergence — it is a documented capability that
  does not exist, in a way that will diverge the moment either runtime
  implements it. It needs a policy decision (see
  [Recording & resolving divergence](#recording--resolving-divergence)) before
  either runtime wires it up, so both converge on one definition and a shared
  fixture lands with the implementation.

> **Note for reviewers:** the campaign brief characterized TS as *enforcing*
> `maxExceptions` total-per-job while Go left it unenforced. On `origin/main`
> that is not the case — **neither** runtime enforces it. The divergence is
> therefore about the *contract* (declared-but-unimplemented, with mismatched
> docs), and the fix must land in both runtimes plus a conformance case at once.

## Divergence register

| ID | Twin | Divergence | Status |
| --- | --- | --- | --- |
| — | `money` | Rounding tie-direction (half-toward-zero today; TS doc says away-from-zero) | Pending PR #88 → half-away-from-zero, both |
| — | `money` | `Absolute(MinInt64)`: Go → `MinInt64`, TS → `2^63` | Pending PR #88 → `0`, both |
| — | `money` | Exchange conversion via float64/Number (lossy > 2^53) | Pending PR #90 → exact-integer, both |
| — | `tempo` | Day/week: Go calendar (`AddDate`), TS fixed-ms (DST-wrong) | Pending PR #87 → TS DST-correct |
| **X13** | `workflow` | `RetryPolicy.MaxExceptions` declared in both, enforced in neither; docs disagree | **Open** — needs policy + dual implementation + fixture |

## Policy: what earns a twin, and what a twin guarantees

### What earns a TS twin

A package gets a `@alloy/sdk/*` twin only when **both** of these hold:

1. **Shared-result logic.** The package computes a value that must be *identical*
   on backend and frontend for the product to be correct — money math, date/time
   arithmetic, workflow state transitions. If a frontend that reimplemented the
   logic could disagree with the backend and cause a user-visible bug, the logic
   belongs in a twin. This is why `money`, `tempo`, and `workflow` are twins.
2. **No runtime-specific dependency.** The logic must be expressible against
   pure language primitives in both runtimes (integers, strings, the platform
   date object) without a server-only dependency (a database handle, a socket, a
   filesystem, an OS signal).

A package is **explicitly not** a twin candidate when it is a server-only
concern with no frontend meaning — `queue`, `container`, `httpx`, `session`,
`database`, `cache`, `bus`, `auth`, etc. These have no "identical result on both
runtimes" to guarantee; a twin would be dead weight and a second surface to keep
in sync for no benefit. Frontend-only tooling (`console`) and
manifest-consumers (`navigator-routes`) are the mirror case: TS-only by the same
test.

**Default to no twin.** Twins double the maintenance surface and, without
conformance coverage, double the *divergence* surface (see every row of the
register above). Add one only when rule 1 is clearly met, and prefer a **narrow
twin** — the minimum surface that carries the shared logic — over a full port.
This matches the `authkit` design spike (`docs/design/authkit-api.md`, plan
025), which recommends shipping auth **Go-first** with only a *narrow* `sdk/*`
twin limited to browser-side ceremony helpers (passkeys, flow-step client),
never a full port of the server composition layer. When a package's shared logic
is a thin slice of a mostly-server-only concern, twin **only that slice.**

### Parity levels

Every twin declares a parity level so consumers know how much the guarantee is
worth:

- **L2 — Enforced.** The twin's behavior is pinned by shared, language-neutral
  conformance fixtures (plan 008) that both the Go and TS suites execute. A
  divergence fails CI. This is the target state for `money` and `tempo`.
- **L1 — Asserted.** Both runtimes have matching unit tests and are believed to
  agree, but there is no shared oracle, so drift can ship undetected. This is
  where `money` and `tempo` land the moment #88/#90/#87 merge and before plan 008.
- **L0 — Known divergence.** A shared surface with at least one documented,
  unresolved behavioral difference. `workflow` is L0 today because of X13. An L0
  twin must carry a divergence-register entry and must not claim parity in its
  package docs.

A twin's level is not permanent: it is the current honest state, recorded here
and re-evaluated whenever the twin or its fixtures change.

### Recording & resolving divergence

1. **Record it here first.** Any known behavioral difference between a twin's two
   implementations gets a row in the divergence register with a stable ID
   (`X##`) if it is not covered by an in-flight PR. The twin drops to **L0**
   until resolved.
2. **Resolve in both runtimes plus a fixture, atomically.** A divergence is
   closed only when (a) both runtimes implement the *same* agreed behavior and
   (b) a conformance case (plan 008) encodes it so it cannot silently reopen.
   Fixing one runtime without the fixture just moves the divergence; fixing the
   doc without the code (as happened with money rounding) is worse — it hides it.
3. **No aspirational surface.** Do not ship a declared-but-unimplemented field on
   a twin (the X13 mistake). If a capability is not yet implemented in both
   runtimes, it is not part of the twin's public surface yet.

## Conformance: the guard that keeps "both" honest

Parity that is only *asserted* drifts — every row of the divergence register is a
case that shipped precisely because no cross-runtime guard existed. **Plan 008**
adds that guard: language-neutral JSON golden fixtures, executed by both a Go
loader test and a TS loader test, so any future divergence fails CI instead of
surfacing as a production discrepancy. Plan 008 is **deferred until the
money/tempo fixes (#87/#88/#90) merge** — writing the fixtures first would just
encode today's divergences.

When plan 008 lands, its fixtures **must encode**, at minimum:

- **money — rounding:** exact-half cases under the agreed policy, including
  `1250 → 1300`, `-1250 → -1300`, `250 → 300` (half-away-from-zero, #88).
- **money — absolute:** `Absolute(MinInt64) → 0` in both runtimes (#88).
- **money — exchange:** an exact-integer conversion above the float-safe range,
  `(2^53 + 1) × 2.0` (#90), plus an overflow marker case.
- **money — arithmetic:** `add`/`subtract`/`multiply` with an overflow → error
  marker, `createFromFloat`, and `avg`.
- **tempo — DST:** `addDays`/`addWeeks` across a DST boundary in a named zone,
  asserting the Go `AddDate` / TS `addCalendarDays` result (#87), plus
  `addMonths` month-end clamping and a non-English-locale `parseFromPattern`.
- **workflow — X13:** once `MaxExceptions` is defined and implemented in both
  runtimes, a case that exercises the exception cap so the two cannot diverge
  again. Until then, X13 stays in the register and `workflow` stays **L0**.

Outputs are rendered as **strings** and error cases matched by **error
identity** (not message text) so JSON float precision and wording never enter the
comparison.

## Recommendation: the next twin (grounded)

**The current three twins are the intended scope for now. Do not add a fourth
speculatively.** The one grounded near-term candidate is a *narrow* auth twin,
and only as a consequence of plan 025 — not as new scope this doc schedules.

Evidence:

- The only in-repo consumer, `web/inertia-demo/api/auth/*`, hand-rolls its
  auth/session logic instead of using `pkg/hub/auth`
  (`docs/design/authkit-api.md` §3.1). That is a demand signal for **`authkit`**
  (a Go composition facade), *not* for a TS port — the assembly burden is
  server-side.
- Where the frontend genuinely shares logic with the backend in auth, it is the
  **browser-side ceremony** only: WebAuthn passkey enrollment and flow-step
  progression. The plan 025 spike therefore recommends a *narrow* `sdk/*` twin
  limited to those helpers. That is the sole next-twin recommendation, and it is
  **gated on `authkit`/`authflows` shipping first** (plan 025), with tradeoffs
  owned there — this doc does not schedule the build.
- `billing` (on `@alloy/sdk/money`) is the obvious composition target *after*
  that, but it is out of scope here and not a twin recommendation.

No other package clears the bar: the remaining Go-only packages are server-only
concerns (rule 1 fails), and the TS-only packages have no backend meaning. If a
future frontend starts reimplementing money/tempo/workflow-shaped logic that a
twin already covers, that is a bug in the consumer, not a signal to add a twin.

## Maintenance

- Update the matrix whenever a package is added, removed, or twinned, and update
  a twin's **parity level** whenever its fixtures or behavior change.
- The plan-008 conformance link is what keeps a **Both** row honest; a twin
  without fixtures is at most **L1** and must say so.
- This page informs plan 025's decision on whether `authkit`/`authflows` get TS
  twins — keep the twin-scope policy and the plan-025 recommendation in sync.

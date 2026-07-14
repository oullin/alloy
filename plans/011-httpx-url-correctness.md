# Plan 011: Correct float route-parameter formatting and signed-URL query normalization

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- pkg/hub/httpx/routing`
> On change, reconcile excerpts; on mismatch, STOP.

## Status

- **Priority**: P2
- **Effort**: M
- **Risk**: MED
- **Depends on**: none
- **Category**: bug
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

Two URL-generation defects:
1. **C15**: `floatToString` returns `intToString(int64(f))`, so a `float64` route/query parameter (`3.14`, a price, a lat/long) is silently rendered as its integer truncation (`"3"`); large floats overflow int64 with undefined output. `stringify` also returns `""` for unhandled types, silently dropping values.
2. **C21**: Signed-URL verification (`HasCorrectSignature`) builds the HMAC input from the raw query string (only stripping `signature=`), while signing (`buildQuery`) sorts keys and `url.QueryEscape`s them. Because verify does not sort/re-encode, a valid signed URL whose params are reordered or re-encoded in transit (proxy, `+` vs `%20`) fails verification. Laravel canonicalizes on both sides.

## Current state

- `pkg/hub/httpx/routing/route_url_generator.go`:
  - `floatToString` (179-182): `return intToString(int64(f))`.
  - `stringify` (~124): dispatches by type; returns `""` for unknown.
  - `buildQuery` (101-122): sorts keys, `url.QueryEscape`.
- `pkg/hub/httpx/routing/url_generator.go`:
  - `HasCorrectSignature` (281-305): builds HMAC input from `request.QueryString()` minus `signature`, no sort/re-encode; uses `hmac.Equal` (constant-time — keep that).

Excerpt (`route_url_generator.go:179-182`):
```go
func floatToString(f float64) string {
	// Minimal float formatting; tests use integer parameters predominantly.
	return intToString(int64(f))
}
```

Convention: use `strconv` for numeric formatting; canonicalize query strings by parsing → sorting keys → re-encoding with the same escaping on both sign and verify. Keep `hmac.Equal` for the comparison.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Go routing tests | `cd pkg/hub && go test ./httpx/routing/...` | exit 0 |
| Full Go suite | `pnpm exec vp run go:test` | exit 0 |
| Format | `pnpm exec vp run format` | exit 0 |

## Scope

**In scope**: `pkg/hub/httpx/routing/route_url_generator.go`, `pkg/hub/httpx/routing/url_generator.go`, and their tests.

**Out of scope**: route matching/perf (plan 015); the HMAC key management; the `MatchableRequest` interface shape.

## Git workflow

- Branch: `advisor/011-httpx-url-correctness`; commit per finding; conventional-commit style.

## Steps

### Step 1 (C15): Format floats and unknown types correctly

Replace `floatToString` with `strconv.FormatFloat(f, 'f', -1, 64)` (shortest exact decimal). Change `stringify` to fall back to `fmt.Sprint(v)` for unhandled types instead of returning `""` (silently dropping a param is worse than a best-effort string). Preserve existing behavior for the common int/string cases.

**Verify**: `cd pkg/hub && go test ./httpx/routing -run 'Stringify|Float'` → `3.14` renders as `"3.14"`; an unknown type is not dropped.

### Step 2 (C21): Canonicalize the query on both sign and verify

Extract a single canonicalization function (parse the query, drop `signature`, sort keys, re-encode with the same escaping `buildQuery` uses) and call it from **both** `buildQuery` (signing) and `HasCorrectSignature` (verifying). The HMAC input must be identical for the same logical query regardless of incoming param order or `+`/`%20` differences. Keep `hmac.Equal`.

**Verify**: `go test ./httpx/routing -run Signature` → a signed URL still verifies after its query params are reordered and after `+`↔`%20` re-encoding; a tampered param fails.

### Step 3: Full suite + format

**Verify**: `pnpm exec vp run go:test` → exit 0; `pnpm exec vp run format` → exit 0.

## Test plan

- `route_url_generator_test.go`: float params (`3.14`, negative, large), unknown-type fallback.
- `url_generator_test.go`: signed URL verifies under reordered params and re-encoding; tampering (added/changed/removed param) fails; the constant-time compare is preserved.

## Done criteria

- [ ] `grep -n "int64(f)" pkg/hub/httpx/routing/route_url_generator.go` returns nothing in `floatToString`.
- [ ] A reordered-query signed URL verifies (test asserts it); a tampered one fails.
- [ ] `pnpm exec vp run go:test` exits 0; no out-of-scope files modified; `plans/README.md` row for 011 updated.

## STOP conditions

- Changing signing canonicalization would invalidate URLs already issued and in use (if the repo signs URLs with a long TTL persisted somewhere) — report; may need a versioned signature or grace window.
- `MatchableRequest` cannot expose the parsed query for canonicalization without an interface change — report.
- Excerpts don't match live code (drift); a verification fails twice.

## Maintenance notes

- Sign and verify must always share the exact same canonicalization function — note this so a future change to one is forced to touch both.
- Reviewer should confirm the float formatting matches whatever the route-parameter constraints expect (e.g. a `{price}` param regex must accept a decimal point).

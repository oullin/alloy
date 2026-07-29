# Plan 009: Fix validation regex fail-open, delimiter parsing, and wildcard `Validated()` omission

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- pkg/hub/validation`
> On change, reconcile excerpts; on mismatch, STOP.

## Status

- **Priority**: P1
- **Effort**: M
- **Risk**: LOW
- **Depends on**: none
- **Category**: bug (C7 is security-adjacent: a validation bypass)
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

Three validation defects:

1. **Fail-open (C7)**: `validateNotRegex` returns `!validateRegex(...)`. When the pattern fails to compile, `validateRegex` returns `false`, so `not_regex` returns `true` — a malformed `not_regex` rule silently accepts all input instead of failing closed. That is a validation bypass for whatever the rule was meant to block.
2. **Delimiter parsing (C8)**: `regex:`/`not_regex:` only strips surrounding `/.../` when both ends are `/`. A Laravel-style `regex:/^\d+$/i` (trailing flag) is not stripped, so compilation fails and the rule always fails — a silent incompatibility for a Laravel-parity primitive.
3. **Wildcard omission (C9)**: `Validator.Validated()`/`Safe()` iterate `parsedRules` keys and `getValue(attr)`; for a wildcard key like `items.*.name`, `getValue` finds no literal key so the attribute is excluded — validated array/nested data is silently dropped from `validated`.

## Current state

- `pkg/hub/validation/rules/regex.go`:
    - `validateRegex` (20-34): strips `/.../` only when `pattern[0]=='/' && pattern[len-1]=='/'`; `regexp.Compile` failure → `return false`.
    - `validateNotRegex` (37-39): `return !validateRegex("", value, params, ctx)`.
    ```go
    re, err := regexp.Compile(pattern)
    if err != nil { return false }
    return re.MatchString(s)
    ...
    func validateNotRegex(_ string, value any, params []string, ctx RuleContext) bool {
        return !validateRegex("", value, params, ctx)
    }
    ```
- `pkg/hub/validation/validator.go`:
    - `buildValidated` (648-664): iterates `v.parsedRules`, skips excluded, `val := v.getValue(attr)`, includes only when `val != nil || v.flatDataHas(attr)`.
    - `validateAll` elsewhere expands wildcards via `ExpandWildcards` (find it: `grep -n ExpandWildcards pkg/hub/validation`) — reuse that expansion in `buildValidated`.

Convention: rules return `bool`; a compile error is a _rule configuration_ problem, distinct from "value didn't match". To fail `not_regex` closed, `validateRegex` needs to signal "couldn't evaluate" separately from "no match" — either an internal tri-state/error helper or a shared `compilePattern` returning `(*regexp.Regexp, error)` that both rules consult.

## Commands you will need

| Purpose             | Command                                  | Expected |
| ------------------- | ---------------------------------------- | -------- |
| Go validation tests | `cd pkg/hub && go test ./validation/...` | exit 0   |
| Full Go suite       | `pnpm exec vp run go:test`               | exit 0   |
| Format              | `pnpm exec vp run format`                | exit 0   |

## Scope

**In scope**: `pkg/hub/validation/rules/regex.go`, `pkg/hub/validation/validator.go`, and their tests. A small shared `compilePattern` helper may be added.

**Out of scope**: other validation rules; the rule-parser grammar beyond delimiter/flag handling; the `Validator` public API shape.

## Git workflow

- Branch: `advisor/009-validation-fixes`; commit per concern; conventional-commit style.

## Steps

### Step 1: Make regex compilation failure fail closed for both rules

Introduce a shared helper that compiles the (delimiter-stripped) pattern and returns `(*regexp.Regexp, error)`. In `validateRegex`, a compile error means the value cannot satisfy the rule → `false`. In `validateNotRegex`, a compile error must **not** pass — return `false` (fail closed) on compile error, and `!matched` only when the pattern compiled successfully. Do not implement `not_regex` as a blind `!validateRegex`.

**Verify**: `cd pkg/hub && go test ./validation/rules -run Regex` → a malformed-pattern case makes both `regex` and `not_regex` fail (return false).

### Step 2: Handle `/pattern/flags` delimiters

When the pattern is delimited (`/.../` optionally followed by flag letters like `i`, `s`, `m`, `U`), strip the delimiters and translate the trailing PCRE flags to a Go `(?flags)` prefix (RE2 supports `i`, `s`, `m`, `U`). If a flag has no RE2 equivalent, document the limitation and reject rather than silently mis-parsing.

**Verify**: `go test ./validation/rules -run RegexFlags` → `regex:/^\d+$/i` matches case-insensitively; a bare `regex:^\d+$` still works.

### Step 3: Include wildcard-validated attributes in `Validated()`

In `buildValidated` (validator.go:648-664), expand wildcard rule keys via the same `ExpandWildcards` mechanism `validateAll` uses, and include the resolved concrete attributes (and their values) in the `validated` map. Nested/array data validated under `items.*.name` must appear in `Validated()`/`Safe()`.

**Verify**: `go test ./validation -run Validated` → a wildcard rule's validated array values appear in `Validated()`.

### Step 4: Full suite + format

**Verify**: `pnpm exec vp run go:test` → exit 0; `pnpm exec vp run format` → exit 0.

## Test plan

- `rules/regex_test.go`: valid match/no-match; malformed pattern fails both `regex` and `not_regex` closed; `/pattern/flags` with `i`/`s`/`m`.
- `validator_test.go`: `Validated()` includes wildcard-expanded attributes; existing non-wildcard behavior unchanged.

## Done criteria

- [ ] `grep -n "return !validateRegex" pkg/hub/validation/rules/regex.go` returns nothing (no blind negation).
- [ ] A malformed `not_regex` pattern returns false (fails closed) — test asserts it.
- [ ] `regex:/…/i`-style rules compile and match; test asserts it.
- [ ] `Validated()` returns wildcard-expanded attributes; test asserts it.
- [ ] `pnpm exec vp run go:test` exits 0; no out-of-scope files modified; `plans/README.md` row for 009 updated.

## STOP conditions

- A PCRE flag used in the codebase's own rules has no RE2 equivalent (e.g. `x`/extended is fine, but backreferences are not) and rejecting it would break an existing rule — report.
- `ExpandWildcards` cannot be reused in `buildValidated` without a larger refactor — report the coupling before widening scope.
- Excerpts don't match live code (drift); a verification fails twice.

## Maintenance notes

- Document the RE2-only contract for `regex:`/`not_regex:` (no backreferences/lookaround) alongside the flag translation, so consumers porting Laravel rules know the boundary.
- Reviewer should scrutinize the fail-closed semantics: a rule that cannot be evaluated must never pass.

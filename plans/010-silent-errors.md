# Plan 010: Stop swallowing errors that cause silent data loss or crashes

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- pkg/hub/session pkg/hub/str pkg/hub/filesystem pkg/hub/collection web/inertia-demo/api/internal/database`
> On change, reconcile the per-finding excerpts; on mismatch, STOP.

## Status

- **Priority**: P1
- **Effort**: M
- **Risk**: MED
- **Depends on**: none
- **Category**: bug
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

The panic-removal refactor left several paths that now swallow errors and silently return zero values or drop data. Each is a place where a failure produces wrong behavior with no signal:

- **C10** `EncryptedStore.Put` drops the value entirely on encrypt failure (neither encrypted nor plain stored) — silent session-data loss (CSRF tokens, auth flags).
- **C11** session ID / CSRF token generators ignore `crypto/rand` errors → all-zero predictable token on RNG failure.
- **C12** `str.generateRandom` ignores the RNG error → `nil.Int64()` panic (while `Password` correctly checks it).
- **C13** `str.Snake` cache key `value+sep` collides across delimiters → wrong cached result.
- **C14** demo API `scanRows` swallows per-row errors and skips `rows.Err()`; count helpers return 0 on error → truncated result sets returned as success.
- **C16** `LockableFile.Write` doesn't truncate → stale trailing bytes on a shorter rewrite.
- **C18** lazy `Remember()` caches a truncated collection when the first iteration stops early → later iterations replay truncated data.

These are grouped because each is a small, independent, well-scoped fix in the same "surface the error / don't silently drop" theme.

## Current state (per finding)

- **C10** `pkg/hub/session/encrypted_store.go:18-32`:
    ```go
    func (s *EncryptedStore) Put(key string, value any) {
        if str, ok := value.(string); ok {
            encrypted, err := s.enc.Encrypt(str)
            if err != nil { return }   // <- value silently dropped
            s.Store.Put(key, encrypted)
            return
        }
        s.Store.Put(key, value)
    }
    ```
    `Put` has no error return — surfacing the error needs an API decision (see step 1).
- **C11** `pkg/hub/session/store.go:754-766`: `generateID`/`generateToken` do `_, _ = rand.Read(b)` then hex-encode regardless of error.
- **C12** `pkg/hub/str/str.go:1602-1603`: `n, _ := rand.Int(rand.Reader, ...)` then `charset[n.Int64()]` (nil deref on error). `Password` (~1651) checks the error — mirror it.
- **C13** `pkg/hub/str/str.go:403-405`: `key := value + sep` — `Snake("a","_")` and `Snake("a_","")` collide.
- **C14** `web/inertia-demo/api/internal/database/database.go:341-355` (`scanRows`: `if err != nil { continue }`, no `rows.Err()`), and count helpers at `:178-184,233-239,286-300` returning 0 on scan error; list helpers at `:131-158` returning nil on query error.
- **C16** `pkg/hub/filesystem/lockable_file.go:91-103`: `Write` seeks to 0 and writes, no truncate; `Truncate()` is a separate method.
- **C18** `pkg/hub/collection/lazy/pipeline.go:63-86`: sets `cached = true` even when the first pass stopped early (`return yield(item)` halts the source with a partial cache).

Convention: prefer returning an error where the signature allows; where it doesn't (C10 `Put`, C11/C12 in a value-returning context), the accepted patterns in this codebase are (a) add an error return, or (b) panic on a genuinely-impossible CSPRNG failure — `str.Password` already checks-and-returns, so match that where a signature permits.

## Commands you will need

| Purpose                 | Command                                                                           | Expected |
| ----------------------- | --------------------------------------------------------------------------------- | -------- |
| Go tests (touched pkgs) | `cd pkg/hub && go test ./session/... ./str/... ./filesystem/... ./collection/...` | exit 0   |
| Demo API tests          | `cd web/inertia-demo/api && go test ./...`                                        | exit 0   |
| Full Go suite           | `pnpm exec vp run go:test`                                                        | exit 0   |
| Format                  | `pnpm exec vp run format`                                                         | exit 0   |

## Scope

**In scope**: the seven files above and their tests. For C10, the `Store.Put`/`EncryptedStore.Put` signature and its direct callers.

**Out of scope**: session performance (plan 017); broader error-wrapping sweep (deferred X10); the encryption implementation itself.

## Git workflow

- Branch: `advisor/010-silent-errors`; **one commit per finding** (C10…C18) so each is independently reviewable/revertable; conventional-commit style.

## Steps

### Step 1 (C10): EncryptedStore.Put must not silently drop

Decide the `Put` error surface. Preferred: give `Put` an `error` return and thread it through the `Store` interface and callers (search `grep -rn "\.Put(" pkg/hub/session`). If changing the interface is too broad, the fallback is to store the plaintext-safe failure loudly (log + a sticky error flag on the store) — but **never** silently skip persistence. **If `Store.Put` is part of a published contract with external implementers, STOP and report** the signature-change blast radius before proceeding.

**Verify**: a test where `enc.Encrypt` returns an error asserts the write is not silently lost (error surfaced or explicit failure signal). `go test ./session -run EncryptedPut` → pass.

### Step 2 (C11, C12): Handle CSPRNG errors

- `session/store.go` `generateID`/`generateToken`: check the `rand.Read` error; propagate it (return an error from the caller — `Start`/`Regenerate`) or panic on RNG failure. Never emit an all-zero token.
- `str/str.go` `generateRandom`: check the `rand.Int` error and mirror `Password`'s handling (return error / documented fallback) rather than dereferencing a possibly-nil `n`.

**Verify**: unit tests using an error-returning RNG (inject a failing reader where the API allows) assert no zero/predictable token and no panic. `go test ./session ./str -run Rand` → pass.

### Step 3 (C13): Disambiguate the Snake cache key

Change the key to an unambiguous form, e.g. `value + "\x00" + sep`, or key on a struct `{value, sep}`.

**Verify**: `go test ./str -run Snake` → `Snake("a","_")` returns `"a"` and `Snake("a_","")` returns `"a_"` even when computed in either order.

### Step 4 (C14): Propagate DB errors in the demo API

Change `scanRows` to return `([]T, error)`: return per-row scan errors and check `rows.Err()` after the loop. Thread the error through callers (`ListContacts`, `ListContactsPaginated`, `ListRecentNotes`, etc.). Make count helpers return `(int, error)` (or propagate) instead of `0` on scan error; make list helpers distinguish query error from empty.

**Verify**: `cd web/inertia-demo/api && go test ./internal/database/...` → a fake DB erroring mid-stream yields an error, not a truncated success.

### Step 5 (C16): Truncate on LockableFile.Write

After writing, truncate the file to the written length (`lf.file.Truncate(int64(n))`), or add a `WriteAll`/`Overwrite` that truncates-then-writes and document `Write` as append-at-offset. Prefer making `Write` (which already seeks to 0) truncate to the new length, since its intent is whole-file replacement.

**Verify**: `go test ./filesystem -run Write` → rewriting a file with shorter content leaves no stale trailing bytes.

### Step 6 (C18): Only cache a fully-drained lazy collection

In `Remember()`, set `cached = true` only when the source was fully consumed. On early stop (`yield` returned false), do not mark complete — either leave `cached=false` (recompute next pass) or continue filling on the next pass. Ensure a completed pass still caches.

**Verify**: `go test ./collection -run Remember` → `.Remember().First()` followed by a full iteration yields the complete collection, not a truncated one.

### Step 7: Full suites + format

**Verify**: `pnpm exec vp run go:test` → exit 0; `pnpm exec vp run format` → exit 0.

## Test plan

One focused test per finding, in the corresponding package's test file, each asserting the failure is now surfaced (or the data is complete) rather than silently swallowed. Model after existing tests in each package.

## Done criteria

- [ ] `grep -n "if err != nil { return }" pkg/hub/session/encrypted_store.go` no longer silently drops (C10 addressed).
- [ ] `grep -n "_, _ = rand.Read" pkg/hub/session/store.go` returns nothing.
- [ ] Snake, LockableFile, lazy Remember, and demo scanRows tests assert the fixed behavior.
- [ ] `pnpm exec vp run go:test` exits 0; each finding is its own commit.
- [ ] No out-of-scope files modified; `plans/README.md` row for 010 updated.

## STOP conditions

- `Store.Put` is an externally-implemented contract (C10 signature change too broad) — report.
- Threading errors through `scanRows` generics fights the Go type system in a way that needs a larger refactor — report.
- Any excerpt doesn't match live code (drift); a verification fails twice.

## Maintenance notes

- C10's decision (error return vs loud failure) sets a precedent for the rest of the session store API — document it.
- Reviewer should confirm no new silent-drop was introduced while threading errors (e.g. a caller that now ignores the returned error).

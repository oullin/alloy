# Plan 019: Add tests for the untested browserx and passwords SQL repositories

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- pkg/hub/auth/browserx pkg/hub/auth/passwords`
> On change, reconcile excerpts; on mismatch, STOP.

## Status

- **Priority**: P1
- **Effort**: S
- **Risk**: LOW
- **Depends on**: none
- **Category**: tests
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

Two security-critical SQL repositories have no `sql_repository_test.go`, while every sibling repo (tokens, teams, passkeys) does — a clear coverage asymmetry. `browserx/sql_repository.go` backs browser-session revocation (revoking a stolen session); `passwords/sql_repository.go` backs password-reset token storage/consumption. A regression in the query construction, hash comparison, or table-name handling would ship silently. The tokens test specifically guards against SQL-unsafe table names (`"personal_access_tokens; DROP TABLE users"`); the two untested repos have `isSafeSQLIdentifier` guards but no test exercising them.

## Current state

- `pkg/hub/auth/browserx/sql_repository.go`: `NewSQLRepository(db SQLQuerier, table string)` (30); methods `FindForUser` (38), `Revoke` (72), `RevokeOther` (76); `isSafeSQLIdentifier` (80). No `sql_repository_test.go` in the dir (only `browser_sessions_test.go`).
- `pkg/hub/auth/passwords/sql_repository.go`: `NewSQLRepository(db SQLQuerier, table string, expiry time.Duration)` (34); methods `Create` (42), `Exists` (66), `RecentlyCreated` (87), `Delete` (102), `DeleteExpired` (106); `isSafeSQLIdentifier` (112). No `sql_repository_test.go` (dir has `broker_test.go`, `token_repository_test.go`).
- Established harness pattern — `pkg/hub/auth/tokens/sql_repository_test.go`:
    ```go
    type tokenSQLDB struct { lastQuery string; lastExec string; created Token; rows []Token }
    type tokenSQLRow struct { ... }
    type tokenSQLRows struct { rows []Token; pos int }
    func TestSQLRepositoryCreatesHashedTokenWithSafeTable(t *testing.T) {
        db := &tokenSQLDB{}
        repo := NewSQLRepository(db, "personal_access_tokens; DROP TABLE users")
        ... assert db.lastQuery prefix, hash stored not plaintext, table-name safety ...
    }
    ```

Convention: fake the `SQLQuerier`/DB interface (as `tokenSQLDB` does), assert the generated query string, the exec/args, hash comparison, and that an unsafe table name is rejected/sanitized by `isSafeSQLIdentifier`. Read the exact `SQLQuerier` interface each repo expects and mirror the tokens fake.

## Commands you will need

| Purpose       | Command                                                          | Expected |
| ------------- | ---------------------------------------------------------------- | -------- |
| Auth tests    | `cd pkg/hub && go test ./auth/browserx/... ./auth/passwords/...` | exit 0   |
| Full Go suite | `pnpm exec vp run go:test`                                       | exit 0   |
| Format        | `pnpm exec vp run format`                                        | exit 0   |

## Scope

**In scope**: create `pkg/hub/auth/browserx/sql_repository_test.go` and `pkg/hub/auth/passwords/sql_repository_test.go`. Test-only — no production code changes.

**Out of scope**: production repo code (if a test reveals a bug, STOP and report it as a finding, don't fix it here); the real-DB integration harness (deferred).

## Git workflow

- Branch: `advisor/019-auth-sql-repo-tests`; one commit per repo; conventional-commit style.

## Steps

### Step 1: browserx SQL repository tests

Create `browserx/sql_repository_test.go` mirroring the tokens harness. Fake the DB, then assert for each method:

- `FindForUser`: query targets the configured table with the user-id bound as a parameter; rows map to `Session`.
- `Revoke`/`RevokeOther`: correct `WHERE` clause (user + session id / user + not-current-session), values bound as parameters.
- Table-name safety: constructing with an unsafe table name (`"sessions; DROP TABLE users"`) is rejected or sanitized by `isSafeSQLIdentifier` (assert the default/safe behavior).

**Verify**: `cd pkg/hub && go test ./auth/browserx -run SQLRepository` → pass.

### Step 2: passwords SQL repository tests

Create `passwords/sql_repository_test.go` mirroring the harness. Assert for each method:

- `Create`: INSERT into the configured table; the token is stored **hashed**, not plaintext (mirror the tokens test's plaintext-leak assertion); expiry applied.
- `Exists`: token compared via hash (constant-time if the repo does so — assert the compare path); correct email/token binding.
- `RecentlyCreated`: time-window predicate uses bound args.
- `Delete`/`DeleteExpired`: correct predicates; expiry cutoff bound as an arg.
- Table-name safety as in step 1.

**Verify**: `cd pkg/hub && go test ./auth/passwords -run SQLRepository` → pass.

### Step 3: Full suite + format

**Verify**: `pnpm exec vp run go:test` → exit 0; `pnpm exec vp run format` → exit 0.

## Test plan

The plan _is_ the test plan: table-driven tests per method, following `tokens/sql_repository_test.go`, covering query construction, parameter binding, hash comparison, and unsafe-table-name handling. Include the negative case (unsafe identifier) explicitly.

## Done criteria

- [ ] `pkg/hub/auth/browserx/sql_repository_test.go` and `pkg/hub/auth/passwords/sql_repository_test.go` exist and pass.
- [ ] Each covers all public methods, parameter binding, hash storage (passwords), and the unsafe-table-name guard.
- [ ] `pnpm exec vp run go:test` exits 0; no production files modified (`git status`); `plans/README.md` row for 019 updated.

## STOP conditions

- A test reveals an actual bug in the production repo (e.g. plaintext token stored, or `isSafeSQLIdentifier` not actually applied) — STOP and report it as a finding rather than changing production code in this test-only plan.
- The repo's DB interface differs materially from tokens' such that the fake can't be reused — report; may need a slightly different fake.
- Excerpts don't match live code (drift).

## Maintenance notes

- These string-level tests verify query construction, not real-DB behavior; note that a real-DB integration harness (deferred X2/related) would complement them.
- Reviewer should confirm the plaintext-leak and unsafe-table assertions are present (the two highest-value guards).

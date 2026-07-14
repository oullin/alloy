# Plan 018: Speed up the Go CI job (parallel modules, drop duplicate vet, cache coverage)

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- infra/scripts/tasks/go-test.sh vite.config.ts .github/actions/setup`
> On change, reconcile excerpts; on mismatch, STOP.

## Status

- **Priority**: P3
- **Effort**: M
- **Risk**: MED
- **Depends on**: none
- **Category**: dx (perf/CI)
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

The Go test job is on the CI critical path and is slower than it needs to be:
- `infra/scripts/tasks/go-test.sh` iterates each `go.mod` **sequentially** in a `while` loop (lines 28-63), running `go vet ./...` then `go test -race ./...` per module.
- `go test` re-runs vet by default, so the explicit `go vet` is a duplicate analysis pass.
- `-race` adds 2–10× runtime and is applied to every module unconditionally.
- `vite.config.ts:83-104` sets `cache: false` on every `run.task` including `go:test`.
- `.github/actions/setup/action.yml` uses `setup-go@v6` without a `cache-dependency-path` spanning the multiple `go.sum` files, risking partial module/build cache hits across the 3 Go modules.

## Current state

- `infra/scripts/tasks/go-test.sh:28-63`:
  ```sh
  while IFS= read -r -d '' gomod; do
      module_dir="$(dirname "${gomod}")"
      ( cd "${module_dir}"; ...
        go vet ./...
        ... go test -race ./...  # (or -coverprofile variant)
      )
  done < <( find "${GO_PATH}" -name go.mod -print0; find "${ROOT_PATH}/web" -path '*/api/go.mod' -print0 ... )
  ```
- `vite.config.ts:83-104`: `run.task` entries with `cache: false`.
- `.github/actions/setup/action.yml`: `setup-go@v6` without `cache-dependency-path`.

Convention: the script already validates module paths and computes per-module `GOWORK` and coverage profiles — preserve all of that. Coverage output (`GO_COVERAGE_DIR/summary.tsv`) must stay intact.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Run the go-test task locally | `pnpm exec vp run go:test` | exit 0, all modules pass |
| Shell lint (if available) | `shellcheck infra/scripts/tasks/go-test.sh` | no new errors |
| Format | `pnpm exec vp run format` | exit 0 |

## Scope

**In scope**: `infra/scripts/tasks/go-test.sh`, `vite.config.ts` (the `go:test` task cache setting only), `.github/actions/setup/action.yml` (cache-dependency-path).

**Out of scope**: the test code itself; other CI workflows; the coverage-summary format. **Note**: editing `.github/**` may require workflow-scope permissions the maintainer must apply (see the repo's gh-token constraint) — flag it in the PR.

## Git workflow

- Branch: `advisor/018-ci-perf`; commit per concern; conventional-commit style.

## Steps

### Step 1: Drop the duplicate vet

Since `go test` runs vet by default, either remove the standalone `go vet ./...` line and rely on the test's built-in vet, or keep the explicit vet and pass `-vet=off` to `go test`. Prefer keeping one explicit `go vet` and adding `-vet=off` to `go test` (so vet failures are attributable), whichever the maintainer prefers — pick one and remove the duplication.

**Verify**: `pnpm exec vp run go:test` → still fails on a deliberately-introduced vet issue exactly once (not double-reported), then remove the deliberate issue.

### Step 2: Run modules concurrently

Change the per-module loop to run modules in parallel (background subshells with a wait + failure aggregation, or GNU `parallel` if the CI image has it). Preserve per-module `GOWORK`, the module-path validation, and the coverage profile/summary writes (ensure concurrent appends to `summary.tsv` don't interleave — write per-module files and concatenate at the end, or lock the append). Fail the task if any module fails.

**Verify**: `pnpm exec vp run go:test` → all modules run, overall exit reflects any single failure; `summary.tsv` has one line per module (no interleaving).

### Step 3: Enable task cache for go:test (if safe)

In `vite.config.ts`, evaluate enabling `cache` for the `go:test` task keyed on the Go sources + `go.sum` files. Only do this if vite-plus task caching keys correctly on Go inputs; if uncertain, leave `cache: false` and note why. (Lower-risk than the parallelism win.)

**Verify**: two consecutive `pnpm exec vp run go:test` runs — the second is a cache hit (if enabled) and still correct.

### Step 4: Fix setup-go cache coverage

In `.github/actions/setup/action.yml`, set `cache-dependency-path` to a glob covering all `go.sum` files (`pkg/hub/go.sum`, `pkg/hub/auth/passkeys/go.sum`, `pkg/hub/queue/drivers/sqs/go.sum`, `web/inertia-demo/api/go.sum` — confirm the full set via `find . -name go.sum -not -path '*/node_modules/*'`), or add an explicit `actions/cache` step for `~/.cache/go-build` + `~/go/pkg/mod`.

**Verify**: (CI-observable) subsequent runs hit the Go module/build cache for all modules. Locally, confirm the YAML is valid and the glob lists every `go.sum`.

### Step 5: (optional) Split `-race` to its own job

Consider running a fast non-race pass as the PR gate and `-race` as a separate (possibly non-blocking or nightly) job, if `-race` dominates wall-clock. Only if the maintainer wants it — otherwise keep `-race` on the main pass.

### Step 6: Format

**Verify**: `pnpm exec vp run format` → exit 0.

## Test plan

- No product tests change. Validate the script still passes/fails correctly: introduce a temporary failing Go test in one module and confirm the task fails; remove it. Confirm coverage summary integrity.

## Done criteria

- [ ] `go vet` is not run twice per module.
- [ ] Modules run concurrently; a single module failure fails the task; `summary.tsv` is not interleaved.
- [ ] `setup-go` caches across all `go.sum` files (or an explicit cache step covers them).
- [ ] `pnpm exec vp run go:test` exits 0 on a clean tree.
- [ ] No out-of-scope files modified; `plans/README.md` row for 018 updated.

## STOP conditions

- Concurrent coverage writes cannot be made race-free within the shell script without significant complexity — fall back to sequential coverage but keep vet/cache wins, and report.
- `.github/**` edits are blocked by workflow-scope token limits — make the `go-test.sh`/`vite.config.ts` changes and flag the `setup/action.yml` change for the maintainer to apply.
- Excerpts don't match live code (drift).

## Maintenance notes

- Note the coverage-aggregation approach (per-module files → concat) so future module additions slot in.
- Reviewer should confirm no loss of vet coverage and that `-race` is still applied where it matters.

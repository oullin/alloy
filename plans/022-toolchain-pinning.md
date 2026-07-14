# Plan 022: Pin the `latest`-floating toolchain versions

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- package.json Makefile .github/workflows/govulncheck.yml pnpm-lock.yaml`
> On change, reconcile; on mismatch, STOP.

## Status

- **Priority**: P2
- **Effort**: S
- **Risk**: LOW
- **Depends on**: none
- **Category**: dependencies
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

The repo SHA-pins every GitHub Action but floats its core toolchain on `latest`:
- **package.json**: `"vite": "npm:@voidzero-dev/vite-plus-core@latest"` and `"vite-plus": "latest"` (and the `pnpm.overrides.vite` alias) — both resolve to a **0.x** release (`@0.2.1` in the lockfile). Vite+ (`vp`) is the entire orchestration layer (every build/lint/test/typecheck/pack routes through it); on a 0.x package, `latest` means the next lockfile-regenerating install can jump to `0.3.x` where SemVer permits breaking changes, silently breaking the whole toolchain with no manifest change.
- **`.github/workflows/govulncheck.yml:35`**: `go install golang.org/x/vuln/cmd/govulncheck@latest`.
- **`Makefile:4,11`**: `TASK_VERSION ?= latest` feeding `go install github.com/go-task/task/v3/cmd/task@$(TASK_VERSION)`.

This is inconsistent with the otherwise strict supply-chain posture and makes CI/security-scan behavior non-reproducible.

## Current state

- `package.json:37-38`:
  ```json
  "vite": "npm:@voidzero-dev/vite-plus-core@latest",
  "vite-plus": "latest"
  ```
  plus `pnpm.overrides.vite` = `npm:@voidzero-dev/vite-plus-core@latest`.
- Lockfile resolves both to `@voidzero-dev/vite-plus-core@0.2.1` / `vite-plus@0.2.1`.
- `.github/workflows/govulncheck.yml:35` — `govulncheck@latest`.
- `Makefile:4` — `TASK_VERSION ?= latest`; `Makefile:11` uses it in `go install ...@$(TASK_VERSION)`.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Current resolved versions | `grep -A2 "vite-plus\|vite-plus-core" pnpm-lock.yaml \| head` | shows 0.2.1 |
| Install with frozen lock | `pnpm install --frozen-lockfile` | exit 0 |
| Toolchain smoke test | `pnpm exec vp check` | exit 0 |
| Format | `pnpm exec vp run format` | exit 0 |

## Scope

**In scope**: `package.json` (the two deps + the override), `.github/workflows/govulncheck.yml`, `Makefile`. Possibly a Renovate/Dependabot entry so these bump deliberately.

**Out of scope**: upgrading to a newer vite-plus version (this plan pins to what's already resolved, `0.2.1`); other dependencies. **Note**: editing `.github/**` may need workflow-scope permissions — flag for the maintainer.

## Git workflow

- Branch: `advisor/022-toolchain-pinning`; commit per file; conventional-commit style.

## Steps

### Step 1: Pin vite-plus / vite-plus-core to the resolved version

In `package.json`, change `"vite"` and `"vite-plus"` (and the `pnpm.overrides.vite`) from `@latest`/`latest` to the exact currently-resolved version (`0.2.1` — confirm from the lockfile). For a pre-1.0 tool, prefer an exact pin over a caret. Keep the `npm:@voidzero-dev/vite-plus-core@` alias form, just with the fixed version.

**Verify**: `pnpm install --frozen-lockfile` → exit 0 with no lockfile change; `pnpm exec vp check` → exit 0.

### Step 2: Pin govulncheck

In `.github/workflows/govulncheck.yml`, replace `govulncheck@latest` with an explicit version tag (the latest known-good release; check `go list -m -versions golang.org/x/vuln` or the project's releases). Add it to whatever bump automation the repo uses (Dependabot config already exists — add an entry if applicable).

**Verify**: YAML is valid; the pinned version is a real release.

### Step 3: Pin the task version

In `Makefile`, set `TASK_VERSION` to an explicit version instead of `latest` (keep it overridable via `?=` so a caller can bump it). Use a current go-task release.

**Verify**: `make <a formatting target that uses task>` or a dry check confirms the pinned version installs; the Makefile still parses.

### Step 4: Format

**Verify**: `pnpm exec vp run format` → exit 0.

## Test plan

- No product tests. Validate: `pnpm install --frozen-lockfile` produces no lockfile diff; `pnpm exec vp check`/`vp test` still run; the pinned CI-tool versions are real releases.

## Done criteria

- [ ] `grep -n "@latest\|\"latest\"" package.json` returns nothing for vite/vite-plus (exact versions pinned, incl. the override).
- [ ] `govulncheck@<version>` (not `@latest`) in the workflow.
- [ ] `TASK_VERSION` defaults to an explicit version.
- [ ] `pnpm install --frozen-lockfile` and `pnpm exec vp check` succeed.
- [ ] No out-of-scope files modified; `plans/README.md` row for 022 updated.

## STOP conditions

- Pinning vite-plus to `0.2.1` fails to install or breaks `vp` (the resolved version differs from what's actually installed) — reconcile against the real installed version and report.
- `.github/**` edits are blocked by token scope — make the `package.json`/`Makefile` changes and flag the workflow pin for the maintainer.
- Excerpts don't match live code (drift).

## Maintenance notes

- These are pre-1.0/floating tools; wire them into Dependabot/Renovate so bumps are reviewed PRs, not silent installs.
- Reviewer should confirm the pinned versions match what CI currently uses (no behavior change, just reproducibility).

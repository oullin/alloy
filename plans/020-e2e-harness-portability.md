# Plan 020: Make the inertia e2e harness portable and de-duplicate the two runners

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- web/inertia-demo/tests/e2e`
> On change, reconcile excerpts; on mismatch, STOP.

## Status

- **Priority**: P2
- **Effort**: M
- **Risk**: LOW
- **Depends on**: none
- **Category**: tests / dx
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

The e2e suite covers the demo's critical paths (auth, CRM CRUD, a 50+ route crawl) but is reproducible only on the original author's machine:
- **X4**: `runner.mjs`/`runner.playwright.mjs` default artifacts to `/Users/gocanto/.cache/codex/browser-artifacts`, default the parity source to a local `/Users/gocanto/Sites/bedrock/...` checkout, and resolve a **Helium** browser at `/Applications/Helium.app/...` plus an `agent-browser` binary. A clean CI runner or any other contributor can't run the primary path without reverse-engineering env overrides.
- **X5**: Two ~690-line runners (`runner.mjs` agent-browser, `runner.playwright.mjs`) duplicate 23 of 34 top-level functions (`startServer`, `runAuthFlow`, `runCrmFlow`, `routeMatrix`, …); adding a route or flow means editing both or letting them drift. The README already calls the Playwright runner "legacy".

## Current state

- `web/inertia-demo/tests/e2e/runner.mjs:13` — `const defaultArtifactsPath = '/Users/gocanto/.cache/codex/browser-artifacts';`
- `runner.mjs:448/485` — default Bedrock parity source `/Users/gocanto/Sites/bedrock/services/demo/inertia`.
- `runner.mjs:634-651` — resolves Helium at `/Applications/Helium.app/...`, shells out to `agent-browser` (devDep `agent-browser@^0.31.1`).
- `runner.playwright.mjs` — 693 lines, shares 23/34 function names with `runner.mjs`; wired via `package.json` `test:alloy:playwright`; README:3 calls it "legacy".

Convention: `package.json` scripts drive the suites (`test:alloy:agent-browser`, `test:alloy:playwright`). Machine-specific values should come from documented env vars with a fail-fast message, not hardcoded absolute paths.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| List e2e scripts | `cat web/inertia-demo/tests/e2e/package.json` (or root) | shows the test:* scripts |
| Run playwright e2e | `pnpm test:e2e:inertia` (confirm the exact script) | runs (may need a browser) |
| Format | `pnpm exec vp run format` | exit 0 |

## Scope

**In scope**: `web/inertia-demo/tests/e2e/runner.mjs`, `runner.playwright.mjs`, any shared module extracted, the e2e `package.json` scripts, and a `.env.example`/README for the e2e dir.

**Out of scope**: the demo app or API code; the flows' assertions themselves (preserve behavior); CI workflow wiring beyond what's needed to point at the canonical runner (flag `.github/**` edits for the maintainer).

## Git workflow

- Branch: `advisor/020-e2e-harness-portability`; commit per concern; conventional-commit style.

## Steps

### Step 1: Move machine defaults into documented env vars

Replace the hardcoded absolute paths (`defaultArtifactsPath`, the Bedrock parity source, the Helium/agent-browser locations) with env vars that have sensible portable defaults (e.g. artifacts under the repo's `web/storage` or an OS temp dir; parity source optional/skipped when unset). When a required external tool (Helium/agent-browser) is unavailable, fail fast with a message naming the env var to set — do not silently fall back to a machine-specific path. Add a `web/inertia-demo/tests/e2e/.env.example` (or README table) listing each var, default, and purpose.

**Verify**: `grep -rn "/Users/gocanto" web/inertia-demo/tests/e2e` returns nothing; running the runner without the optional vars either runs (portable default) or fails with a clear, actionable message.

### Step 2: Make Playwright/Chromium the canonical default

Since the Playwright runner is the portable one, make it the default CI/local path; keep agent-browser/Helium as an opt-in (documented) alternative rather than the primary. Update the README so the canonical command is the portable one.

**Verify**: the documented default e2e command uses Chromium/Playwright and runs on a machine without Helium.

### Step 3: Extract shared flow/route-matrix/server-lifecycle module

Pull the 23 shared functions (`startServer`/`stopServer`, `runAuthFlow`, `runCrmFlow`, `routeMatrix`, `freePort`, `canLoad`, …) into one module (e.g. `shared.mjs`). Reduce each runner to a thin driver adapter (agent-browser vs Playwright) over the shared logic. If the Playwright runner is truly the future and agent-browser is legacy, an acceptable alternative is to delete `runner.mjs` and keep only the Playwright runner — confirm the maintainer's intent before deleting (STOP gate).

**Verify**: adding a hypothetical route to the matrix requires editing only `shared.mjs`; both runners (or the single kept one) still execute the flows. `grep -c "^async function\|^function" runner*.mjs` shows the duplication removed.

### Step 4: Format

**Verify**: `pnpm exec vp run format` → exit 0.

## Test plan

- The e2e suite itself is the test. After refactoring, run the canonical (Playwright) suite against the demo and confirm the auth flow, CRM CRUD, and route crawl still pass. Behavior must be unchanged — this is a portability/dedup refactor, not a test-logic change.

## Done criteria

- [ ] `grep -rn "/Users/gocanto\|/Applications/Helium" web/inertia-demo/tests/e2e` returns nothing.
- [ ] The canonical e2e command runs on a machine without Helium/agent-browser (Playwright/Chromium).
- [ ] Shared flows live in one module; the runners are thin adapters (or only the Playwright runner remains).
- [ ] `web/inertia-demo/tests/e2e/.env.example` (or README) documents every env var.
- [ ] No out-of-scope files modified; `plans/README.md` row for 020 updated.

## STOP conditions

- Deleting `runner.mjs` (agent-browser) is on the table but the maintainer hasn't confirmed agent-browser is being retired — keep both as thin adapters and report.
- The Playwright runner can't actually cover a flow the agent-browser one does (feature gap) — report the gap rather than dropping coverage.
- Excerpts don't match live code (drift).

## Maintenance notes

- After extraction, a route/flow change is a one-file edit; note this so future contributors don't reintroduce a second copy.
- Reviewer should confirm the portable default truly runs with no machine-specific setup, and that CI points at it.

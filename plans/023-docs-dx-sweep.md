# Plan 023: Fix stale docs, add a root AGENTS.md, and unify the Node baseline

> **Executor instructions**: Follow step by step; verify each step; STOP on any STOP condition; update `plans/README.md` when done.
>
> **Drift check (run first)**: `git diff --stat bfface5..HEAD -- README.md CONTRIBUTING.md web/docs .github/actions/setup`
> On change, reconcile; on mismatch, STOP.

## Status

- **Priority**: P2
- **Effort**: M
- **Risk**: LOW
- **Depends on**: none
- **Category**: docs / dx
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

First-contact documentation is actively wrong after the `packages/foundation → pkg/hub` rename, and there's no in-repo agent brief despite the project explicitly marketing AI-assisted contribution:
- **X8**: `README.md` links `[Development workflow](docs/development.md)` but no root `docs/` dir exists; README/CONTRIBUTING say TypeScript packages live under `packages/` but they live under `sdk/`; `web/docs/architecture/drivers.md` presents `log`/`mailx`/`concurrency`/`notifications` packages (with GitHub source links) that don't exist; the empty `packages/` dir remains (removal handled in plan 021).
- **X9**: No root `CLAUDE.md`/`AGENTS.md` though `web/docs/getting-started.md:144` pitches "easy for AI agents to understand and contribute", and the repo has non-obvious conventions (Vite+ owns orchestration, Docker-backed formatting, `pkg/hub` Go / `sdk/*` TS layout, `doc.go`/`errors.go` per package). Plus three conflicting Node baselines: README says 22, CONTRIBUTING says 20, CI (`.github/actions/setup/action.yml`) uses 24, with `engine-strict=true` set.

## Current state

- `README.md:4,27` — "reusable packages under `packages/`" / "TypeScript packages live directly ... under `packages/`" (should be `sdk/`); `:100-103` links `docs/development.md` (nonexistent) and package READMEs; `:31` — "Node.js 22 or newer".
- `CONTRIBUTING.md:9` — "Node.js 20 or newer"; `:38` — "TypeScript packages live directly under `packages/<name>`".
- `web/docs/getting-started.md:42` — layout diagram shows "`packages/  TypeScript packages and the foundation module`"; package-index tables link `/packages/<name>`; `:129-143` lists unbuilt roadmap packages (with a caveat).
- `web/docs/architecture/drivers.md:214-218` — "Manager source" table lists `log`/`mailx`/`concurrency`/`notifications` with `pkg/hub/<name>/manager.go` GitHub links; those packages do not exist (`ls pkg/hub` confirms).
- `.github/actions/setup/action.yml` — `node-version: 24`; `.npmrc` sets `engine-strict=true`.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| Find stale path refs | `grep -rn "packages/foundation\|under \`packages\`\|/packages/" README.md CONTRIBUTING.md web/docs` | enumerate hits to fix |
| Build docs site | `pnpm docs:build` | exit 0 (links resolve) |
| Node baseline refs | `grep -rn "Node.js\|node-version\|engines" README.md CONTRIBUTING.md .github package.json` | shows the three values |
| Format | `pnpm exec vp run format` | exit 0 |

## Scope

**In scope**: `README.md`, `CONTRIBUTING.md`, `web/docs/**` (getting-started, drivers.md, any other stale-path page), a new root `AGENTS.md` (with `CLAUDE.md` symlink/pointer if the repo wants both), the Node baseline in README/CONTRIBUTING/`.github/actions/setup/action.yml` and a root `engines` field.

**Out of scope**: removing the empty `packages/` dir (plan 021); building the not-yet-existent packages (that's roadmap); the actual toolchain versions (plan 022). **Note**: `.github/**` edits may need workflow-scope permissions — flag for the maintainer.

## Git workflow

- Branch: `advisor/023-docs-dx-sweep`; commit per concern; conventional-commit style.

## Steps

### Step 1: Sweep `packages/` → `sdk/` (TS) / `pkg/hub/` (Go) across docs

Fix every stale path/claim: README `:4,27`, CONTRIBUTING `:38`, `web/docs/getting-started.md:42` and its package-index link targets, dropping the "foundation module" language. Repoint or remove the README `docs/development.md` link (either create the doc, point to the existing `web/docs` equivalent, or remove the link). Verify no remaining `packages/foundation` or `/packages/<name>` references for TS packages.

**Verify**: `grep -rn "packages/foundation\|/packages/" README.md CONTRIBUTING.md web/docs` returns nothing stale; `pnpm docs:build` → exit 0.

### Step 2: Fix `drivers.md` phantom packages

Split the "Manager source" table into shipped vs. roadmap, or remove the `log`/`mailx`/`concurrency`/`notifications` rows (and their 404 source links) until those packages land. Do not present unbuilt packages as available.

**Verify**: every source link in `drivers.md` points to a path that exists (`for p in log mailx concurrency notifications; do ls pkg/hub/$p 2>/dev/null; done` returns nothing → those rows must be gone or clearly marked roadmap).

### Step 3: Unify the Node baseline

Pick one supported Node major (recommend aligning to CI's `24`, or the lowest you actually test — decide and state it). Set it consistently in README `:31`, CONTRIBUTING `:9`, `.github/actions/setup/action.yml` `node-version`, and add a root `package.json` `engines.node` field. Given `engine-strict=true`, the documented baseline must match what CI validates.

**Verify**: `grep -rn "Node" README.md CONTRIBUTING.md` and the CI `node-version` all state the same major; `pnpm install` respects `engines` without error on that version.

### Step 4: Add a root AGENTS.md

Write a concise `AGENTS.md` capturing the non-obvious conventions an executor/contributor needs: run everything through `vp` (`pnpm exec vp check/lint/test`, `vp run go:test`); formatting is Docker-backed (`pnpm exec vp run format`); layout (`pkg/hub` Go module, `sdk/*` TS packages, nested Go modules at `queue/drivers/sqs` and `auth/passkeys`, demo at `web/inertia-demo`); the `doc.go`/`errors.go`-per-package convention; `workspace:*`-only internal deps; the chosen Node baseline (from step 3). If the repo wants `CLAUDE.md` too, make it a pointer/symlink to `AGENTS.md` rather than a second copy. Keep it short — link out rather than duplicating the README.

**Verify**: `AGENTS.md` exists at the repo root and its stated commands actually work (`pnpm exec vp check` runs); no duplicated content that will drift from README.

### Step 5: Build docs + format

**Verify**: `pnpm docs:build` → exit 0; `pnpm exec vp run format` → exit 0.

## Test plan

- Docs build passes (dead internal links fail the VuePress build — that's the gate for steps 1–2).
- Manual: the AGENTS.md commands run; the Node baseline is identical in all four places.

## Done criteria

- [ ] No stale `packages/`/`packages/foundation` references for TS packages in README/CONTRIBUTING/web/docs; `docs/development.md` link resolved or removed.
- [ ] `drivers.md` no longer links nonexistent packages as shipped.
- [ ] One Node baseline across README, CONTRIBUTING, CI, and `engines`.
- [ ] A concise root `AGENTS.md` exists with working commands.
- [ ] `pnpm docs:build` exits 0; no out-of-scope files modified; `plans/README.md` row for 023 updated.

## STOP conditions

- The maintainer must decide the canonical Node major (24 vs the lowest tested) — implement the recommended default (match CI = 24) and note the decision for confirmation.
- `.github/**` edits are blocked by token scope — make the docs/README/CONTRIBUTING/`engines` changes and flag the `setup/action.yml` `node-version` change for the maintainer.
- `pnpm docs:build` fails for a reason unrelated to your link fixes — report.

## Maintenance notes

- Keep AGENTS.md short and link-based so it doesn't drift from README; the package layout is the part most likely to go stale after future moves.
- Reviewer should confirm the AGENTS.md commands are the ones CI actually uses.

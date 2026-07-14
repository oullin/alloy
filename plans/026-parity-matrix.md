# Plan 026: Publish a cross-runtime parity matrix and the policy for what earns a TS twin

> **Executor instructions**: This is primarily a **documentation/decision** task. Produce a parity matrix doc and a short policy; do not change package code. Deliverable: a docs page (e.g. `web/docs/architecture/parity.md`) plus a linked policy. Update `plans/README.md` when done.

## Status

- **Priority**: P3
- **Effort**: M
- **Risk**: LOW
- **Depends on**: none (but coordinate with plans 008 and 025)
- **Category**: direction / docs
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

The value proposition is "cross-runtime primitives", but only 3 of ~22 Go packages have TS twins (`tempo`, `money`, `workflow`); `console` and `navigator-routes` are TS-only. `README.md:6` ("Tempo is the most complete cross-runtime package") implies parity is a goal, yet the asymmetry is undocumented, so consumers can't tell which primitives exist in their runtime. Either a roadmap toward more twins or an explicit scoping decision is missing. A published parity matrix + a "what earns a twin" policy resolves the ambiguity cheaply and guides future work (including plan 025's authkit cross-runtime question).

## Current state (facts to tabulate)

- Go packages (`pkg/hub`): auth, bus, cache, collection, config, container, contracts, cookie, database, encryption, events, filesystem, hashing, httpx, inertia, money, queue, seo, session, str, tempo, validation, workflow (+ nested modules `queue/drivers/sqs`, `auth/passkeys`).
- TS packages (`sdk/*`): `tempo`, `money`, `workflow` (Go twins), `console`, `navigator-routes` (TS-only).
- Plan 008 (if landed) adds conformance fixtures for the money/tempo twins — reference it as the mechanism that *enforces* parity where it's claimed.

## Scope

**In scope**: a parity matrix doc listing every package as Go-only / TS-only / both, plus a short written policy for what qualifies a primitive for a twin and how parity is verified (pointing at plan 008's fixtures). Optionally a small check that flags a claimed-twin without conformance coverage.

**Out of scope**: building new twins; the conformance fixtures themselves (plan 008); the authkit cross-runtime decision (plan 025, which this informs).

## Steps

### Step 1: Build the matrix

Enumerate every Go and TS package and classify each as Go-only / TS-only / both. For the "both" rows, note whether conformance coverage exists (plan 008). Source the list from `ls pkg/hub` and `ls sdk` (don't hand-wave — list them all).

### Step 2: Write the policy

Define, in a few sentences each: what earns a TS twin (e.g. primitives whose logic must produce identical results on backend and frontend — money, dates, workflow state — vs. server-only concerns like queue/container/httpx that have no frontend meaning), and the rule that any claimed twin must have conformance fixtures (plan 008) so parity is mechanically enforced, not asserted.

### Step 3: Recommend the next 1–2 twins (grounded)

Based on demand signals in the repo (what the demo frontend hand-rolls, what `README`/`getting-started` imply), recommend which packages, if any, should become twins next — or explicitly state that the current 3 are the intended scope. Ground each recommendation in evidence, not breadth.

### Step 4: Publish and link

Add the matrix + policy as `web/docs/architecture/parity.md` (or the fitting docs location), link it from the README's cross-runtime claim (`README.md:6`) and getting-started, and ensure `pnpm docs:build` passes.

### Step 5 (optional): Parity lint

If cheap, add a check that fails when a package exists in both `pkg/hub` and `sdk` but has no conformance fixture (ties to plan 008). Only if it's a small, reliable check.

## Commands you will need

| Purpose | Command | Expected |
|---------|---------|----------|
| List packages | `ls pkg/hub; ls sdk` | full package lists |
| Build docs | `pnpm docs:build` | exit 0 |
| Format | `pnpm exec vp run format` | exit 0 |

## Deliverable / done criteria

- [ ] `web/docs/architecture/parity.md` (or equivalent) lists every package classified Go-only / TS-only / both, with conformance-coverage notes for twins.
- [ ] A written policy states what earns a twin and that twins require conformance fixtures.
- [ ] A grounded recommendation on the next twins (or an explicit "current scope is intentional").
- [ ] The doc is linked from README/getting-started; `pnpm docs:build` exits 0.
- [ ] No package code changed; `plans/README.md` row for 026 updated.

## STOP conditions

- The "next twins" recommendation would commit real engineering the maintainer hasn't agreed to — keep it a recommendation with trade-offs, don't schedule the build.
- `pnpm docs:build` fails for reasons unrelated to your additions — report.

## Maintenance notes

- Keep the matrix updated whenever a package is added or twinned; the plan-008 conformance link is what keeps "both" honest.
- This doc should inform plan 025's decision on whether `authkit`/`authflows` get TS twins.

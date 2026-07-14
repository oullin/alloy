# Plan 024 (SPIKE): Design the private-package distribution path for the subscription model

> **Executor instructions**: This is a **design/investigation spike**, not a build task. Produce a written design document plus a minimal proof-of-concept where noted — do not build the full pipeline. The deliverable is `docs/design/private-distribution.md` (create the `docs/design/` dir) answering the open questions with a recommendation. Update `plans/README.md` when done.

## Status

- **Priority**: P3
- **Effort**: L (spike-scoped: ~1–2 days of investigation + a POC)
- **Risk**: MED (touches release workflows and access control — but the spike only prototypes)
- **Depends on**: none
- **Category**: direction
- **Planned at**: commit `bfface5`, 2026-07-14

## Why this matters

The commercial model is subscription access to private packages, but there is **no delivery mechanism**. `CONTRIBUTING.md:40-42` states all `@alloy/*` packages are `"private": true` and "consumed as `workspace:*` dependencies or release tarballs — never published to a public registry." `release-ts.yml` only packs `.tgz` and attaches them to a GitHub release; there is no registry publish step and no Go-module distribution story for paying consumers. Today a customer cannot `pnpm add` an Alloy package or `go get` a hub package — they'd hand-install tarballs. This spike defines how customers actually receive and version the packages.

## Current state (facts to ground the design)

- TS packages under `sdk/*`, all `"private": true`, internal deps via `workspace:*`.
- `release-ts.yml` — packs tarballs, attaches to a GitHub release (read it to confirm the exact current flow).
- `release-go.yml` — read it for the current Go release/tag flow; the Go module is `github.com/oullin/alloy/pkg/hub` with nested modules (`queue/drivers/sqs`, `auth/passkeys`).
- Memory/context: commercial goal is private packages behind a subscription; breaking changes acceptable pre-1.0; there is intentionally no public `alloy.dev` registry domain.

## Scope

**In scope**: a design doc and a **minimal** POC (e.g. publishing one package to one candidate private registry in a throwaway/dry-run mode). No production pipeline, no entitlement system build-out, no customer-facing changes.

**Out of scope**: actually publishing packages; building auth/entitlement infrastructure; pricing/packaging business decisions (surface them as open questions for the maintainer).

## Investigation steps

### Step 1: Map the requirements

Document what "distribution" must satisfy: normal dependency resolution (`pnpm add @alloy/...`, `go get`), semver versioning, access gating by subscription, and how the ~5 TS packages + Go modules each flow. Note the constraint that there is deliberately no public registry.

### Step 2: Evaluate registry options (TS)

Compare, with trade-offs: GitHub Packages (npm registry, token-gated by repo/org access), a self-hosted Verdaccio, and a commercial private registry (e.g. npmjs private, Cloudsmith). For each: how a customer authenticates, how entitlement maps to access, CI publish integration, and cost/ops burden.

### Step 3: Evaluate the Go-module story

Go modules can't live in a classic npm-style private registry. Options: a private Go module proxy (Athens, or a commercial proxy), `GOPRIVATE` + a git-access model (customers get read access to a distribution repo), or vendored tarballs. Document how versioning/tagging works for the multi-module layout (`pkg/hub` + nested modules).

### Step 4: Entitlement/token model

Sketch how a subscription maps to a credential the customer uses (registry token / git deploy key / proxy token), how it's issued and revoked, and where that lives. Keep it a sketch — flag the build-out as follow-up work.

### Step 5: Minimal POC

Prove the happy path for **one** TS package to the recommended registry in dry-run/throwaway mode (e.g. `npm publish --dry-run` against the candidate registry, or a local Verdaccio). Document the exact commands and what worked/didn't. Do **not** publish to any real customer-facing location.

### Step 6: Write the design doc + recommendation

In `docs/design/private-distribution.md`: the requirements, the options table (TS + Go), a single recommendation with rationale, the entitlement sketch, the POC results, and an ordered list of follow-up implementation plans (this spike's output feeds those).

## Deliverable / done criteria

- [ ] `docs/design/private-distribution.md` exists with: requirements, TS-registry options + trade-offs, Go-module options + trade-offs, an entitlement/token sketch, a clear recommendation, POC results, and a follow-up plan list.
- [ ] The POC commands are documented and were run in a throwaway/dry-run mode (nothing published to a real customer channel).
- [ ] Open business decisions (per-package vs per-bundle subscription, pricing) are listed as questions for the maintainer, not decided.
- [ ] `plans/README.md` row for 024 updated.

## STOP conditions

- The POC would require publishing to, or configuring, a real production/customer registry or creating real credentials — STOP; the spike must stay in dry-run/throwaway mode.
- A candidate registry requires a paid account the maintainer hasn't provisioned — document it as an option with the cost, don't sign up.

## Maintenance notes

- This spike's recommendation should spawn concrete implementation plans (publish pipeline, entitlement service); it deliberately doesn't build them.
- Revisit if the "no public registry" constraint changes.

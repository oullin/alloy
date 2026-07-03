# Contributing to Alloy

Alloy is a private Go + TypeScript workspace. This guide covers the daily
workflow; see [README.md](README.md) for the workspace layout.

## Requirements

- Node.js 20 or newer, pnpm 10.33.0
- Go 1.26.4 (for Go package checks)
- Docker or Docker Compose (formatting and Go test tasks are Docker-backed
  so local output matches CI)

## Setup

```sh
pnpm install
pnpm exec vp run monorepo:initialise   # creates packages/foundation/go.work
```

## Everyday tasks

Everything runs through Vite+ (`vp`), configured in [vite.config.ts](vite.config.ts):

```sh
pnpm exec vp check            # typecheck
pnpm exec vp test             # TypeScript tests (Vitest)
pnpm exec vp pack             # bundle packages
pnpm lint                     # oxlint + workspace import checks
pnpm exec vp run go:test      # Go vet + tests with -race (Docker-backed)
pnpm exec vp run format       # format changed files (Docker-backed)
pnpm exec vp run format-all   # format everything + vp check --fix
```

The `Makefile` forwards any target to `vp run` (e.g. `make go-test`).

## Conventions

- TypeScript packages live directly under `packages/<name>` — no language
  suffixes in paths.
- All `@alloy/*` packages are `"private": true` and are consumed as
  `workspace:*` dependencies or release tarballs — never published to a
  public registry (see [LICENSE.md](LICENSE.md)).
- Library code returns errors; panics are reserved for documented `Must*`
  helpers.
- New Go packages get a `doc.go` package comment and an `errors.go` with
  exported sentinel errors.
- Prefer adding tests next to the code (Go) or in the package's `tests/`
  workspace (TypeScript, run under `vp test`).

## Pull requests

- Branch from `main`; keep PRs scoped to one concern.
- CI detects changed surfaces (Go / TS) and runs only the relevant checks.
  The Inertia demo E2E suite runs when the `inertial-tests` label is applied.
- Run `pnpm exec vp run format` before pushing — formatting diffs fail CI.

## Releases

Releases are tag-driven:

- TypeScript: push a `ts/vX.Y.Z` tag (or run the `release-ts` workflow
  manually). Packages are built, packed to `.tgz`, and attached to a GitHub
  release.
- Go: push a `go/vX.Y.Z` tag (or run the `release-go` workflow manually).

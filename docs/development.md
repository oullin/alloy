# Development

This repository uses Vite+ as the primary orchestration layer for TypeScript,
Go, packaging, and custom maintenance tasks. The root `Makefile` stays in place
as a compatibility wrapper for existing `make` commands.

## Install

```sh
pnpm install
pnpm exec vp run monorepo:initialise
```

The initialise task runs `go work sync` inside `api/` and expects `api/go.work`
to describe the Go modules in the workspace.

## Common Commands

| Command | Purpose |
| --- | --- |
| `pnpm exec vp check` | Type-check and lint the TypeScript workspace. |
| `pnpm exec vp test` | Run TypeScript tests through Vite+. |
| `pnpm exec vp pack` | Build the TypeScript package output configured in `vite.config.ts`. |
| `pnpm exec vp run go:test` | Run Go vet and race-enabled tests for every module under `api/`. |
| `pnpm exec vp run format` | Format changed Go and TypeScript files with the formatter container. |
| `pnpm exec vp run format-all` | Format the repository and run `vp check --fix`. |

The root package scripts map to the common TypeScript workflows:

```sh
pnpm build
pnpm lint
pnpm test
pnpm typecheck
```

## Formatting

Formatting runs through the `fmt` service in `docker-compose.yml`, which uses
`ghcr.io/oullin/go-fmt:v0.4.1-full` by default. Keep `make format` and
`make format-all` working when changing automation; other tools can move, but
these entrypoints are part of the contributor contract.

The formatter task picks up changed or untracked Go and TypeScript files for
`format`, while `format-all` runs the full formatter container and then applies
Vite+ fixes.

## Go Workspace

The Go workspace lives under `api/`. Each API module owns its own `go.mod`, and
`api/go.work` ties those modules together for local checks.

`go:test` iterates over every module under `api/`, verifies it is using
`api/go.work`, then runs:

```sh
go vet ./...
go test -race ./...
```

## TypeScript Workspace

pnpm workspace packages are defined in `pnpm-workspace.yaml`. The main package
surfaces are:

- `@alloy/tempo` in `packages/tempo/tempo-ts`.
- `@alloy/tempo-acceptance` in `packages/tempo/tempo-ts/tests`.
- `@alloy/console` in `packages/console`.
- `@alloy/infra` in `infra`.

Path aliases and packaging are configured in `vite.config.ts`.

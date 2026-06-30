# Development

This repository uses Vite+ as the primary orchestration layer for TypeScript,
Go, packaging, and custom maintenance tasks. The root `Makefile` stays in place
as a compatibility wrapper for existing `make` commands.

## Install

```sh
pnpm install
pnpm exec vp run monorepo:initialise
```

The initialise task creates `go/go.work` from `go/go.work.example` when
needed, then runs `go work sync` inside `go/`.

## Common Commands

| Command | Purpose |
| --- | --- |
| `pnpm exec vp check` | Type-check and lint the TypeScript workspace. |
| `pnpm exec vp test` | Run TypeScript tests through Vite+. |
| `pnpm exec vp pack` | Build the TypeScript package output configured in `vite.config.ts`. |
| `pnpm exec vp run go:test` | Run Go vet and race-enabled tests for Go module(s) under `go/` and web demo APIs. |
| `pnpm exec vp run format` | Format changed Go and TypeScript files with the formatter container. |
| `pnpm exec vp run format-all` | Format the repository and run `vp check --fix`. |
| `pnpm web:build` | Build the VuePress documentation workspace. |
| `pnpm inertia-demo:build` | Build the Inertia demo frontend into `web/storage/inertia-demo`. |

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

The framework Go code lives under `go/`. The current public module path is
`alloy.dev/go`, declared in `go/go.mod`; `go/go.work` is optional local
workspace glue. The Inertia demo API lives under `web/inertia-demo/api` as
`alloy.dev/inertia-demo`.

`go:test` iterates over Go module(s) under `go/` and web demo APIs, uses
`go/go.work` for the framework module when it exists, then runs:

```sh
go vet ./...
go test -race ./...
```

## TypeScript Workspace

pnpm workspace packages are defined in `pnpm-workspace.yaml`. The main package
surfaces are:

- `@alloy/tempo` in `ts/tempo`.
- `@alloy/tempo-acceptance` in `ts/tempo/tests`.
- `@alloy/console` in `ts/console`.
- `@alloy/web` in `web`.
- `@alloy/inertia-demo-app` in `web/inertia-demo/app`.
- `@alloy/inertia-demo-e2e` in `web/inertia-demo/tests/e2e`.
- `@alloy/infra` in `infra`.

Package paths should match the package family directly; avoid language suffixes
such as `*-ts` in directory names. Acceptance-test packages live below the
package they validate, as `ts/tempo/tests` does for `@alloy/tempo`.

Path aliases and packaging are configured in `vite.config.ts`. The Tempo package
build uses `ts/tempo/src/index.ts`, `ts/tempo/tsconfig.json`, and
emits to `ts/tempo/dist`.

Tempo acceptance tests can be run through the package script:

```sh
pnpm --filter @alloy/tempo-acceptance test:tempo
```

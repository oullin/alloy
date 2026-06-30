# Alloy

Alloy is a Go and TypeScript workspace for reusable application primitives.
The repository currently includes the Go module under `go/`, the TypeScript
packages under `ts/`, the web workspace under `web/`, and the shared automation
under `infra/`.

Tempo is the most complete cross-runtime package in the workspace. It is
available as both a Go package (`alloy.dev/go/tempo`) and a
TypeScript package (`@alloy/tempo`).

## Workspace

- `go`: Go module containing the Alloy packages.
- `go/tempo`: Go Tempo package.
- `go/collection`: Go collection utilities and package docs.
- `ts/tempo`: TypeScript Tempo package.
- `ts/tempo/tests`: TypeScript Tempo acceptance tests.
- `ts/money`: TypeScript Money package.
- `ts/console`: TypeScript terminal UI helpers.
- `web`: VuePress documentation site package.
- `web/inertia-demo`: Inertia demo app, Go API, and browser E2E suite.
- `web/storage`: local runtime and cache data for web demos.
- `infra`: repo automation, cache paths, scripts, and shared TypeScript config.
- `vite.config.ts`: Vite+ orchestration for checks, tests, packaging, and custom tasks.
- `docker-compose.yml`: container definitions for Go execution and formatting.

TypeScript packages live directly at their package directory under `ts/`.
Do not add language suffixes such as `*-ts` to package paths.

## Requirements

- Node.js 20 or newer.
- pnpm 10.33.0.
- Go 1.26.4 for Go package checks.
- Docker or Docker Compose for the formatter tasks.

## Setup

```sh
pnpm install
pnpm exec vp run monorepo:initialise
```

`monorepo:initialise` creates and syncs the optional Go workspace at `go/go.work`.

## Checks

```sh
pnpm exec vp check
pnpm exec vp test
pnpm exec vp pack
pnpm exec vp run go:test
pnpm web:build
pnpm inertia-demo:build
pnpm exec vp run format-all
```

Tempo-specific acceptance coverage can also be run directly:

```sh
pnpm --filter @alloy/tempo-acceptance test:tempo
```

The root package scripts are aliases for the main TypeScript checks:

```sh
pnpm build
pnpm lint
pnpm test
pnpm typecheck
```

## Formatting

Formatting is intentionally Docker-backed so local output matches CI and the
shared formatter image. Use one of the stable entrypoints:

```sh
pnpm exec vp run format
pnpm exec vp run format-all
make format
make format-all
```

`format` targets changed Go and TypeScript files. `format-all` formats the full
repository with the formatter container and then runs `vp check --fix`.

## Automation

Vite+ owns orchestration in this repository. Custom task definitions live in
`vite.config.ts`; shell implementations for multi-step tasks live under
`infra/scripts/tasks`.

The root `Makefile` is a compatibility delegate for existing `make <target>`
commands. It forwards targets to Vite+ while preserving the Docker-backed
formatter commands.

## More Documentation

- [Development workflow](docs/development.md)
- [Go Tempo](go/tempo/README.md)
- [Go Collection](go/collection/README.md)
- [TypeScript Tempo](ts/tempo/README.md)

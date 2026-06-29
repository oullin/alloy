# Alloy

Alloy is a Go and TypeScript workspace for reusable application primitives.
The repository currently includes the Go backend module under `packages/backend/`, the
TypeScript packages under `packages/`, and the shared automation under
`infra/`.

Tempo is the most complete cross-runtime package in the workspace. It is
available as both a Go package (`alloy.dev/backend/tempo`) and a
TypeScript package (`@alloy/tempo`).

## Workspace

- `packages/backend`: Go module containing the Alloy backend packages.
- `packages/backend/tempo`: Go Tempo package.
- `packages/backend/collection`: Go collection utilities and package docs.
- `packages/tempo`: TypeScript Tempo package.
- `packages/tempo/tests`: TypeScript Tempo acceptance tests.
- `packages/money`: TypeScript Money package.
- `packages/console`: TypeScript terminal UI helpers.
- `infra`: repo automation, cache paths, scripts, and shared TypeScript config.
- `vite.config.ts`: Vite+ orchestration for checks, tests, packaging, and custom tasks.
- `docker-compose.yml`: container definitions for Go execution and formatting.

TypeScript packages live directly at their package directory under `packages/`.
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

`monorepo:initialise` syncs the root Go workspace at `go.work`.

## Checks

```sh
pnpm exec vp check
pnpm exec vp test
pnpm exec vp pack
pnpm exec vp run backend:test
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
- [Go Tempo](packages/backend/tempo/README.md)
- [Go Collection](packages/backend/collection/README.md)
- [TypeScript Tempo](packages/tempo/README.md)

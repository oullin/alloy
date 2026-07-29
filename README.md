# Alloy

Alloy is a Go and TypeScript workspace for reusable application primitives.
The repository includes the Go packages under `pkg/hub`, the TypeScript
packages under `sdk/`, the web workspace under `web/`, and the shared
automation under `infra/`.

Tempo is the most complete cross-runtime package in the workspace. It is
available as both a Go package (`hara.sh/alloy/tempo`) and a
TypeScript package (`@hara/sdk-tempo`). Only a deliberate few packages are
cross-runtime twins; the [Cross-Runtime Parity Matrix](web/docs/architecture/parity.md)
maps which primitives exist in which runtime, how much parity each twin
guarantees, and the policy for what earns a twin.

## Workspace

- `pkg/hub`: Go module containing the Alloy packages.
- `pkg/hub/tempo`: Go Tempo package.
- `pkg/hub/collection`: Go collection utilities and package docs.
- `sdk/tempo`: TypeScript Tempo package.
- `sdk/tempo/tests`: TypeScript Tempo acceptance tests.
- `sdk/money`: TypeScript Money package.
- `sdk/console`: TypeScript terminal UI helpers.
- `web/docs`: VuePress documentation site package.
- `web/inertia-demo`: Inertia demo app, Go API, and browser E2E suite.
- `web/storage`: local runtime and cache data for web demos.
- `infra`: repo automation, cache paths, scripts, and shared TypeScript config.
- `vite.config.ts`: Vite+ orchestration for checks, tests, packaging, and custom tasks.
- `docker-compose.yml`: container definitions for Go execution and formatting.

TypeScript packages live directly at their package directory under `sdk/`.
Do not add language suffixes such as `*-ts` to package paths.

## Requirements

- Node.js 24 or newer.
- pnpm 10.33.0.
- Go 1.26.5 for Go package checks.
- [fmtkit](https://github.com/oullin/fmtkit) for the formatter tasks
  (`brew install --cask fmtkit`).

## Setup

```sh
pnpm install
pnpm exec vp run monorepo:initialise
```

`monorepo:initialise` creates and syncs the optional Go workspace at `pkg/hub/go.work`.

## Checks

```sh
pnpm exec vp check
pnpm exec vp test
pnpm -r --filter './sdk/*' build
pnpm exec vp run go:test
pnpm docs:build
pnpm inertia-demo:build
pnpm exec vp run format-all
```

Tempo-specific acceptance coverage can also be run directly:

```sh
pnpm --filter @hara/sdk-tempo-tests test:tempo
```

The root package scripts are aliases for the main TypeScript checks:

```sh
pnpm build
pnpm lint
pnpm test
pnpm typecheck
```

## Formatting

Formatting runs through [fmtkit](https://github.com/oullin/fmtkit), a single
binary that formats both Go and TS/Vue with no image pull and no daemon. Use
one of the stable entrypoints:

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
commands. It forwards targets to Vite+ while preserving the formatter
commands.

## More Documentation

- [Development workflow](CONTRIBUTING.md)
- [Go Tempo](pkg/hub/tempo/README.md)
- [Go Collection](pkg/hub/collection/README.md)
- [TypeScript Tempo](sdk/tempo/README.md)

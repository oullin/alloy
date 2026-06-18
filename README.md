# Tempo

Tempo is a Go and TypeScript date/time library.

The public API is branded as Tempo across both runtimes.

## Workspace

- `packages/tempo-ts`: TypeScript package.
- `packages/tempo-go`: Go module.
- `packages/tempo-ts/tests`: TypeScript acceptance tests.
- `infra`: repo automation, Task scripts, cache paths, and TypeScript config exports.

## Checks

```sh
pnpm install
pnpm exec task monorepo:initialise
pnpm exec task typecheck
pnpm exec task test
pnpm exec task build
pnpm exec task go:test
pnpm exec task format-all
```

The root `Makefile` is a compatibility delegate for existing `make <target>`
commands. Task owns orchestration, and task command bodies live under
`infra/scripts/tasks`.

# Tempo

Tempo is a Go and TypeScript date/time library.

The public API is branded as Tempo across both runtimes.

## Workspace

- `packages/tempo/tempo-ts`: TypeScript package.
- `packages/tempo/tempo-ts/tests`: TypeScript acceptance tests.
- `api`: Go workspace containing all API modules.
- `api/tempo`: Go Tempo module.
- `api/routing/generator`: TypeScript route helper generator folded into the routing module.
- `infra`: repo automation, Task scripts, cache paths, and TypeScript config exports.

## Checks

```sh
pnpm install
vp run monorepo:initialise
vp check
vp test
vp pack
vp run go:test
vp run format-all
```

The root `Makefile` is a compatibility delegate for existing `make <target>`
commands. Vite+ owns orchestration, and task command bodies live under
`vite.config.ts` or `infra/scripts/tasks`.

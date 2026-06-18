# Tempo

Tempo is a Go and TypeScript date/time library.

The public API is branded as Tempo across both runtimes.

## Workspace

- `packages/tempo-ts`: TypeScript package.
- `packages/tempo-go`: Go module.
- `packages/tempo-ts/tests`: TypeScript acceptance tests.
- `infra`: repo automation, Make targets, cache paths, and TypeScript config exports.

## Checks

```sh
pnpm install
pnpm typecheck
pnpm test
pnpm build
make go-test
make format-all
```

# Tempo

Tempo is a Go and TypeScript date/time library.

The public API is branded as Tempo across both runtimes.

## Workspace

- `packages/tempo/ts`: TypeScript package.
- `packages/tempo/go`: Go module.
- `packages/tests`: shared TypeScript acceptance tests.
- `packages/artefacts`: shared repo artifact paths and TypeScript config exports.
- `provision/tooling`: repo automation, Make targets, and validation scripts.

## Checks

```sh
pnpm install
pnpm typecheck
pnpm test
pnpm build
make go-test
make format-all
```

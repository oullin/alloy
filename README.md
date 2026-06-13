# Tempo

Tempo is a Go and TypeScript date/time library being ported from Carbon `v3.11.4`.

The public API is branded as Tempo. Carbon remains the pinned behavioral oracle for generated fixtures and compatibility tests.

## Workspace

- `packages/tempo/ts`: TypeScript package.
- `packages/tempo/go`: Go module.
- `packages/tempo/spec`: shared fixtures and API mapping notes.
- `packages/tests`: shared TypeScript acceptance tests.
- `packages/artefacts`: shared repo artifact paths and TypeScript config exports.
- `packages/carbon-oracle`: TypeScript fixture generator for Carbon compatibility cases.
- `provision`: repo automation, Make targets, and validation scripts.

## Checks

```sh
pnpm install
pnpm typecheck
pnpm test
pnpm build
make go-test
make oracle-check
make format-all
```

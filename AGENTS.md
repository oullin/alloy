# Agent guide

Conventions an agent or contributor needs before changing this repo. See
[README.md](README.md) for the workspace layout and [CONTRIBUTING.md](CONTRIBUTING.md)
for the full workflow; this file only captures the non-obvious parts.

## Toolchain

- Node.js 24 or newer, pnpm 10.33.0 (`engine-strict=true`, so the version is enforced).
- Go 1.26.5 for Go package checks.
- [fmtkit](https://github.com/oullin/fmtkit) for formatting: `brew install --cask fmtkit`.
- Docker or Docker Compose — the Go test task is container-backed so local
  output matches CI.

## Run everything through Vite+ (`vp`)

Vite+ owns orchestration; task definitions live in `vite.config.ts`. Do not call
`tsc`, `vitest`, or `go test` directly — go through `vp`:

```sh
pnpm exec vp check          # typecheck
pnpm exec vp test           # TypeScript tests (Vitest)
pnpm -r --filter './sdk/*' build  # build publishable packages
pnpm lint                   # oxlint + workspace import checks
pnpm exec vp run go:test    # Go vet + tests with -race (Docker-backed)
pnpm exec vp run format     # format changed files (fmtkit)
pnpm exec vp run format-all # format everything + vp check --fix
pnpm docs:build             # build the VuePress docs site
```

Formatting runs through fmtkit, a single binary that covers both Go and
TS/Vue. Run `pnpm exec vp run format` before pushing — formatting diffs fail CI.

## Layout

- `pkg/hub` — the Go module holding the Alloy packages.
- Nested Go modules live at `pkg/hub/queue/drivers/sqs` and `pkg/hub/auth/passkeys`;
  they have their own `go.mod` and need `go mod tidy` when `pkg/hub` deps move.
- `sdk/*` — TypeScript packages (`console`, `money`, `navigator-routes`, `tempo`,
  `workflow`). No language suffixes such as `*-ts` in paths.
- `web/inertia-demo` — Inertia demo app, Go API, and browser E2E suite.
- `web/docs` — the VuePress documentation site.
- `infra` — repo automation, scripts, and shared TypeScript config.

## Package conventions

- New Go packages get a `doc.go` package comment and an `errors.go` with exported
  sentinel errors.
- Library code returns errors; panics are reserved for documented `Must*` helpers.
- All `@alloy/*` packages are `"private": true` and are consumed as `workspace:*`
  dependencies or release tarballs — never published to a public registry.
- Add tests next to the code (Go) or in the package's `tests/` workspace
  (TypeScript, run under `vp test`).

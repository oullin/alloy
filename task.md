# Vite+ Migration Tasks

- [x] Execute `vp migrate --no-interactive`
- [x] Merge and consolidate root `vite.config.ts`
  - [x] Add resolve aliases and cache directory
  - [x] Add `go:test`, `format`, and `format-all` tasks
- [x] Migrate local configurations
  - [x] Migrate `packages/console` to use `vite.config.ts` and `vite-plus`
  - [x] Migrate `packages/tempo/tempo-ts` to use `vite.config.ts` with `pack` config
  - [x] Delete `tsdown.config.ts`
- [x] Clean up legacy build tools and configs
  - [x] Delete `Taskfile.yml`
  - [x] Delete `turbo.json`
  - [x] Update root `package.json` scripts and devDependencies
  - [x] Clean up workspace package.json files
- [x] Reinstall dependencies cross-platform
- [x] Update repo entrypoints (workflows, README)
- [x] Run verification tests

# Alloy Inertia E2E

This suite runs the imported Inertia demo against the Alloy Go API and drives it through Agent Browser using Helium.

Artifacts are written outside the repository:

```sh
/Users/gocanto/.cache/codex/browser-artifacts/alloy-inertia/<run-id>
```

Run the Alloy suite:

```sh
pnpm test:e2e:inertia
```

Run the same flow against the source Bedrock demo when comparing parity:

```sh
pnpm --filter @alloy/inertia-e2e test:bedrock
```

The runner requires Helium at `HELIUM_EXECUTABLE_PATH`, `AGENT_BROWSER_EXECUTABLE_PATH`, or `/Applications/Helium.app/Contents/MacOS/Helium`.

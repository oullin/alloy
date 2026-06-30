# Alloy Inertia E2E

This suite runs the Inertia demo against the Alloy Go API and drives it through Agent Browser with Helium. The Playwright runner remains available only for explicit legacy checks.

Artifacts are written outside the repository:

```sh
/Users/gocanto/.cache/codex/browser-artifacts/alloy-inertia-demo/<run-id>
```

Run the Alloy suite:

```sh
pnpm test:e2e:inertia
```

Run the local Agent Browser + Helium suite:

```sh
pnpm --filter @alloy/inertia-demo-e2e test:alloy:agent-browser
```

Run the same flow against the source Bedrock demo when comparing parity:

```sh
pnpm --filter @alloy/inertia-demo-e2e test:bedrock
```

The Playwright runner uses `PLAYWRIGHT_EXECUTABLE_PATH`, `CHROME_EXECUTABLE_PATH`, `GOOGLE_CHROME_BIN`, or `CHROMIUM_BIN` when set. In CI, `.github/workflows/ci-inertia-e2e.yml` installs Chromium with `playwright install --with-deps chromium` and stores artifacts under the runner temp directory before uploading them.

The Agent Browser runner uses `HELIUM_EXECUTABLE_PATH`, `AGENT_BROWSER_EXECUTABLE_PATH`, or `/Applications/Helium.app/Contents/MacOS/Helium`.

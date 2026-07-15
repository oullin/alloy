# Alloy Inertia E2E

This suite runs the Inertia demo against the Alloy Go API and drives it through
the browser. The canonical, portable runner uses **Playwright/Chromium** and
needs no machine-specific browser install — it is what CI runs. An Agent Browser
(Helium) runner remains available as an opt-in alternative.

Both runners share their flows, route matrix and server lifecycle via
`shared.mjs`; each runner is a thin driver adapter. Adding or changing a route or
flow is a one-file edit in `shared.mjs`.

## Run the canonical (Playwright) suite

```sh
pnpm --filter @alloy/inertia-demo-e2e test
```

The first run needs the Chromium build Playwright manages:

```sh
pnpm exec vp exec -F @alloy/inertia-demo-e2e -- playwright install --with-deps chromium
```

Artifacts (screenshots, traces, logs, database) are written outside the
repository, under an OS temp directory by default:

```sh
$TMPDIR/alloy-inertia-e2e/browser-artifacts/alloy-inertia-demo/<run-id>
```

## Run the opt-in Agent Browser (Helium) suite

```sh
HELIUM_EXECUTABLE_PATH=/path/to/Helium pnpm --filter @alloy/inertia-demo-e2e test:alloy:agent-browser
```

## Run the Bedrock parity target

Compares the same flows against a local Bedrock demo checkout:

```sh
BEDROCK_INERTIA_SOURCE=/path/to/bedrock/services/demo/inertia pnpm --filter @alloy/inertia-demo-e2e test:bedrock
```

## Environment variables

| Variable | Applies to | Default | Purpose |
|----------|-----------|---------|---------|
| `CODEX_BROWSER_ARTIFACTS_DIR` | both runners | `$TMPDIR/alloy-inertia-e2e/browser-artifacts` | Base directory for run artifacts. CI points this at the runner temp dir. |
| `PLAYWRIGHT_HEADED` | Playwright | `true` locally, `false` in CI | `true` runs headed Chromium; anything else runs headless. |
| `PLAYWRIGHT_EXECUTABLE_PATH` / `CHROME_EXECUTABLE_PATH` / `GOOGLE_CHROME_BIN` / `CHROMIUM_BIN` | Playwright | unset (use the Playwright-managed Chromium) | Absolute path to an alternative Chromium/Chrome binary. First set wins. |
| `PLAYWRIGHT_BROWSER_ARGS` (or `AGENT_BROWSER_ARGS`) | Playwright | none | Comma-separated extra Chromium launch args (CI uses `--no-sandbox,--disable-dev-shm-usage`). |
| `PLAYWRIGHT_BROWSERS_PATH` | Playwright | Playwright default | Where Playwright stores/looks up its browser builds. |
| `HELIUM_EXECUTABLE_PATH` (or `AGENT_BROWSER_EXECUTABLE_PATH`) | Agent Browser | a Helium install under `~/Applications`, if present | Path to the Helium binary. Required when Helium lives elsewhere; the runner fails fast naming this variable when it cannot find Helium. |
| `AGENT_BROWSER_BIN` | Agent Browser | `agent-browser` (on `PATH`) | Path to the `agent-browser` CLI. |
| `AGENT_BROWSER_HEADED` | Agent Browser | `true` | `true` runs headed Helium; anything else runs headless. |
| `BEDROCK_INERTIA_SOURCE` | `--target bedrock` | unset (required for bedrock) | Path to a local Bedrock inertia demo checkout (e.g. `<bedrock>/services/demo/inertia`). The runner fails fast when unset. |

In CI, `.github/workflows/ci-inertia-app-tests.yml` installs Chromium with
`playwright install --with-deps chromium` and stores artifacts under the runner
temp directory before uploading them.

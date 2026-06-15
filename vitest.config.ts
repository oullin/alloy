import { fileURLToPath, URL } from "node:url";
import { defineConfig } from "vitest/config";

const repoPath = (path: string): string =>
  fileURLToPath(new URL(path, import.meta.url));

export default defineConfig({
  cacheDir: fileURLToPath(
    new URL("./packages/artefacts/.cache/vitest", import.meta.url),
  ),
  resolve: {
    alias: {
      "@tempo/artefacts": repoPath("./packages/artefacts/src"),
      "@alloy/tempo": repoPath("./packages/tempo/ts/src"),
      "@tempo/tests": repoPath("./packages/tests/src"),
    },
  },
  test: {
    passWithNoTests: true,
  },
});

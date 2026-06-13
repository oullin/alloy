import { fileURLToPath, URL } from "node:url";
import { workspaceAliases } from "@tempo/artefacts";
import { defineConfig } from "vitest/config";

export default defineConfig({
  cacheDir: fileURLToPath(
    new URL("./packages/artefacts/.cache/vitest", import.meta.url),
  ),
  resolve: {
    alias: workspaceAliases(),
  },
  test: {
    passWithNoTests: true,
  },
});

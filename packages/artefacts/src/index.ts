import { resolve } from "node:path";
import { fileURLToPath } from "node:url";

export const artefactsRoot = resolve(
  fileURLToPath(new URL("..", import.meta.url)),
);
export const repoRoot = resolve(artefactsRoot, "..", "..");

export const artefactsPath = (...segments: string[]): string =>
  resolve(artefactsRoot, ...segments);
export const cacheDir = (...segments: string[]): string =>
  artefactsPath(".cache", ...segments);
export const distDir = (...segments: string[]): string =>
  artefactsPath("dist", ...segments);
export const logDir = (...segments: string[]): string =>
  artefactsPath(".logs", ...segments);

const packageRoot = resolve(repoRoot, "packages");
export const workspaceAliases = (): Record<string, string> => ({
  "@tempo/artefacts": resolve(packageRoot, "artefacts", "src"),
  "@tempo/tempo": resolve(packageRoot, "tempo", "ts", "src"),
  "@tempo/tests": resolve(packageRoot, "tests", "src"),
});

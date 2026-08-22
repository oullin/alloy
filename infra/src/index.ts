import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

export const infraRoot = resolve(fileURLToPath(new URL('..', import.meta.url)));

export const repoRoot = resolve(infraRoot, '..');

export const infraPath = (...segments: string[]): string => resolve(infraRoot, ...segments);

export const cacheDir = (...segments: string[]): string => infraPath('.cache', ...segments);

export const distDir = (...segments: string[]): string => infraPath('dist', ...segments);

export const logDir = (...segments: string[]): string => infraPath('.logs', ...segments);

const sdkRoot = resolve(repoRoot, 'sdk');

export const workspaceAliases = (): Record<string, string> => ({
	'@alloy/infra': resolve(infraRoot, 'src'),
	'@hara/sdk-container': resolve(sdkRoot, 'container', 'src'),
	'@hara/sdk-httpx': resolve(sdkRoot, 'httpx', 'src'),
	'@hara/sdk-tempo': resolve(sdkRoot, 'tempo', 'src'),
	'@hara/sdk-tempo-tests': resolve(sdkRoot, 'tempo', 'tests', 'src'),
	'@hara/sdk-money': resolve(sdkRoot, 'money', 'src'),
	'@hara/sdk-console': resolve(sdkRoot, 'console', 'src'),
	'@hara/sdk-navigator-routes': resolve(sdkRoot, 'navigator-routes', 'src'),
	'@hara/sdk-workflow': resolve(sdkRoot, 'workflow', 'src'),
});

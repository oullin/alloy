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
	'@alloy/sdk/tempo': resolve(sdkRoot, 'tempo', 'src'),
	'@alloy/sdk/tempo-tests': resolve(sdkRoot, 'tempo', 'tests', 'src'),
	'@alloy/sdk/money': resolve(sdkRoot, 'money', 'src'),
	'@alloy/sdk/console': resolve(sdkRoot, 'console', 'src'),
	'@alloy/sdk/navigator-routes': resolve(sdkRoot, 'navigator-routes', 'src'),
	'@alloy/sdk/workflow': resolve(sdkRoot, 'workflow', 'src'),
});

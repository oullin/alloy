import { resolve } from 'node:path';
import { fileURLToPath } from 'node:url';

export const infraRoot = resolve(fileURLToPath(new URL('..', import.meta.url)));

export const repoRoot = resolve(infraRoot, '..');

export const infraPath = (...segments: string[]): string => resolve(infraRoot, ...segments);

export const cacheDir = (...segments: string[]): string => infraPath('.cache', ...segments);

export const distDir = (...segments: string[]): string => infraPath('dist', ...segments);

export const logDir = (...segments: string[]): string => infraPath('.logs', ...segments);

const packageRoot = resolve(repoRoot, 'packages');

export const workspaceAliases = (): Record<string, string> => ({
	'@alloy/infra': resolve(infraRoot, 'src'),
	'@alloy/tempo': resolve(packageRoot, 'tempo', 'src'),
	'@alloy/tempo-tests': resolve(packageRoot, 'tempo', 'tests', 'src'),
	'@alloy/console': resolve(packageRoot, 'console', 'src'),
	'@alloy/expose-routes': resolve(packageRoot, 'expose-routes', 'src'),
});

import { fileURLToPath, URL } from 'node:url';
import { defineConfig } from 'vitest/config';

const repoPath = (path: string): string => fileURLToPath(new URL(path, import.meta.url));

export default defineConfig({
	cacheDir: fileURLToPath(new URL('./infra/.cache/vitest', import.meta.url)),
	resolve: {
		alias: {
			'@alloy/infra': repoPath('./infra/src'),
			'@alloy/tempo': repoPath('./packages/tempo/tempo-ts/src'),
			'@alloy/tempo-tests': repoPath('./packages/tempo/tempo-ts/tests/src'),
			'@alloy/console': repoPath('./packages/console/src'),
		},
	},
	test: {
		passWithNoTests: true,
	},
});

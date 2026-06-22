import { dirname, join } from 'node:path';
import { defineConfig } from 'vitest/config';
import { fileURLToPath } from 'node:url';

const packagePath = dirname(fileURLToPath(import.meta.url));
const workspacePath = dirname(dirname(packagePath));
const cachePath = join(workspacePath, 'infra', '.cache', 'vitest', 'console');
const sourcePath = join(packagePath, 'src', '$1');

export default defineConfig({
	cacheDir: cachePath,
	resolve: {
		alias: [
			{
				find: /^#console\/(.+)$/u,
				replacement: sourcePath,
			},
		],
	},
	test: {
		environment: 'node',
		globals: false,
	},
});

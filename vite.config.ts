import { fileURLToPath, URL } from 'node:url';
import { defineConfig } from 'vite-plus';

const repoPath = (path: string): string => fileURLToPath(new URL(path, import.meta.url));

export default defineConfig({
	cacheDir: fileURLToPath(new URL('./infra/.cache/vitest', import.meta.url)),
	resolve: {
		alias: [
			{
				find: '@alloy/infra',
				replacement: repoPath('./infra/src'),
			},
			{
				find: '@alloy/tempo',
				replacement: repoPath('./packages/tempo/tempo-ts/src'),
			},
			{
				find: '@alloy/tempo-tests',
				replacement: repoPath('./packages/tempo/tempo-ts/tests/src'),
			},
			{
				find: '@alloy/console',
				replacement: repoPath('./packages/console/src'),
			},
			{
				find: /^#console\/(.+)$/u,
				replacement: repoPath('./packages/console/src/$1'),
			},
		],
	},
	test: {
		passWithNoTests: true,
		environment: 'node',
		globals: false,
	},
	pack: {
		entry: [repoPath('./packages/tempo/tempo-ts/src/index.ts')],
		tsconfig: './packages/tempo/tempo-ts/tsconfig.json',
		outDir: './packages/tempo/tempo-ts/dist',
		dts: true,
		format: ['esm'],
		clean: true,
	},
	lint: {
		jsPlugins: [{ name: 'vite-plus', specifier: 'vite-plus/oxlint-plugin' }],
		rules: { 'vite-plus/prefer-vite-plus-imports': 'error' },
		options: { typeAware: true, typeCheck: true },
	},
	fmt: {
		singleQuote: true,
		useTabs: true,
		printWidth: 190,
		ignorePatterns: ['**/*.md', '**/*.json', '**/*.yml', '**/*.yaml', '**/*.css', '**/*.html', '.agents/**', 'dist/**', 'infra/.cache/**'],
	},
	run: {
		tasks: {
			format: { command: 'bash infra/scripts/tasks/format-files.sh changed', cache: false },
			'format-all': { command: 'bash infra/scripts/tasks/docker-compose-run.sh fmt format-all && vp fmt', cache: false },
			'go:test': { command: 'bash infra/scripts/tasks/go-test.sh', cache: false },
			'monorepo:initialise': { command: 'bash infra/scripts/tasks/monorepo-initialise.sh', cache: false },
		},
	},
});

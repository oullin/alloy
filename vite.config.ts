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
				replacement: repoPath('./packages/tempo/src'),
			},
			{
				find: '@alloy/money',
				replacement: repoPath('./packages/money/src'),
			},
			{
				find: /^#money\/(.+)$/u,
				replacement: repoPath('./packages/money/src/$1'),
			},
			{
				find: '@alloy/tempo-tests',
				replacement: repoPath('./packages/tempo/tests/src'),
			},
			{
				find: '@alloy/console',
				replacement: repoPath('./packages/console/src'),
			},
			{
				find: '@alloy/workflow',
				replacement: repoPath('./packages/workflow/src'),
			},
			{
				find: /^@alloy\/workflow\/(.+)$/u,
				replacement: repoPath('./packages/workflow/src/$1'),
			},
			{
				find: /^#workflow\/(.+)$/u,
				replacement: repoPath('./packages/workflow/src/$1'),
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
		coverage: {
			reportsDirectory: repoPath('./infra/.cache/vitest/coverage'),
		},
	},
	pack: {
		entry: [repoPath('./packages/tempo/src/index.ts')],
		tsconfig: './packages/tempo/tsconfig.json',
		outDir: './packages/tempo/dist',
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
			'cache:setup': { command: 'bash infra/scripts/tasks/cache-setup.sh', cache: false },
			format: { command: 'bash infra/scripts/tasks/format-files.sh changed', cache: false },
			'format-all': { command: "bash infra/scripts/tasks/docker-compose-run.sh fmt format-all && bash -lc 'source infra/scripts/tasks/cache-env.sh && vp check --fix'", cache: false },
			'go:test': { command: 'bash infra/scripts/tasks/go-test.sh', cache: false },
			'monorepo:initialise': { command: 'bash infra/scripts/tasks/monorepo-initialise.sh', cache: false },
		},
	},
});

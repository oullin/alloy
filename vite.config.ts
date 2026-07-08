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
				find: '@alloy/sdk/tempo',
				replacement: repoPath('./sdk/tempo/src'),
			},
			{
				find: '@alloy/sdk/money',
				replacement: repoPath('./sdk/money/src'),
			},
			{
				find: /^#money\/(.+)$/u,
				replacement: repoPath('./sdk/money/src/$1'),
			},
			{
				find: '@alloy/sdk/tempo-tests',
				replacement: repoPath('./sdk/tempo/tests/src'),
			},
			{
				find: '@alloy/sdk/console',
				replacement: repoPath('./sdk/console/src'),
			},
			{
				find: '@alloy/sdk/workflow',
				replacement: repoPath('./sdk/workflow/src'),
			},
			{
				find: /^@alloy\/sdk\/workflow\/(.+)$/u,
				replacement: repoPath('./sdk/workflow/src/$1'),
			},
			{
				find: /^#workflow\/(.+)$/u,
				replacement: repoPath('./sdk/workflow/src/$1'),
			},
			{
				find: /^#console\/(.+)$/u,
				replacement: repoPath('./sdk/console/src/$1'),
			},
		],
	},
	test: {
		passWithNoTests: true,
		environment: 'node',
		globals: false,
		coverage: {
			reportsDirectory: repoPath('./infra/.cache/vitest/coverage'),
			reporter: ['text-summary', 'html', 'json-summary'],
		},
	},
	pack: {
		entry: [repoPath('./sdk/tempo/src/index.ts')],
		tsconfig: './sdk/tempo/tsconfig.json',
		outDir: './sdk/tempo/dist',
		dts: true,
		format: ['esm'],
		clean: true,
	},
	lint: {
		jsPlugins: [{ name: 'vite-plus', specifier: 'vite-plus/oxlint-plugin' }],
		rules: {
			'typescript/no-useless-default-assignment': 'off',
			'vite-plus/prefer-vite-plus-imports': 'error',
		},
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
			'check:imports': { command: 'vp exec node infra/scripts/tasks/check-package-imports.mjs', cache: false },
			'inertia-demo-app:build': { command: 'vp build --config web/inertia-demo/app/vite.config.js', cache: false },
			'inertia-demo-app:dev': { command: 'vp dev --config web/inertia-demo/app/vite.config.js', cache: false },
			'inertia-demo:e2e': { command: 'vp run inertia-demo:e2e:agent-browser', cache: false },
			'inertia-demo:e2e:agent-browser': { command: 'vp run inertia-demo-app:build && vp exec node web/inertia-demo/tests/e2e/runner.mjs --target alloy', cache: false },
			'inertia-demo:e2e:playwright': { command: 'vp run inertia-demo-app:build && vp exec node web/inertia-demo/tests/e2e/runner.playwright.mjs --target alloy', cache: false },
			'inertia-demo:e2e:bedrock': { command: 'vp exec node web/inertia-demo/tests/e2e/runner.mjs --target bedrock --build-app', cache: false },
			'inertia-app:build': { command: 'vp run inertia-demo-app:build', cache: false },
			'inertia-app:dev': { command: 'vp run inertia-demo-app:dev', cache: false },
			'inertia:e2e': { command: 'vp run inertia-demo:e2e', cache: false },
			'inertia:e2e:agent-browser': { command: 'vp run inertia-demo:e2e:agent-browser', cache: false },
			'inertia:e2e:playwright': { command: 'vp run inertia-demo:e2e:playwright', cache: false },
			'inertia:e2e:bedrock': { command: 'vp run inertia-demo:e2e:bedrock', cache: false },
			'monorepo:initialise': { command: 'bash infra/scripts/tasks/monorepo-initialise.sh', cache: false },
		},
	},
});

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
				find: '@hara/sdk-tempo',
				replacement: repoPath('./sdk/tempo/src'),
			},
			{
				find: '@hara/sdk-money',
				replacement: repoPath('./sdk/money/src'),
			},
			{
				find: /^#money\/(.+)$/u,
				replacement: repoPath('./sdk/money/src/$1'),
			},
			{
				find: '@hara/sdk-tempo-tests',
				replacement: repoPath('./sdk/tempo/tests/src'),
			},
			{
				find: '@hara/sdk-console',
				replacement: repoPath('./sdk/console/src'),
			},
			{
				find: '@hara/sdk-workflow',
				replacement: repoPath('./sdk/workflow/src'),
			},
			{
				find: /^@hara\/sdk-workflow\/(.+)$/u,
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
			{
				find: /^#navigator-routes\/(.+)$/u,
				replacement: repoPath('./sdk/navigator-routes/src/$1'),
			},
			// Tempo's internal specifiers are bare (`#core`, not `#tempo/core`).
			// The package `imports` map resolves them to `dist` for consumers; these
			// aliases keep the test run on `src` so it never depends on a prior build.
			{
				find: /^#types$/u,
				replacement: repoPath('./sdk/tempo/src/types.ts'),
			},
			{
				find: /^#(calendar|config|core|duration|factory|formatting|parsing|ranges|runtime)$/u,
				replacement: repoPath('./sdk/tempo/src/$1/index.ts'),
			},
		],
	},
	test: {
		passWithNoTests: true,
		// Agent worktrees and pnpm's project-store symlinks can mirror test files
		// under git-excluded paths that do not belong to the working tree.
		exclude: ['**/node_modules/**', '**/dist/**', '**/.claude/worktrees/**', '**/.agents/**', '**/infra/.cache/**'],
		environment: 'node',
		globals: false,
		coverage: {
			reportsDirectory: repoPath('./infra/.cache/vitest/coverage'),
			reporter: ['text-summary', 'html', 'json-summary'],
		},
	},
	lint: {
		jsPlugins: [{ name: 'vite-plus', specifier: 'vite-plus/oxlint-plugin' }],
		rules: {
			'typescript/no-useless-default-assignment': 'off',
			'vite-plus/prefer-vite-plus-imports': 'error',
		},
		options: { typeAware: true, typeCheck: true },
		ignorePatterns: ['.agents/**', '.claude/**'],
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
			'format-all': { command: "bash infra/scripts/tasks/format-files.sh all && bash -lc 'source infra/scripts/tasks/cache-env.sh && vp check --fix'", cache: false },
			'go:test': { command: 'bash infra/scripts/tasks/go-test.sh', cache: false },
			'go:artifacts': { command: 'bash infra/scripts/tasks/check-go-artifacts.sh', cache: false },
			'check:imports': { command: 'vp exec node infra/scripts/tasks/check-package-imports.mjs', cache: false },
			'inertia-demo-app:build': { command: 'vp build --config web/inertia-demo/app/vite.config.js', cache: false },
			'inertia-demo-app:dev': { command: 'vp dev --config web/inertia-demo/app/vite.config.js', cache: false },
			'inertia-demo:e2e': { command: 'vp run inertia-demo:e2e:playwright', cache: false },
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

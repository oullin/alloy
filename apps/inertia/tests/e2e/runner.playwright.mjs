#!/usr/bin/env node
import { spawn, spawnSync } from 'node:child_process';
import { accessSync, appendFileSync, constants, createWriteStream, existsSync, mkdirSync, statSync, writeFileSync } from 'node:fs';
import http from 'node:http';
import net from 'node:net';
import path from 'node:path';
import process from 'node:process';
import { setTimeout as delay } from 'node:timers/promises';
import { fileURLToPath } from 'node:url';
import { chromium } from 'playwright';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, '../../../..');
const defaultArtifactsPath = '/Users/gocanto/.cache/codex/browser-artifacts';
const target = option('target') ?? 'alloy';
const buildApp = hasFlag('build-app');
const headed = option('headed') ?? process.env.PLAYWRIGHT_HEADED ?? (isCI() ? 'false' : 'true');

if (!['alloy', 'bedrock'].includes(target)) {
	fail(`Unknown target "${target}". Expected "alloy" or "bedrock".`);
}

const runId = `${new Date().toISOString().replaceAll(/[:.]/g, '-')}-${process.pid}`;
const artifactsRoot = path.resolve(process.env.CODEX_BROWSER_ARTIFACTS_DIR ?? defaultArtifactsPath, 'alloy-inertia', runId);
const screenshotsPath = path.join(artifactsRoot, 'screenshots');
const tracesPath = path.join(artifactsRoot, 'traces');
const logsPath = path.join(artifactsRoot, 'logs');
const databasePath = path.join(artifactsRoot, 'database');

for (const directory of [artifactsRoot, screenshotsPath, tracesPath, logsPath, databasePath]) {
	mkdirSync(directory, { recursive: true });
}

let serverProcess;
let serverExited = false;
let serverLogStreams = [];
let browser;
let context;
let page;
let pageIssues = [];

try {
	console.log(`Artifacts: ${artifactsRoot}`);
	console.log(`Playwright headed: ${headed}`);

	const port = await freePort();
	const app = await targetConfig(port);

	serverProcess = await startServer(app);
	await launchBrowser();
	await page.goto(`${app.baseUrl}/login`, { waitUntil: 'domcontentloaded' });

	await step('auth flow', () => runAuthFlow(app.baseUrl));
	await step('crm flow', () => runCrmFlow(app.baseUrl));
	await step('organization flow', () => runOrganizationFlow(app.baseUrl));
	await step('feature interactions', () => runFeatureInteractions(app.baseUrl));
	await step('route crawl', () => runRouteCrawl(app.baseUrl));

	await stopTrace('complete');
	await closeBrowser();
	await stopServer();

	console.log(`Inertia Playwright E2E passed for ${target}.`);
	console.log(`Artifacts: ${artifactsRoot}`);
	process.exit(0);
} catch (error) {
	console.error(error instanceof Error ? error.message : error);
	await failureArtifacts().catch(() => {});
	await closeBrowser().catch(() => {});
	await stopServer().catch(() => {});
	process.exit(1);
}

async function runAuthFlow(baseUrl) {
	clearBrowserIssues();
	await page.goto(`${baseUrl}/dashboard`, { waitUntil: 'domcontentloaded' });
	await waitForSelector('#email');
	await screenshot('auth-login');

	await fill('#email', 'wrong@example.test');
	await fill('#password', 'incorrect');
	await submitNearestForm('#email');
	await waitForText('Use test@example.com and password to sign in.');
	await screenshot('auth-invalid-login');

	await fill('#email', 'test@example.com');
	await fill('#password', 'password');
	await submitNearestForm('#email');
	await waitForUrl(/\/dashboard$/u);
	await waitForText('Recent Activity');
	await screenshot('auth-dashboard');
	assertBrowserClean('auth flow');
}

async function runCrmFlow(baseUrl) {
	const unique = `${Date.now()}-${process.pid}`;
	const firstName = 'E2E';
	const lastName = `Run ${unique}`;
	const updatedFirstName = 'E2E Updated';
	const email = `e2e-${unique}@example.test`;
	const phone = '+1 555 4242';
	const note = `E2E note ${unique}`;

	await openAppPath(baseUrl, '/contacts/create', 'Create Contact');
	await fill('#first_name', firstName);
	await fill('#last_name', lastName);
	await fill('#email', email);
	await fill('#phone', phone);
	await submitNearestForm('#first_name');
	await waitForText(`${firstName} ${lastName}`);
	await waitForText(email);
	await screenshot('crm-contact-created');

	await clickExactText('Favorite');
	await waitForText('Favorited');
	await screenshot('crm-contact-favorited');

	await fill('#note-body', note);
	await submitNearestForm('#note-body');
	await waitForText(note);
	await screenshot('crm-contact-note');

	await clickExactText('Edit');
	await waitForText('Edit Contact');
	await fill('#first_name', updatedFirstName);
	await submitNearestForm('#first_name');
	await waitForText(`${updatedFirstName} ${lastName}`);
	await screenshot('crm-contact-updated');

	page.once('dialog', (dialog) => dialog.accept());
	await clickExactText('Delete');
	await waitForUrl(/\/contacts$/u);
	await waitForText('Contact deleted');
	await screenshot('crm-contact-deleted');
	assertBrowserClean('crm flow');
}

async function runOrganizationFlow(baseUrl) {
	const name = `Acme Ventures E2E ${process.pid}`;

	await openAppPath(baseUrl, '/organizations', 'Organizations');
	await openAppPath(baseUrl, '/organizations/1', 'Organization name');
	await fill('#name', name);
	await submitNearestForm('#name');
	await waitForText(name);
	await screenshot('crm-organization-updated');
	assertBrowserClean('organization flow');
}

async function runFeatureInteractions(baseUrl) {
	await openAppPath(baseUrl, '/features/forms/validation', 'Validation');
	await submitNearestForm('#name');
	await waitForText('Name is required.');
	await fill('#name', 'Valid User');
	await fill('#email', 'valid@example.test');
	await fill('#age', '32');
	await submitNearestForm('#name');
	await waitForText('All fields passed validation.');
	await screenshot('feature-validation');

	await openAppPath(baseUrl, '/features/state/flash-data', 'Flash Data');
	const flashResponses = Promise.all([waitForResponse('/features/state/flash-data', 'POST'), waitForResponse('/features/state/flash-data', 'GET')]);
	await clickButtonText('Success Flash');
	await flashResponses;
	await waitForPath('/features/state/flash-data');
	await waitForText('This is a success flash message.');
	await waitForPageSettled();
	await screenshot('feature-flash');
	await resetPage();

	await openAppPath(baseUrl, '/features/state/remember', 'Remember');
	await clickButtonText('Reset');
	await clickButtonText('Increment');
	await clickButtonText('Increment');
	await fill('#name', 'E2E Remembered Name');
	await clickLinkText('Navigate Away');
	await waitForPath('/features/state/flash-data');
	await waitForText('Flash Data');
	await page.goBack({ waitUntil: 'domcontentloaded' });
	await waitForUrl(/\/features\/state\/remember$/u);
	await waitForText('Remember');
	const rememberedValue = await locator('#name').inputValue();
	if (!rememberedValue.includes('E2E Remembered Name')) {
		throw new Error(`remembered input was not restored. Output: ${rememberedValue}`);
	}
	await screenshot('feature-remember');

	await openAppPath(baseUrl, '/features/data-loading/partial-reloads', 'Partial Reloads');
	await clickButtonText('Reload Timestamp');
	await waitForText('Random Number');
	await screenshot('feature-partial-reload');

	await openAppPath(baseUrl, '/features/http/use-http', 'useHttp');
	await fill('#name', 'Alloy Browser');
	await fill('#email', 'alloy-browser@example.test');
	await submitNearestForm('#name');
	await waitForText('Hello, Alloy Browser!');
	await screenshot('feature-http');
	assertBrowserClean('feature interactions');
}

async function runRouteCrawl(baseUrl) {
	let index = 0;

	for (const route of routeMatrix()) {
		index += 1;
		try {
			await openAppPath(baseUrl, route.path, route.text, { minTextLength: route.minTextLength ?? 60 });
			await screenshot(`route-${String(index).padStart(2, '0')}-${slug(route.path)}`);
			assertBrowserClean(`route ${route.path}`);
		} catch (error) {
			const reason = error instanceof Error ? error.message : String(error);
			throw new Error(`route crawl failed for ${route.path}: ${reason}`);
		}
	}
}

async function openAppPath(baseUrl, routePath, text, options = {}) {
	try {
		await page.goto(`${baseUrl}${routePath}`, { waitUntil: 'domcontentloaded' });
		await waitForPath(routePath);
		await page.waitForFunction((minTextLength) => document.querySelector('#app') && document.body.innerText.trim().length > minTextLength, options.minTextLength ?? 60);

		if (text) {
			await waitForText(text);
		}
	} catch (error) {
		const reason = error instanceof Error ? error.message : String(error);
		throw new Error(`open ${routePath} failed at ${page?.url() ?? 'no page'}: ${reason}`);
	}
}

async function fill(selector, value) {
	await locator(selector).fill(value);
}

async function clickButtonText(text) {
	await page.getByRole('button', { name: text, exact: true }).click();
}

async function clickLinkText(text) {
	await page.getByRole('link', { name: text, exact: true }).click();
}

async function clickExactText(text) {
	await page.getByText(text, { exact: true }).first().click();
}

async function submitNearestForm(selector) {
	await locator(selector).evaluate((element) => {
		const form = element.closest('form');
		if (!(form instanceof HTMLFormElement)) {
			throw new Error('missing form');
		}

		form.requestSubmit();
	});
}

async function waitForSelector(selector) {
	await locator(selector).waitFor({ state: 'visible' });
}

async function waitForText(text) {
	try {
		await page.waitForFunction((expected) => document.body.innerText.includes(expected), text);
	} catch (error) {
		const reason = error instanceof Error ? error.message : String(error);
		throw new Error(`wait for text "${text}" failed at ${page?.url() ?? 'no page'}: ${reason}`);
	}
}

async function waitForUrl(pattern) {
	try {
		if (pattern instanceof RegExp) {
			await page.waitForFunction(({ source, flags }) => new RegExp(source, flags).test(window.location.href), {
				source: pattern.source,
				flags: pattern.flags,
			});

			return;
		}

		await page.waitForFunction((expected) => window.location.href === expected, pattern);
	} catch (error) {
		const reason = error instanceof Error ? error.message : String(error);
		throw new Error(`wait for url "${pattern}" failed at ${page?.url() ?? 'no page'}: ${reason}`);
	}
}

async function waitForPath(pathname) {
	try {
		await page.waitForFunction((expected) => window.location.pathname === expected, pathname);
	} catch (error) {
		const reason = error instanceof Error ? error.message : String(error);
		throw new Error(`wait for path "${pathname}" failed at ${page?.url() ?? 'no page'}: ${reason}`);
	}
}

async function waitForResponse(pathname, method) {
	await page.waitForResponse((response) => {
		const url = new URL(response.url());

		return url.pathname === pathname && response.request().method() === method && response.status() < 500;
	});
}

async function waitForPageSettled() {
	await page.waitForLoadState('networkidle', { timeout: 10_000 }).catch(() => {});
	await page.waitForTimeout(250);
}

function locator(selector) {
	return page.locator(selector).first();
}

async function screenshot(name) {
	const outputPath = path.join(screenshotsPath, `${name}.png`);
	await page.screenshot({ path: outputPath, fullPage: true });
	return outputPath;
}

async function failureArtifacts() {
	if (!page) {
		return;
	}

	writeFileSync(path.join(logsPath, 'failure-page.html'), await page.content());
	appendFileSync(path.join(logsPath, 'failure-page.log'), `url: ${page.url()}\n`);
	appendFileSync(path.join(logsPath, 'browser-issues.log'), `${pageIssues.join('\n')}\n`);

	try {
		await screenshot('failure');
	} catch {}

	try {
		await stopTrace('failure');
	} catch {}
}

async function startServer(app) {
	const stdout = createWriteStream(path.join(logsPath, `${target}-server.stdout.log`), { flags: 'a' });
	const stderr = createWriteStream(path.join(logsPath, `${target}-server.stderr.log`), { flags: 'a' });
	serverLogStreams = [stdout, stderr];
	serverExited = false;

	const child = spawn('go', ['run', './cmd'], {
		cwd: app.apiPath,
		env: app.env,
		stdio: ['ignore', 'pipe', 'pipe'],
	});

	child.stdout.pipe(stdout);
	child.stderr.pipe(stderr);

	let exited = false;
	let exitCode = null;

	child.on('exit', (code, signal) => {
		exited = true;
		serverExited = true;
		exitCode = signal ?? code;
		stdout.end();
		stderr.end();
	});

	const timeoutAt = Date.now() + 120_000;
	while (Date.now() < timeoutAt) {
		if (exited) {
			throw new Error(`server exited before becoming ready: ${exitCode}`);
		}

		if (await canLoad(`${app.baseUrl}/login`)) {
			await delay(250);
			return child;
		}

		await delay(500);
	}

	throw new Error(`server did not become ready at ${app.baseUrl}`);
}

async function launchBrowser() {
	browser = await chromium.launch({
		args: browserArgs(),
		executablePath: resolveBrowserExecutablePath(),
		headless: headed !== 'true',
	});

	context = await browser.newContext({
		ignoreHTTPSErrors: true,
		viewport: { width: 1440, height: 1000 },
	});

	await context.tracing.start({
		screenshots: true,
		snapshots: true,
		sources: true,
	});

	await createPage();
}

async function createPage() {
	page = await context.newPage();
	page.setDefaultTimeout(30_000);
	page.setDefaultNavigationTimeout(30_000);

	page.on('pageerror', (error) => {
		pageIssues.push(`pageerror: ${error.message}`);
	});

	page.on('console', (message) => {
		if (message.type() === 'error') {
			pageIssues.push(`console error: ${message.text()}`);
		}
	});
}

async function resetPage() {
	await page?.close().catch(() => {});
	clearBrowserIssues();
	await createPage();
}

async function closeBrowser() {
	await context?.close().catch(() => {});
	await browser?.close().catch(() => {});
	context = undefined;
	browser = undefined;
	page = undefined;
}

async function stopTrace(reason) {
	if (!context) {
		return;
	}

	await context.tracing.stop({ path: path.join(tracesPath, `${reason}.zip`) });
}

async function stopServer() {
	if (!serverProcess) {
		return;
	}

	if (!serverExited && !serverProcess.killed) {
		serverProcess.kill('SIGTERM');
		await Promise.race([
			new Promise((resolve) => serverProcess.once('exit', resolve)),
			delay(5_000).then(() => {
				if (!serverExited && !serverProcess.killed) {
					serverProcess.kill('SIGKILL');
				}
			}),
		]);
	}

	for (const stream of serverLogStreams) {
		if (!stream.destroyed) {
			stream.end();
		}
	}

	serverProcess = undefined;
	serverLogStreams = [];
}

function assertBrowserClean(contextName) {
	const issues = pageIssues.filter((issue) => /error|exception|failed|rejected|typeerror|referenceerror/iu.test(issue));
	if (issues.length) {
		throw new Error(`Browser page issues during ${contextName}:\n${issues.join('\n')}`);
	}

	clearBrowserIssues();
}

function clearBrowserIssues() {
	pageIssues = [];
}

async function targetConfig(port) {
	if (target === 'bedrock') {
		const sourceRoot = path.resolve(process.env.BEDROCK_INERTIA_SOURCE ?? '/Users/gocanto/Sites/bedrock/services/demo/inertia');
		const appPath = path.join(sourceRoot, 'app');
		const apiPath = path.join(sourceRoot, 'api');

		if (buildApp) {
			runCommand('pnpm', ['--dir', appPath, 'build'], { cwd: sourceRoot });
		}

		return {
			apiPath,
			baseUrl: `http://127.0.0.1:${port}`,
			env: {
				...process.env,
				PORT: String(port),
				BEDROCK_DEMO_DB: path.join(databasePath, 'bedrock-beacon.db'),
				APP_VERSION: `e2e-${runId}`,
			},
		};
	}

	const distPath = path.join(repoRoot, 'storage/apps/inertia/dist/app');
	if (!existsSync(distPath) || !statSync(distPath).isDirectory()) {
		throw new Error(`Missing Alloy Inertia dist at ${distPath}. Run pnpm --filter @alloy/inertia-app build first.`);
	}

	return {
		apiPath: path.join(repoRoot, 'apps/inertia/api'),
		baseUrl: `http://127.0.0.1:${port}`,
		env: {
			...process.env,
			PORT: String(port),
			GOWORK: 'off',
			ALLOY_INERTIA_DIST_PATH: distPath,
			ALLOY_INERTIA_DB_PATH: path.join(databasePath, 'alloy-beacon.db'),
			APP_VERSION: `e2e-${runId}`,
		},
	};
}

function runCommand(command, args, options = {}) {
	const result = spawnSync(command, args, {
		cwd: options.cwd ?? repoRoot,
		encoding: 'utf8',
		env: process.env,
		stdio: 'pipe',
		maxBuffer: 20 * 1024 * 1024,
	});

	if (result.status !== 0) {
		throw new Error(`${command} ${args.join(' ')} failed:\n${result.stdout}\n${result.stderr}`);
	}
}

async function canLoad(url) {
	return new Promise((resolve) => {
		const request = http.get(url, { timeout: 2_000 }, (response) => {
			response.resume();
			resolve(response.statusCode !== undefined && response.statusCode >= 200 && response.statusCode < 500);
		});

		request.on('timeout', () => {
			request.destroy();
			resolve(false);
		});

		request.on('error', () => resolve(false));
	});
}

async function freePort() {
	return new Promise((resolve, reject) => {
		const server = net.createServer();
		server.listen(0, '127.0.0.1', () => {
			const address = server.address();
			server.close(() => {
				if (address && typeof address === 'object') {
					resolve(address.port);
				} else {
					reject(new Error('failed to reserve a free port'));
				}
			});
		});
		server.on('error', reject);
	});
}

async function step(label, fn) {
	console.log(`==> ${label}`);
	try {
		await fn();
	} catch (error) {
		try {
			await screenshot(`failure-${slug(label)}`);
		} catch {}

		throw error;
	}
}

function routeMatrix() {
	return [
		{ path: '/dashboard', text: 'Recent Activity' },
		{ path: '/contacts', text: 'Contacts' },
		{ path: '/contacts/create', text: 'Create Contact' },
		{ path: '/organizations', text: 'Organizations' },
		{ path: '/features/forms/use-form', text: 'useForm' },
		{ path: '/features/forms/form-component', text: 'Form Component' },
		{ path: '/features/forms/file-uploads', text: 'File Uploads' },
		{ path: '/features/forms/validation', text: 'Validation' },
		{ path: '/features/forms/httppreview', text: 'HTTPPreview' },
		{ path: '/features/forms/optimistic-updates', text: 'Optimistic Updates' },
		{ path: '/features/forms/use-form-context', text: 'useFormContext' },
		{ path: '/features/forms/dotted-keys', text: 'Dotted Keys' },
		{ path: '/features/forms/routegen', text: 'RouteGen' },
		{ path: '/features/navigation/links', text: 'Links' },
		{ path: '/features/navigation/preserve-state', text: 'Preserve State' },
		{ path: '/features/navigation/preserve-scroll', text: 'Preserve Scroll' },
		{ path: '/features/navigation/view-transitions', text: 'View Transitions' },
		{ path: '/features/navigation/history-management', text: 'History Management' },
		{ path: '/features/navigation/async-requests', text: 'Async Requests' },
		{ path: '/features/navigation/async-slow', text: 'Async Requests' },
		{ path: '/features/navigation/manual-visits', text: 'Manual Visits' },
		{ path: '/features/navigation/redirects', text: 'Redirects' },
		{ path: '/features/navigation/scroll-management', text: 'Scroll Management' },
		{ path: '/features/navigation/instant-visits', text: 'Instant Visits' },
		{ path: '/features/navigation/instant-visit-target', text: 'Instant Visit Target' },
		{ path: '/features/navigation/url-fragments', text: 'URL Fragments' },
		{ path: '/features/data-loading/deferred-props', text: 'Deferred Props' },
		{ path: '/features/data-loading/partial-reloads', text: 'Partial Reloads' },
		{ path: '/features/data-loading/infinite-scroll', text: 'Infinite Scroll' },
		{ path: '/features/data-loading/when-visible', text: 'When Visible' },
		{ path: '/features/data-loading/polling', text: 'Polling' },
		{ path: '/features/data-loading/prop-merging', text: 'Prop Merging' },
		{ path: '/features/data-loading/optional-props', text: 'Optional Props' },
		{ path: '/features/data-loading/once-props/1', text: 'Once Props' },
		{ path: '/features/prefetching/link-prefetch', text: 'Link Prefetch' },
		{ path: '/features/prefetching/stale-while-revalidate', text: 'Stale While Revalidate' },
		{ path: '/features/prefetching/manual-prefetch', text: 'Manual Prefetch' },
		{ path: '/features/prefetching/cache-management', text: 'Cache Management' },
		{ path: '/features/state/remember', text: 'Remember' },
		{ path: '/features/state/flash-data', text: 'Flash Data' },
		{ path: '/features/state/shared-props', text: 'Shared Props' },
		{ path: '/features/layouts/persistent-layouts', text: 'Persistent Layouts' },
		{ path: '/features/layouts/persistent-layouts/page-2', text: 'Page 2' },
		{ path: '/features/layouts/nested-layouts', text: 'Nested Layouts' },
		{ path: '/features/layouts/head', text: 'Head' },
		{ path: '/features/layouts/layout-props', text: 'Layout Props' },
		{ path: '/features/events/global-events', text: 'Global Events' },
		{ path: '/features/events/visit-callbacks', text: 'Visit Callbacks' },
		{ path: '/features/events/progress', text: 'Progress' },
		{ path: '/features/events/progress/slow', text: 'Progress' },
		{ path: '/features/http/use-http', text: 'useHttp' },
		{ path: '/features/errors/http-error', text: 'HTTP Exceptions' },
		{ path: '/features/errors/network-errors', text: 'Network Errors' },
	];
}

function browserArgs() {
	const raw = process.env.PLAYWRIGHT_BROWSER_ARGS ?? process.env.AGENT_BROWSER_ARGS ?? '';
	return raw
		.split(',')
		.map((value) => value.trim())
		.filter(Boolean);
}

function resolveBrowserExecutablePath() {
	const candidate = option('browser-executable') ?? process.env.PLAYWRIGHT_EXECUTABLE_PATH ?? process.env.CHROME_EXECUTABLE_PATH ?? process.env.GOOGLE_CHROME_BIN ?? process.env.CHROMIUM_BIN;

	if (!candidate) {
		return undefined;
	}

	try {
		accessSync(candidate, constants.X_OK);
		return candidate;
	} catch {
		throw new Error(`Browser executable is not executable: ${candidate}`);
	}
}

function option(name) {
	const index = process.argv.indexOf(`--${name}`);
	if (index === -1) {
		return null;
	}

	return process.argv[index + 1] ?? null;
}

function hasFlag(name) {
	return process.argv.includes(`--${name}`);
}

function slug(value) {
	return value
		.toLowerCase()
		.replaceAll(/[^a-z0-9]+/g, '-')
		.replaceAll(/^-|-$/g, '')
		.slice(0, 120);
}

function isCI() {
	return process.env.CI === 'true';
}

function fail(message) {
	console.error(message);
	process.exit(1);
}

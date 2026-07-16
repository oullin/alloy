// Shared, driver-agnostic logic for the Inertia demo e2e runners.
//
// Both runners (agent-browser via Helium, and Playwright/Chromium) drive the
// exact same flows and route matrix. This module owns everything that does not
// depend on the browser driver: CLI parsing, artifact paths, server lifecycle,
// the target configuration, and the flow orchestration. Each runner passes a
// thin "driver" object that implements the browser primitives the flows call.
//
// Adding or changing a route or flow is a one-file edit here.

import { spawn, spawnSync } from 'node:child_process';
import { createWriteStream, existsSync, mkdirSync, statSync } from 'node:fs';
import http from 'node:http';
import net from 'node:net';
import os from 'node:os';
import path from 'node:path';
import process from 'node:process';
import { setTimeout as delay } from 'node:timers/promises';

// Portable default: artifacts land in an OS temp directory so a clean CI runner
// or any contributor can run the suite without machine-specific setup. Override
// with CODEX_BROWSER_ARTIFACTS_DIR (CI points this at the runner temp dir).
export const defaultArtifactsPath = path.join(os.tmpdir(), 'alloy-inertia-e2e', 'browser-artifacts');

export function option(name) {
	const index = process.argv.indexOf(`--${name}`);
	if (index === -1) {
		return null;
	}

	return process.argv[index + 1] ?? null;
}

export function hasFlag(name) {
	return process.argv.includes(`--${name}`);
}

export function fail(message) {
	console.error(message);
	process.exit(1);
}

export function isCI() {
	return process.env.CI === 'true';
}

export function slug(value) {
	return value
		.toLowerCase()
		.replaceAll(/[^a-z0-9]+/g, '-')
		.replaceAll(/^-|-$/g, '')
		.slice(0, 120);
}

export function parseTarget() {
	const target = option('target') ?? 'alloy';
	if (!['alloy', 'bedrock'].includes(target)) {
		fail(`Unknown target "${target}". Expected "alloy" or "bedrock".`);
	}

	return target;
}

// Builds the run id and artifact directories. `extraDirs` lets a driver request
// additional subdirectories (agent-browser needs a downloads dir).
export function createArtifacts(extraDirs = []) {
	const runId = `${new Date().toISOString().replaceAll(/[:.]/g, '-')}-${process.pid}`;
	const artifactsRoot = path.resolve(process.env.CODEX_BROWSER_ARTIFACTS_DIR ?? defaultArtifactsPath, 'alloy-inertia-demo', runId);
	const screenshotsPath = path.join(artifactsRoot, 'screenshots');
	const tracesPath = path.join(artifactsRoot, 'traces');
	const logsPath = path.join(artifactsRoot, 'logs');
	const databasePath = path.join(artifactsRoot, 'database');
	const named = { runId, artifactsRoot, screenshotsPath, tracesPath, logsPath, databasePath };

	const extras = {};
	for (const name of extraDirs) {
		extras[name] = path.join(artifactsRoot, name);
	}

	for (const directory of [artifactsRoot, screenshotsPath, tracesPath, logsPath, databasePath, ...Object.values(extras)]) {
		mkdirSync(directory, { recursive: true });
	}

	return { ...named, ...extras };
}

export async function freePort() {
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

export async function canLoad(url) {
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

function runCommand(command, args, options = {}) {
	const result = spawnSync(command, args, {
		cwd: options.cwd ?? process.cwd(),
		encoding: 'utf8',
		env: process.env,
		stdio: 'pipe',
		maxBuffer: 20 * 1024 * 1024,
	});

	// spawnSync reports a failure to launch (command missing from PATH) via
	// result.error with a null status; surface that instead of a generic
	// message with undefined output.
	if (result.error) {
		throw result.error;
	}

	if (result.status !== 0) {
		throw new Error(`${command} ${args.join(' ')} failed:\n${result.stdout}\n${result.stderr}`);
	}
}

// Resolves the API path, base URL and environment for the chosen target.
export function resolveTargetConfig({ target, port, repoRoot, databasePath, runId, buildApp }) {
	if (target === 'bedrock') {
		const source = process.env.BEDROCK_INERTIA_SOURCE;
		if (!source) {
			throw new Error('Set BEDROCK_INERTIA_SOURCE to your local Bedrock inertia demo checkout (e.g. <bedrock>/services/demo/inertia) to run the bedrock parity target.');
		}

		const sourceRoot = path.resolve(source);
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

	const distPath = path.join(repoRoot, 'web/storage/inertia-demo/dist/app');
	if (!existsSync(distPath) || !statSync(distPath).isDirectory()) {
		throw new Error(`Missing Alloy Inertia dist at ${distPath}. Run pnpm inertia-demo:build first.`);
	}

	return {
		apiPath: path.join(repoRoot, 'web/inertia-demo/api'),
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

// Handle over the spawned Go server. Encapsulates the log streams and exit
// tracking so runners do not manage module-level server state themselves.
export class ServerHandle {
	constructor(child, streams) {
		this.child = child;
		this.streams = streams;
		this.exited = false;

		child.on('exit', () => {
			this.exited = true;
			for (const stream of streams) {
				stream.end();
			}
		});
	}

	async stop() {
		if (!this.exited && !this.child.killed) {
			this.child.kill('SIGTERM');
			await Promise.race([
				new Promise((resolve) => this.child.once('exit', resolve)),
				delay(5_000).then(() => {
					if (!this.exited && !this.child.killed) {
						this.child.kill('SIGKILL');
					}
				}),
			]);
		}

		for (const stream of this.streams) {
			if (!stream.destroyed) {
				stream.end();
			}
		}
	}
}

// Starts the Go API server and waits for it to answer on /login. `onReady` runs
// once the server is up (agent-browser uses it to start its trace).
export async function startServer(app, { target, logsPath, onReady } = {}) {
	const stdout = createWriteStream(path.join(logsPath, `${target}-server.stdout.log`), { flags: 'a' });
	const stderr = createWriteStream(path.join(logsPath, `${target}-server.stderr.log`), { flags: 'a' });

	const child = spawn('go', ['run', './cmd'], {
		cwd: app.apiPath,
		env: app.env,
		stdio: ['ignore', 'pipe', 'pipe'],
	});

	child.stdout.pipe(stdout);
	child.stderr.pipe(stderr);

	const handle = new ServerHandle(child, [stdout, stderr]);
	let exitCode = null;
	let spawnError = null;
	child.on('exit', (code, signal) => {
		exitCode = signal ?? code;
	});
	// Without a listener a failed spawn (go missing from PATH) raises an
	// uncaught 'error' event instead of a readable failure.
	child.on('error', (error) => {
		spawnError = error;
	});

	const timeoutAt = Date.now() + 120_000;
	while (Date.now() < timeoutAt) {
		if (spawnError) {
			throw new Error(`server failed to start: ${spawnError.message}`);
		}

		if (handle.exited) {
			throw new Error(`server exited before becoming ready: ${exitCode}`);
		}

		if (await canLoad(`${app.baseUrl}/login`)) {
			await delay(250);
			if (onReady) {
				await onReady();
			}

			return handle;
		}

		await delay(500);
	}

	throw new Error(`server did not become ready at ${app.baseUrl}`);
}

export async function step(driver, label, fn) {
	console.log(`==> ${label}`);
	try {
		await fn();
	} catch (error) {
		try {
			await driver.screenshot(`failure-${slug(label)}`);
		} catch {}

		throw error;
	}
}

// ---------------------------------------------------------------------------
// Flows. Each takes a driver implementing the browser primitives and runs the
// same sequence of steps regardless of which browser backend is in use.
// ---------------------------------------------------------------------------

export async function runAllFlows(driver, baseUrl) {
	await step(driver, 'auth flow', () => runAuthFlow(driver, baseUrl));
	await step(driver, 'crm flow', () => runCrmFlow(driver, baseUrl));
	await step(driver, 'organization flow', () => runOrganizationFlow(driver, baseUrl));
	await step(driver, 'feature interactions', () => runFeatureInteractions(driver, baseUrl));
	await step(driver, 'route crawl', () => runRouteCrawl(driver, baseUrl));
}

export async function runAuthFlow(driver, baseUrl) {
	driver.clearIssues();
	await driver.open(`${baseUrl}/dashboard`);
	await driver.waitForSelector('#email');
	await driver.screenshot('auth-login');

	await driver.fill('#email', 'wrong@example.test');
	await driver.fill('#password', 'incorrect');
	await driver.submitForm('#email');
	await driver.waitForText('Use test@example.com and password to sign in.');
	await driver.screenshot('auth-invalid-login');

	await driver.fill('#email', 'test@example.com');
	await driver.fill('#password', 'password');
	await driver.submitForm('#email');
	await driver.waitForPath('/dashboard');
	await driver.waitForText('Recent Activity');
	await driver.screenshot('auth-dashboard');
	driver.assertClean('auth flow');
}

export async function runCrmFlow(driver, baseUrl) {
	const unique = `${Date.now()}-${process.pid}`;
	const firstName = 'E2E';
	const lastName = `Run ${unique}`;
	const updatedFirstName = 'E2E Updated';
	const email = `e2e-${unique}@example.test`;
	const phone = '+1 555 4242';
	const note = `E2E note ${unique}`;

	await driver.openPath(baseUrl, '/contacts/create', 'Create Contact');
	await driver.fill('#first_name', firstName);
	await driver.fill('#last_name', lastName);
	await driver.fill('#email', email);
	await driver.fill('#phone', phone);
	await driver.submitForm('#first_name');
	await driver.waitForText(`${firstName} ${lastName}`);
	await driver.waitForText(email);
	await driver.screenshot('crm-contact-created');

	await driver.clickText('Favorite');
	await driver.waitForText('Favorited');
	await driver.screenshot('crm-contact-favorited');

	await driver.fill('#note-body', note);
	await driver.submitForm('#note-body');
	await driver.waitForText(note);
	await driver.screenshot('crm-contact-note');

	await driver.clickText('Edit');
	await driver.waitForText('Edit Contact');
	await driver.fill('#first_name', updatedFirstName);
	await driver.submitForm('#first_name');
	await driver.waitForText(`${updatedFirstName} ${lastName}`);
	await driver.screenshot('crm-contact-updated');

	await driver.acceptNextDialog();
	await driver.clickText('Delete');
	await driver.waitForPath('/contacts');
	await driver.waitForText('Contact deleted');
	await driver.screenshot('crm-contact-deleted');
	driver.assertClean('crm flow');
}

export async function runOrganizationFlow(driver, baseUrl) {
	const name = `Acme Ventures E2E ${process.pid}`;

	await driver.openPath(baseUrl, '/organizations', 'Organizations');
	await driver.openPath(baseUrl, '/organizations/1', 'Organization name');
	await driver.fill('#name', name);
	await driver.submitForm('#name');
	await driver.waitForText(name);
	await driver.screenshot('crm-organization-updated');
	driver.assertClean('organization flow');
}

export async function runFeatureInteractions(driver, baseUrl) {
	await driver.openPath(baseUrl, '/features/forms/validation', 'Validation');
	await driver.submitForm('#name');
	await driver.waitForText('Name is required.');
	await driver.fill('#name', 'Valid User');
	await driver.fill('#email', 'valid@example.test');
	await driver.fill('#age', '32');
	await driver.submitForm('#name');
	await driver.waitForText('All fields passed validation.');
	await driver.screenshot('feature-validation');

	await driver.openPath(baseUrl, '/features/state/flash-data', 'Flash Data');
	const flash = driver.beforeFlash('/features/state/flash-data');
	await driver.clickButtonText('Success Flash');
	await driver.afterFlash(flash);
	await driver.waitForText('This is a success flash message.');
	await driver.settle();
	await driver.screenshot('feature-flash');
	await driver.resetPage();

	await driver.openPath(baseUrl, '/features/state/remember', 'Remember');
	await driver.clickButtonText('Reset');
	await driver.clickButtonText('Increment');
	await driver.clickButtonText('Increment');
	await driver.fill('#name', 'E2E Remembered Name');
	await driver.clickLinkText('Navigate Away');
	await driver.waitForPath('/features/state/flash-data');
	await driver.waitForText('Flash Data');
	await driver.goBack();
	await driver.waitForPath('/features/state/remember');
	await driver.waitForText('Remember');
	const rememberedValue = await driver.getInputValue('#name');
	if (!rememberedValue.includes('E2E Remembered Name')) {
		throw new Error(`remembered input was not restored. Output: ${rememberedValue}`);
	}
	await driver.screenshot('feature-remember');

	await driver.openPath(baseUrl, '/features/data-loading/partial-reloads', 'Partial Reloads');
	await driver.clickButtonText('Reload Timestamp');
	await driver.waitForText('Random Number');
	await driver.screenshot('feature-partial-reload');

	await driver.openPath(baseUrl, '/features/http/use-http', 'useHttp');
	await driver.fill('#name', 'Alloy Browser');
	await driver.fill('#email', 'alloy-browser@example.test');
	await driver.submitForm('#name');
	await driver.waitForText('Hello, Alloy Browser!');
	await driver.screenshot('feature-http');
	driver.assertClean('feature interactions');
}

export async function runRouteCrawl(driver, baseUrl) {
	let index = 0;

	for (const route of routeMatrix()) {
		index += 1;
		try {
			await driver.openPath(baseUrl, route.path, route.text, { minTextLength: route.minTextLength ?? 60 });
			await driver.screenshot(`route-${String(index).padStart(2, '0')}-${slug(route.path)}`);
			driver.assertClean(`route ${route.path}`);
		} catch (error) {
			const reason = error instanceof Error ? error.message : String(error);
			throw new Error(`route crawl failed for ${route.path}: ${reason}`);
		}
	}
}

export function routeMatrix() {
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

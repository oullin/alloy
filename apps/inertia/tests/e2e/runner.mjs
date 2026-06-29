#!/usr/bin/env node
import { spawn, spawnSync } from 'node:child_process';
import { accessSync, appendFileSync, constants, createWriteStream, existsSync, mkdirSync, statSync } from 'node:fs';
import http from 'node:http';
import net from 'node:net';
import path from 'node:path';
import process from 'node:process';
import { setTimeout as delay } from 'node:timers/promises';
import { fileURLToPath } from 'node:url';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, '../../../..');
const defaultArtifactsPath = '/Users/gocanto/.cache/codex/browser-artifacts';
const target = option('target') ?? 'alloy';
const buildApp = hasFlag('build-app');
const headed = option('headed') ?? process.env.AGENT_BROWSER_HEADED ?? 'true';

if (!['alloy', 'bedrock'].includes(target)) {
	fail(`Unknown target "${target}". Expected "alloy" or "bedrock".`);
}

const runId = `${new Date().toISOString().replaceAll(/[:.]/g, '-')}-${process.pid}`;
const artifactsRoot = path.resolve(process.env.CODEX_BROWSER_ARTIFACTS_DIR ?? defaultArtifactsPath, 'alloy-inertia', runId);
const screenshotsPath = path.join(artifactsRoot, 'screenshots');
const tracesPath = path.join(artifactsRoot, 'traces');
const downloadsPath = path.join(artifactsRoot, 'downloads');
const logsPath = path.join(artifactsRoot, 'logs');
const databasePath = path.join(artifactsRoot, 'database');

for (const directory of [artifactsRoot, screenshotsPath, tracesPath, downloadsPath, logsPath, databasePath]) {
	mkdirSync(directory, { recursive: true });
}

const commandLogPath = path.join(logsPath, 'agent-browser.log');
const session = `alloy-inertia-e2e-${process.pid}`;
const browserBin = process.env.AGENT_BROWSER_BIN ?? 'agent-browser';
const heliumPath = resolveHeliumPath();

const browserEnv = {
	...process.env,
	AGENT_BROWSER_SESSION: session,
	AGENT_BROWSER_EXECUTABLE_PATH: heliumPath,
	AGENT_BROWSER_SCREENSHOT_DIR: screenshotsPath,
	AGENT_BROWSER_DOWNLOAD_PATH: downloadsPath,
	AGENT_BROWSER_DEFAULT_TIMEOUT: '30000',
	AGENT_BROWSER_ENGINE: 'chrome',
	AGENT_BROWSER_HEADED: headed,
};

let serverProcess;
let serverExited = false;
let serverLogStreams = [];

try {
	console.log(`Artifacts: ${artifactsRoot}`);
	console.log(`Agent Browser session: ${session}`);
	console.log(`Helium: ${heliumPath}`);

	await verifyAgentBrowser();

	const port = await freePort();
	const app = await targetConfig(port);

	serverProcess = await startServer(app);

	await step('auth flow', () => runAuthFlow(app.baseUrl));
	await step('crm flow', () => runCrmFlow(app.baseUrl));
	await step('organization flow', () => runOrganizationFlow(app.baseUrl));
	await step('feature interactions', () => runFeatureInteractions(app.baseUrl));
	await step('route crawl', () => runRouteCrawl(app.baseUrl));

	await stopTrace('complete');
	await closeBrowser();
	await stopServer();

	console.log(`Inertia E2E passed for ${target}.`);
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
	clearBrowserLogs();
	browser('open', [`${baseUrl}/dashboard`]);
	browser('wait', ['#email']);
	await screenshot('auth-login');

	browser('fill', ['#email', 'wrong@example.test']);
	browser('fill', ['#password', 'incorrect']);
	submitForm();
	browser('wait', ['--text', 'Use test@example.com and password to sign in.']);
	await screenshot('auth-invalid-login');

	browser('fill', ['#email', 'test@example.com']);
	browser('fill', ['#password', 'password']);
	browser('click', ['button[type="submit"]']);
	browser('wait', ['--url', '**/dashboard']);
	browser('wait', ['--text', 'Recent Activity']);
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
	browser('fill', ['#first_name', firstName]);
	browser('fill', ['#last_name', lastName]);
	browser('fill', ['#email', email]);
	browser('fill', ['#phone', phone]);
	submitForm();
	browser('wait', ['--text', `${firstName} ${lastName}`]);
	browser('wait', ['--text', email]);
	await screenshot('crm-contact-created');

	browser('find', ['text', 'Favorite', 'click']);
	browser('wait', ['--text', 'Favorited']);
	await screenshot('crm-contact-favorited');

	browser('fill', ['#note-body', note]);
	submitForm();
	browser('wait', ['--text', note]);
	await screenshot('crm-contact-note');

	browser('find', ['text', 'Edit', 'click', '--exact']);
	browser('wait', ['--text', 'Edit Contact']);
	browser('fill', ['#first_name', updatedFirstName]);
	submitForm();
	browser('wait', ['--text', `${updatedFirstName} ${lastName}`]);
	await screenshot('crm-contact-updated');

	browser('eval', ['window.confirm = () => true']);
	browser('find', ['text', 'Delete', 'click', '--exact']);
	browser('wait', ['--url', '**/contacts']);
	browser('wait', ['--text', 'Contact deleted']);
	await screenshot('crm-contact-deleted');
	assertBrowserClean('crm flow');
}

async function runOrganizationFlow(baseUrl) {
	const name = `Acme Ventures E2E ${process.pid}`;

	await openAppPath(baseUrl, '/organizations', 'Organizations');
	await openAppPath(baseUrl, '/organizations/1', 'Organization name');
	browser('wait', ['--text', 'Organization name']);
	browser('fill', ['#name', name]);
	submitForm();
	browser('wait', ['--text', name]);
	await screenshot('crm-organization-updated');
	assertBrowserClean('organization flow');
}

async function runFeatureInteractions(baseUrl) {
	await openAppPath(baseUrl, '/features/forms/validation', 'Validation');
	submitForm();
	browser('wait', ['--text', 'Name is required.']);
	browser('fill', ['#name', 'Valid User']);
	browser('fill', ['#email', 'valid@example.test']);
	browser('fill', ['#age', '32']);
	submitForm();
	browser('wait', ['--text', 'All fields passed validation.']);
	await screenshot('feature-validation');

	await openAppPath(baseUrl, '/features/state/flash-data', 'Flash Data');
	clickButtonText('Success Flash');
	browser('wait', ['--text', 'This is a success flash message.']);
	await screenshot('feature-flash');

	await openAppPath(baseUrl, '/features/state/remember', 'Remember');
	clickButtonText('Reset');
	clickButtonText('Increment');
	clickButtonText('Increment');
	browser('fill', ['#name', 'E2E Remembered Name']);
	clickLinkText('Navigate Away');
	browser('wait', ['--text', 'Flash Data']);
	browser('back');
	browser('wait', ['--url', '**/features/state/remember']);
	browser('wait', ['--text', 'Remember']);
	const rememberedValue = browser('get', ['value', '#name']);
	if (!rememberedValue.includes('E2E Remembered Name')) {
		throw new Error(`remembered input was not restored. Output: ${rememberedValue}`);
	}
	await screenshot('feature-remember');

	await openAppPath(baseUrl, '/features/data-loading/partial-reloads', 'Partial Reloads');
	clickButtonText('Reload Timestamp');
	browser('wait', ['--text', 'Random Number']);
	await screenshot('feature-partial-reload');

	await openAppPath(baseUrl, '/features/http/use-http', 'useHttp');
	browser('fill', ['#name', 'Alloy Browser']);
	browser('fill', ['#email', 'alloy-browser@example.test']);
	submitForm();
	browser('wait', ['--text', 'Hello, Alloy Browser!']);
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
	browser('open', [`${baseUrl}${routePath}`]);
	browser('wait', ['--fn', `document.querySelector("#app") && document.body.innerText.trim().length > ${options.minTextLength ?? 60}`]);

	if (text) {
		browser('wait', ['--text', text]);
	}
}

async function screenshot(name) {
	const outputPath = path.join(screenshotsPath, `${name}.png`);
	browser('screenshot', ['--full', outputPath]);
	return outputPath;
}

function submitForm(selector = 'form') {
	const script = `(() => {
		const form = document.querySelector(${JSON.stringify(selector)});
		if (!(form instanceof HTMLFormElement)) {
			throw new Error('missing form: ${selector}');
		}
		form.requestSubmit();
		return true;
	})()`;

	browser('eval', [script]);
}

function clickButtonText(text) {
	clickElementText('button', text);
}

function clickLinkText(text) {
	clickElementText('a', text);
}

function clickElementText(selector, text) {
	const script = `(() => {
		const normalize = (value) => (value ?? '').replace(/\\s+/g, ' ').trim();
		const expected = ${JSON.stringify(text)};
		const element = [...document.querySelectorAll(${JSON.stringify(selector)})].find((candidate) => normalize(candidate.textContent) === expected);
		if (!(element instanceof HTMLElement)) {
			throw new Error('missing ${selector} with text: ${text}');
		}
		element.scrollIntoView({ block: 'center', inline: 'center' });
		element.click();
		return normalize(element.textContent);
	})()`;

	browser('eval', [script]);
}

async function failureArtifacts() {
	const network = browser('network', ['requests'], { optional: true });
	if (network) {
		appendFileSync(path.join(logsPath, 'network-requests.log'), `${network}\n`);
	}

	const url = browser('get', ['url'], { optional: true });
	if (url) {
		appendFileSync(path.join(logsPath, 'failure-page.log'), `url: ${url}\n`);
	}

	const body = browser('get', ['text', 'body'], { optional: true });
	if (body) {
		appendFileSync(path.join(logsPath, 'failure-page.log'), `body:\n${body}\n`);
	}

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
			browser('trace', ['start']);
			return child;
		}

		await delay(500);
	}

	throw new Error(`server did not become ready at ${app.baseUrl}`);
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

async function closeBrowser() {
	browser('close', [], { optional: true });
}

async function stopTrace(reason) {
	const outputPath = path.join(tracesPath, `${reason}.json`);
	browser('trace', ['stop', outputPath], { optional: true });
}

function assertBrowserClean(context) {
	const errors = browser('errors', ['--clear'], { optional: true });
	if (errors && !/no page errors/i.test(errors) && /error|exception|failed|rejected|typeerror|referenceerror/i.test(errors)) {
		throw new Error(`Browser page errors during ${context}:\n${errors}`);
	}

	browser('console', ['--clear'], { optional: true });
}

function clearBrowserLogs() {
	browser('errors', ['--clear'], { optional: true });
	browser('console', ['--clear'], { optional: true });
	browser('network', ['requests', '--clear'], { optional: true });
}

function browser(command, args = [], options = {}) {
	const commandArgs = [
		'--session',
		session,
		'--executable-path',
		heliumPath,
		'--screenshot-dir',
		screenshotsPath,
		'--download-path',
		downloadsPath,
		'--engine',
		'chrome',
		'--headed',
		headed,
		command,
		...args,
	];

	appendFileSync(commandLogPath, `$ ${browserBin} ${commandArgs.map(quoteShell).join(' ')}\n`);

	const result = spawnSync(browserBin, commandArgs, {
		encoding: 'utf8',
		env: browserEnv,
		input: options.input,
		maxBuffer: 20 * 1024 * 1024,
	});

	appendFileSync(commandLogPath, result.stdout ? `${result.stdout}\n` : '');
	appendFileSync(commandLogPath, result.stderr ? `${result.stderr}\n` : '');

	if (result.status !== 0 && !options.optional) {
		throw new Error(`agent-browser ${command} failed (${result.status}):\n${result.stdout}\n${result.stderr}`);
	}

	return `${result.stdout ?? ''}${result.stderr ?? ''}`.trim();
}

async function verifyAgentBrowser() {
	const result = spawnSync(browserBin, ['--version'], {
		encoding: 'utf8',
		env: browserEnv,
	});

	if (result.status !== 0) {
		throw new Error(`agent-browser is not available via "${browserBin}".\n${result.stdout}\n${result.stderr}`);
	}
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

function resolveHeliumPath() {
	const candidates = [
		process.env.HELIUM_EXECUTABLE_PATH,
		process.env.AGENT_BROWSER_EXECUTABLE_PATH,
		'/Applications/Helium.app/Contents/MacOS/Helium',
		'/Applications/Helium Browser.app/Contents/MacOS/Helium',
		path.join(process.env.HOME ?? '', 'Applications/Helium.app/Contents/MacOS/Helium'),
		path.join(process.env.HOME ?? '', 'Applications/Helium Browser.app/Contents/MacOS/Helium'),
	].filter(Boolean);

	for (const candidate of candidates) {
		try {
			accessSync(candidate, constants.X_OK);
			return candidate;
		} catch {}
	}

	throw new Error(`Helium browser executable not found. Set HELIUM_EXECUTABLE_PATH or install Helium at /Applications/Helium.app/Contents/MacOS/Helium.`);
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

function quoteShell(value) {
	if (/^[a-zA-Z0-9_./:=@-]+$/u.test(value)) {
		return value;
	}

	return `'${value.replaceAll("'", "'\\''")}'`;
}

function fail(message) {
	console.error(message);
	process.exit(1);
}

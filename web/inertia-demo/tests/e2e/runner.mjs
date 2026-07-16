#!/usr/bin/env node
// Legacy/opt-in e2e runner: drives the Inertia demo through Agent Browser
// (Helium). The portable, canonical runner is ./runner.playwright.mjs.
//
// The flows, route matrix and server lifecycle live in ./shared.mjs. This file
// is a thin driver adapter that implements the browser primitives the shared
// flows call by shelling out to the agent-browser CLI.

import { spawnSync } from 'node:child_process';
import { accessSync, appendFileSync, constants } from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { setTimeout as delay } from 'node:timers/promises';
import { fileURLToPath } from 'node:url';
import { createArtifacts, freePort, hasFlag, option, parseTarget, resolveTargetConfig, runAllFlows, startServer } from './shared.mjs';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, '../../../..');
const target = parseTarget();
const buildApp = hasFlag('build-app');
const headed = option('headed') ?? process.env.AGENT_BROWSER_HEADED ?? 'true';

const { runId, artifactsRoot, screenshotsPath, tracesPath, downloadsPath, logsPath, databasePath } = createArtifacts(['downloads']);
const driver = createDriver({ screenshotsPath, tracesPath, downloadsPath, logsPath, headed });
let server;

try {
	console.log(`Artifacts: ${artifactsRoot}`);
	console.log(`Agent Browser session: ${driver.session}`);
	console.log(`Helium: ${driver.heliumPath}`);

	driver.verify();

	const port = await freePort();
	const app = resolveTargetConfig({ target, port, repoRoot, databasePath, runId, buildApp });

	server = await startServer(app, { target, logsPath, onReady: () => driver.startTrace() });

	await runAllFlows(driver, app.baseUrl);

	await driver.stopTrace('complete');
	await driver.close();
	await server.stop();

	console.log(`Inertia E2E passed for ${target}.`);
	console.log(`Artifacts: ${artifactsRoot}`);
	process.exit(0);
} catch (error) {
	console.error(error instanceof Error ? error.message : error);
	await driver.failureArtifacts().catch(() => {});
	await driver.close().catch(() => {});
	await server?.stop().catch(() => {});
	process.exit(1);
}

// Agent Browser driver adapter: implements the primitives the shared flows call
// by shelling out to the agent-browser CLI against a Helium engine.
function createDriver({ screenshotsPath, tracesPath, downloadsPath, logsPath, headed }) {
	const session = `alloy-inertia-demo-e2e-${process.pid}`;
	const browserBin = process.env.AGENT_BROWSER_BIN ?? 'agent-browser';
	const heliumPath = resolveHeliumPath();
	const commandLogPath = path.join(logsPath, 'agent-browser.log');

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

		if (result.error && !options.optional) {
			throw result.error;
		}

		if (result.status !== 0 && !options.optional) {
			throw new Error(`agent-browser ${command} failed (${result.status}):\n${result.stdout}\n${result.stderr}`);
		}

		return `${result.stdout ?? ''}${result.stderr ?? ''}`.trim();
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

	function extractUrl(output) {
		return (
			(output ?? '')
				.split(/\r?\n/u)
				.map((line) => line.trim())
				.find((line) => /^https?:\/\//u.test(line)) ?? ''
		);
	}

	return {
		session,
		heliumPath,

		verify() {
			const result = spawnSync(browserBin, ['--version'], {
				encoding: 'utf8',
				env: browserEnv,
			});

			if (result.error) {
				throw new Error(`agent-browser is not available via "${browserBin}": ${result.error.message}`);
			}

			if (result.status !== 0) {
				throw new Error(`agent-browser is not available via "${browserBin}".\n${result.stdout}\n${result.stderr}`);
			}
		},

		startTrace() {
			browser('trace', ['start']);
		},

		async open(url) {
			browser('open', [url]);
		},

		async openPath(baseUrl, routePath, text, options = {}) {
			browser('open', [`${baseUrl}${routePath}`]);
			browser('wait', ['--fn', `document.querySelector("#app") && document.body.innerText.trim().length > ${options.minTextLength ?? 60}`]);

			if (text) {
				browser('wait', ['--text', text]);
			}
		},

		async fill(selector, value) {
			browser('fill', [selector, value]);
		},

		async submitForm(anchorSelector) {
			const script = `(() => {
				const anchor = document.querySelector(${JSON.stringify(anchorSelector)});
				const form = anchor instanceof Element ? anchor.closest('form') : document.querySelector('form');
				if (!(form instanceof HTMLFormElement)) {
					throw new Error('missing form for anchor: ${anchorSelector}');
				}
				form.requestSubmit();
				return true;
			})()`;

			browser('eval', [script]);
		},

		async clickText(text) {
			browser('find', ['text', text, 'click', '--exact']);
		},

		async clickButtonText(text) {
			clickElementText('button', text);
		},

		async clickLinkText(text) {
			clickElementText('a', text);
		},

		async getInputValue(selector) {
			return browser('get', ['value', selector]);
		},

		async goBack() {
			browser('back');
		},

		async acceptNextDialog() {
			browser('eval', ['window.confirm = () => true']);
		},

		async waitForSelector(selector) {
			browser('wait', [selector]);
		},

		async waitForText(text) {
			browser('wait', ['--text', text]);
		},

		async waitForPath(expectedPath, timeoutMs = 30_000) {
			const deadline = Date.now() + timeoutMs;

			while (Date.now() < deadline) {
				const currentUrl = extractUrl(browser('get', ['url'], { optional: true }));

				if (currentUrl) {
					try {
						if (new URL(currentUrl).pathname === expectedPath) {
							return;
						}
					} catch {}
				}

				await delay(250);
			}

			throw new Error(`Timed out waiting for URL path ${expectedPath}. Current URL: ${extractUrl(browser('get', ['url'], { optional: true }))}`);
		},

		// Agent Browser has no dedicated response-wait/reset step; the flows above
		// wait on rendered text instead, so these hooks are no-ops here.
		beforeFlash() {
			return undefined;
		},

		async afterFlash() {},

		async settle() {},

		async resetPage() {},

		async screenshot(name) {
			const outputPath = path.join(screenshotsPath, `${name}.png`);
			browser('screenshot', ['--full', outputPath]);
			return outputPath;
		},

		assertClean(context) {
			const errors = browser('errors', ['--clear'], { optional: true });
			if (errors && !/no page errors/i.test(errors) && /error|exception|failed|rejected|typeerror|referenceerror/i.test(errors)) {
				throw new Error(`Browser page errors during ${context}:\n${errors}`);
			}

			browser('console', ['--clear'], { optional: true });
		},

		clearIssues() {
			browser('errors', ['--clear'], { optional: true });
			browser('console', ['--clear'], { optional: true });
			browser('network', ['requests', '--clear'], { optional: true });
		},

		async stopTrace(reason) {
			const outputPath = path.join(tracesPath, `${reason}.json`);
			browser('trace', ['stop', outputPath], { optional: true });
		},

		async close() {
			browser('close', [], { optional: true });
		},

		async failureArtifacts() {
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
				await this.screenshot('failure');
			} catch {}

			try {
				await this.stopTrace('failure');
			} catch {}
		},
	};
}

function resolveHeliumPath() {
	const candidates = [
		process.env.HELIUM_EXECUTABLE_PATH,
		process.env.AGENT_BROWSER_EXECUTABLE_PATH,
		path.join(process.env.HOME ?? '', 'Applications/Helium.app/Contents/MacOS/Helium'),
		path.join(process.env.HOME ?? '', 'Applications/Helium Browser.app/Contents/MacOS/Helium'),
	].filter(Boolean);

	for (const candidate of candidates) {
		try {
			accessSync(candidate, constants.X_OK);
			return candidate;
		} catch {}
	}

	throw new Error('Helium browser executable not found. Set HELIUM_EXECUTABLE_PATH (or AGENT_BROWSER_EXECUTABLE_PATH) to your Helium binary to run the agent-browser suite.');
}

function quoteShell(value) {
	if (/^[a-zA-Z0-9_./:=@-]+$/u.test(value)) {
		return value;
	}

	return `'${value.replaceAll("'", "'\\''")}'`;
}

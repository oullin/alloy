#!/usr/bin/env node
// Canonical, portable e2e runner: drives the Inertia demo through
// Playwright/Chromium. Requires no machine-specific browser install.
//
// The flows, route matrix and server lifecycle live in ./shared.mjs. This file
// is a thin driver adapter that implements the browser primitives the shared
// flows call using the Playwright API.

import { accessSync, appendFileSync, constants, writeFileSync } from 'node:fs';
import path from 'node:path';
import process from 'node:process';
import { fileURLToPath } from 'node:url';
import { chromium } from 'playwright';
import { createArtifacts, freePort, hasFlag, isCI, option, parseTarget, resolveTargetConfig, runAllFlows, startServer } from './shared.mjs';

const __dirname = path.dirname(fileURLToPath(import.meta.url));
const repoRoot = path.resolve(__dirname, '../../../..');
const target = parseTarget();
const buildApp = hasFlag('build-app');
const headed = option('headed') ?? process.env.PLAYWRIGHT_HEADED ?? (isCI() ? 'false' : 'true');

const { runId, artifactsRoot, screenshotsPath, tracesPath, logsPath, databasePath } = createArtifacts();
const driver = createDriver({ screenshotsPath, tracesPath, logsPath, headed });
let server;

try {
	console.log(`Artifacts: ${artifactsRoot}`);
	console.log(`Playwright headed: ${headed}`);

	const port = await freePort();
	const app = resolveTargetConfig({ target, port, repoRoot, databasePath, runId, buildApp });

	server = await startServer(app, { target, logsPath });
	await driver.launch();
	await driver.open(`${app.baseUrl}/login`);

	await runAllFlows(driver, app.baseUrl, app.seedPassword);

	await driver.stopTrace('complete');
	await driver.close();
	await server.stop();

	console.log(`Inertia Playwright E2E passed for ${target}.`);
	console.log(`Artifacts: ${artifactsRoot}`);
	process.exit(0);
} catch (error) {
	console.error(error instanceof Error ? error.message : error);
	await driver.failureArtifacts().catch(() => {});
	await driver.close().catch(() => {});
	await server?.stop().catch(() => {});
	process.exit(1);
}

// Playwright driver adapter: implements the primitives the shared flows call.
function createDriver({ screenshotsPath, tracesPath, logsPath, headed }) {
	let browser;
	let context;
	let page;
	let pageIssues = [];

	const locator = (selector) => page.locator(selector).first();

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

	async function waitForResponse(pathname, method) {
		await page.waitForResponse((response) => {
			const url = new URL(response.url());

			return url.pathname === pathname && response.request().method() === method && response.status() < 500;
		});
	}

	function clearIssues() {
		pageIssues = [];
	}

	return {
		async launch() {
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
		},

		async open(url) {
			await page.goto(url, { waitUntil: 'domcontentloaded' });
		},

		async openPath(baseUrl, routePath, text, options = {}) {
			try {
				await page.goto(`${baseUrl}${routePath}`, { waitUntil: 'domcontentloaded' });
				await this.waitForPath(routePath);
				await page.waitForFunction((minTextLength) => document.querySelector('#app') && document.body.innerText.trim().length > minTextLength, options.minTextLength ?? 60);

				if (text) {
					await this.waitForText(text);
				}
			} catch (error) {
				const reason = error instanceof Error ? error.message : String(error);
				throw new Error(`open ${routePath} failed at ${page?.url() ?? 'no page'}: ${reason}`);
			}
		},

		async fill(selector, value) {
			await locator(selector).fill(value);
		},

		async submitForm(anchorSelector) {
			await locator(anchorSelector).evaluate((element) => {
				const form = element.closest('form');
				if (!(form instanceof HTMLFormElement)) {
					throw new Error('missing form');
				}

				form.requestSubmit();
			});
		},

		async clickText(text) {
			await page.getByText(text, { exact: true }).first().click();
		},

		async clickButtonText(text) {
			await page.getByRole('button', { name: text, exact: true }).click();
		},

		async clickLinkText(text) {
			await page.getByRole('link', { name: text, exact: true }).click();
		},

		async getInputValue(selector) {
			return locator(selector).inputValue();
		},

		async goBack() {
			await page.goBack({ waitUntil: 'domcontentloaded' });
		},

		async acceptNextDialog() {
			page.once('dialog', (dialog) => dialog.accept());
		},

		async waitForSelector(selector) {
			await locator(selector).waitFor({ state: 'visible' });
		},

		async waitForText(text) {
			try {
				await page.waitForFunction((expected) => document.body.innerText.includes(expected), text);
			} catch (error) {
				const reason = error instanceof Error ? error.message : String(error);
				throw new Error(`wait for text "${text}" failed at ${page?.url() ?? 'no page'}: ${reason}`);
			}
		},

		async waitForPath(pathname) {
			try {
				await page.waitForFunction((expected) => window.location.pathname === expected, pathname);
			} catch (error) {
				const reason = error instanceof Error ? error.message : String(error);
				throw new Error(`wait for path "${pathname}" failed at ${page?.url() ?? 'no page'}: ${reason}`);
			}
		},

		beforeFlash(pathname) {
			return Promise.all([waitForResponse(pathname, 'POST'), waitForResponse(pathname, 'GET')]);
		},

		async afterFlash(handle) {
			await handle;
			await this.waitForPath('/features/state/flash-data');
		},

		async settle() {
			await page.waitForLoadState('networkidle', { timeout: 10_000 }).catch(() => {});
			await page.waitForTimeout(250);
		},

		async resetPage() {
			await page?.close().catch(() => {});
			clearIssues();
			await createPage();
		},

		async screenshot(name) {
			const outputPath = path.join(screenshotsPath, `${name}.png`);
			await page.screenshot({ path: outputPath, fullPage: true });
			return outputPath;
		},

		assertClean(contextName) {
			const issues = pageIssues.filter((issue) => /error|exception|failed|rejected|typeerror|referenceerror/iu.test(issue));
			if (issues.length) {
				throw new Error(`Browser page issues during ${contextName}:\n${issues.join('\n')}`);
			}

			clearIssues();
		},

		clearIssues,

		async stopTrace(reason) {
			if (!context) {
				return;
			}

			await context.tracing.stop({ path: path.join(tracesPath, `${reason}.zip`) });
		},

		async close() {
			await context?.close().catch(() => {});
			await browser?.close().catch(() => {});
			context = undefined;
			browser = undefined;
			page = undefined;
		},

		async failureArtifacts() {
			if (!page) {
				return;
			}

			writeFileSync(path.join(logsPath, 'failure-page.html'), await page.content());
			appendFileSync(path.join(logsPath, 'failure-page.log'), `url: ${page.url()}\n`);
			appendFileSync(path.join(logsPath, 'browser-issues.log'), `${pageIssues.join('\n')}\n`);

			try {
				await this.screenshot('failure');
			} catch {}

			try {
				await this.stopTrace('failure');
			} catch {}
		},
	};
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

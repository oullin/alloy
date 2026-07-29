// Validates a packed .tgz before it is uploaded to the Hara registry.
//
// The packages are consumed by paying customers through standard `pnpm add`,
// so a manifest that only resolves inside this workspace is a shipped bug. The
// checks below encode the failures that actually occurred here: sdk/tempo
// declared exports that its build never emitted, and workspace-only specifiers
// resolve to nothing once a tarball leaves the monorepo.

import { execFileSync } from 'node:child_process';

const REGISTRY = 'https://hara.sh/npm/';
const NAME_PATTERN = /^@hara\/sdk-[a-z][a-z0-9-]*$/u;
const LOCAL_PROTOCOLS = ['workspace:', 'file:', 'link:', 'portal:'];
const DEPENDENCY_FIELDS = ['dependencies', 'devDependencies', 'peerDependencies', 'optionalDependencies'];

const [tarball, expectedVersion] = process.argv.slice(2);

if (!tarball || !expectedVersion) {
	console.error('usage: check-packed-manifest.mjs <tarball> <expected-version>');
	process.exit(2);
}

const failures = [];
const fail = (message) => failures.push(message);

const entries = execFileSync('tar', ['-tzf', tarball], { encoding: 'utf8' })
	.split('\n')
	.filter(Boolean)
	// Directory entries carry a trailing slash; normalise so lookups match files only.
	.map((entry) => entry.replace(/\/$/u, ''));

const entrySet = new Set(entries);
const manifest = JSON.parse(execFileSync('tar', ['-xzOf', tarball, 'package/package.json'], { encoding: 'utf8' }));

if (!NAME_PATTERN.test(manifest.name)) {
	fail(`name "${manifest.name}" is not a valid published Hara SDK name (expected @hara/sdk-<name>)`);
}

if (manifest.version !== expectedVersion) {
	fail(`version "${manifest.version}" does not match release version "${expectedVersion}"`);
}

// Kept deliberately: the pipeline uploads artifacts directly, so `private` costs
// nothing at install time and blocks an accidental `pnpm publish` to npmjs.com,
// which LICENSE.md forbids.
if (manifest.private !== true) {
	fail('"private": true is missing; it guards against publishing to a public registry');
}

if (manifest.publishConfig?.registry !== REGISTRY) {
	fail(`publishConfig.registry is "${manifest.publishConfig?.registry}", expected "${REGISTRY}"`);
}

for (const field of DEPENDENCY_FIELDS) {
	for (const [name, range] of Object.entries(manifest[field] ?? {})) {
		const protocol = LOCAL_PROTOCOLS.find((candidate) => String(range).startsWith(candidate));

		if (protocol) {
			fail(`${field}["${name}"] is "${range}"; "${protocol}" does not resolve outside this workspace`);
		}
	}
}

// Walks exports/imports, which nest arbitrarily deep through condition objects
// and subpath keys, and yields every relative target the map can resolve to.
const collectTargets = (node, trail, out) => {
	if (typeof node === 'string') {
		out.push({ target: node, trail });
		return;
	}

	if (Array.isArray(node)) {
		node.forEach((item, index) => collectTargets(item, `${trail}[${index}]`, out));
		return;
	}

	if (node && typeof node === 'object') {
		for (const [key, value] of Object.entries(node)) {
			collectTargets(value, `${trail}.${key}`, out);
		}
	}
};

for (const field of ['exports', 'imports']) {
	const targets = [];
	collectTargets(manifest[field], field, targets);

	for (const { target, trail } of targets) {
		if (!target.startsWith('./')) {
			continue;
		}

		// A wildcard target stands for many files; assert the directory it roots
		// shipped at all, which is what catches a whole missing build output.
		if (target.includes('*')) {
			const prefix = `package/${target.slice(2).split('*')[0]}`;

			if (!entries.some((entry) => entry.startsWith(prefix))) {
				fail(`${trail} -> "${target}" matches nothing in the tarball`);
			}

			continue;
		}

		if (!entrySet.has(`package/${target.slice(2)}`)) {
			fail(`${trail} -> "${target}" is not in the tarball`);
		}
	}
}

if (failures.length > 0) {
	console.error(`\n${manifest.name} (${tarball}) is not publishable:`);

	for (const failure of failures) {
		console.error(`  - ${failure}`);
	}

	process.exit(1);
}

console.log(`  ok ${manifest.name}@${manifest.version} (${entries.length} entries)`);

// Gates are-the-types-wrong output on the resolution modes customers actually use.
//
// Every SDK package is "type": "module" and requires node >=22, so problems
// reported for node10 (which predates "exports") and node16-cjs (`require` of an
// ESM package) are expected and not defects. attw exits non-zero for those
// regardless of --profile, and blanket --ignore-rules would also hide the same
// problem kinds in ESM mode, which is exactly what this gate exists to catch.
// So: read the JSON and fail only on the modes that matter.

import { readFileSync } from 'node:fs';

const SUPPORTED_RESOLUTIONS = new Set(['node16-esm', 'bundler']);

const [reportPath, label = 'package'] = process.argv.slice(2);

if (!reportPath) {
	console.error('usage: check-packed-types.mjs <attw-json-report> [label]');
	process.exit(2);
}

// Read from a file, never a pipe. are-the-types-wrong exits without draining
// stdout, so piping its JSON silently truncates at ~128KB -- the console report
// is over 800KB. Redirecting to a file writes synchronously and stays intact.
const raw = readFileSync(reportPath, 'utf8');

// `pnpm dlx` prints installation progress before the JSON on a cold cache.
const start = raw.indexOf('{');
let report;

try {
	report = JSON.parse(start === -1 ? raw : raw.slice(start));
} catch {
	console.error(`could not parse are-the-types-wrong output for ${label}`);
	console.error(raw.slice(0, 2000));
	process.exit(1);
}

// attw reports a top-level `problems` key on failure to analyse at all, which is
// a different shape from the per-entrypoint analysis and must not pass silently.
if (report.analysis === undefined) {
	console.error(`are-the-types-wrong could not analyse ${label}:`);
	console.error(JSON.stringify(report, null, 2).slice(0, 2000));
	process.exit(1);
}

// Two different keys are called "problems": the top level groups them by kind,
// while analysis.problems is the flat list carrying resolutionKind. Use the flat
// one -- the grouped entries label the mode `resolutionOption` instead.
const problems = report.analysis.problems ?? [];
const blocking = problems.filter((problem) => SUPPORTED_RESOLUTIONS.has(problem.resolutionKind));

if (blocking.length > 0) {
	console.error(`\n${label} has type-resolution problems in supported modes:`);

	for (const problem of blocking) {
		console.error(`  - ${problem.kind} at "${problem.entrypoint}" (${problem.resolutionKind})`);
	}

	process.exit(1);
}

const ignored = problems.length - blocking.length;
console.log(`  ok ${report.analysis.packageName} types resolve for node16-esm and bundler (${ignored} node10/cjs problems ignored)`);

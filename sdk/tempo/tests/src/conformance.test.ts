import { readFileSync } from 'node:fs';

import { Tempo } from '@hara/sdk-tempo';
import { describe, expect, it } from 'vite-plus/test';

// Shared Go<->TS golden vectors. This is the TS half of the cross-runtime drift
// guard (plan 008); the Go half lives in pkg/hub/tempo/conformance_test.go and
// reads the same JSON. Each case builds an instant from calendar components in a
// named zone, applies one operation, and renders the result as a string.

interface TempoConformanceComponents {
	year: number;
	month: number;
	day: number;
	hour?: number;
	minute?: number;
	second?: number;
	timeZone: string;
}

interface TempoConformanceCase {
	op: string;
	base?: TempoConformanceComponents;
	other?: TempoConformanceComponents;
	arg?: number;
	input?: string;
	pattern?: string;
	timeZone?: string;
	render?: string;
	expected: string;
	note: string;
}

interface TempoConformanceFile {
	cases: TempoConformanceCase[];
}

const fixturePath = new URL('../../../../conformance/tempo.json', import.meta.url);
const fixture = JSON.parse(readFileSync(fixturePath, 'utf8')) as TempoConformanceFile;

const buildTempo = (components: TempoConformanceComponents): Tempo =>
	Tempo.create({
		day: components.day,
		hour: components.hour ?? 0,
		minute: components.minute ?? 0,
		month: components.month,
		second: components.second ?? 0,
		timeZone: components.timeZone,
		year: components.year,
	});

const render = (value: Tempo, mode: string | undefined): string => {
	switch (mode) {
		case 'iso':
			return value.toISOString();

		case 'date':
			return value.toDateString();

		default:
			throw new Error(`unknown tempo render mode: ${mode}`);
	}
};

const runTempoOp = (testCase: TempoConformanceCase): string => {
	const { arg, base, op, other } = testCase;

	switch (op) {
		case 'addDays':
			return render(buildTempo(base as TempoConformanceComponents).addDays(arg as number), testCase.render);

		case 'addWeeks':
			return render(buildTempo(base as TempoConformanceComponents).addWeeks(arg as number), testCase.render);

		case 'addHours':
			return render(buildTempo(base as TempoConformanceComponents).addHours(arg as number), testCase.render);

		case 'addMonths':
			return render(buildTempo(base as TempoConformanceComponents).addMonths(arg as number), testCase.render);

		case 'addMonthsNoOverflow':
			return render(buildTempo(base as TempoConformanceComponents).addMonthsNoOverflow(arg as number), testCase.render);

		case 'diffInMonths':
			return String(buildTempo(base as TempoConformanceComponents).diffInMonths(buildTempo(other as TempoConformanceComponents)));

		case 'diffInYears':
			return String(buildTempo(base as TempoConformanceComponents).diffInYears(buildTempo(other as TempoConformanceComponents)));

		case 'parseFromPattern':
			return render(Tempo.fromFormat(testCase.input as string, testCase.pattern as string, { timeZone: testCase.timeZone }), testCase.render);

		default:
			throw new Error(`unknown tempo conformance op: ${op}`);
	}
};

describe('tempo cross-runtime conformance', () => {
	expect(fixture.cases.length).toBeGreaterThan(0);

	for (const testCase of fixture.cases) {
		it(`${testCase.op}: ${testCase.note}`, () => {
			expect(runTempoOp(testCase), testCase.note).toBe(testCase.expected);
		});
	}
});

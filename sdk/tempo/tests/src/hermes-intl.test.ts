import { beforeAll, describe, expect, it } from 'vite-plus/test';

/**
 * Regression guard for runtimes whose `Intl` is partial.
 *
 * `Intl.DateTimeFormat` ignores an option it does not implement rather than
 * throwing, so a formatter asked for `hourCycle: 'h23'` can answer in the
 * locale's own cycle instead. React Native's Hermes is the engine that does
 * this in practice. Before the fix this file guards, every construction path
 * in the package threw `RangeError: Invalid Tempo local date/time components`
 * there — including plain `Tempo.parse('2026-08-19')` — which made the whole
 * SDK unusable on a mobile client.
 *
 * The stub is installed before the SDK is imported because `getFormatter`
 * memoises the formatters it builds; a stub applied later would be bypassed by
 * the cache.
 */
/** Exactly what a partial implementation does: keep what it knows, drop the rest. */
const supportedOptionsOnly = (options?: Intl.DateTimeFormatOptions): Intl.DateTimeFormatOptions | undefined => {
	if (options === undefined) {
		return undefined;
	}

	const { hourCycle: _hourCycle, fractionalSecondDigits: _fractionalSecondDigits, ...supported } = options;

	return supported;
};

const stripUnsupportedOptions = (): void => {
	// A Proxy rather than a wrapper function so the prototype and the static
	// `supportedLocalesOf` keep working without being reassigned.
	Intl.DateTimeFormat = new Proxy(Intl.DateTimeFormat, {
		apply: (target, _thisArg, args: [string?, Intl.DateTimeFormatOptions?]) => target(args[0], supportedOptionsOnly(args[1])),
		construct: (target, args: [string?, Intl.DateTimeFormatOptions?]) => new target(args[0], supportedOptionsOnly(args[1])),
	});
};

let Tempo: typeof import('@hara/sdk-tempo').Tempo;

beforeAll(async () => {
	stripUnsupportedOptions();
	({ Tempo } = await import('@hara/sdk-tempo'));
});

describe('Tempo on a runtime that ignores hourCycle', () => {
	it('parses an ISO calendar day', () => {
		expect(Tempo.parse('2026-08-19').toDateString()).toBe('2026-08-19');
	});

	it('builds strict components', () => {
		expect(Tempo.createStrict({ day: 19, month: 8, year: 2026 }).toDateString()).toBe('2026-08-19');
	});

	it('reads back the components it was given', () => {
		const day = Tempo.createStrict({ day: 19, month: 8, year: 2026 });

		expect(day.year).toBe(2026);
		expect(day.month).toBe(8);
		expect(day.day).toBe(19);
	});

	it('keeps midnight on the same day', () => {
		// The h12 rendering of midnight is `12 AM`. Reading that as hour 12 is
		// what moved the instant half a day and failed the safety assertion.
		const midnight = Tempo.createStrict({ day: 19, hour: 0, month: 8, year: 2026 });

		expect(midnight.hour).toBe(0);
		expect(midnight.toDateString()).toBe('2026-08-19');
	});

	it('keeps noon at hour twelve', () => {
		expect(Tempo.createStrict({ day: 19, hour: 12, month: 8, year: 2026 }).hour).toBe(12);
	});

	it('reads an afternoon hour as h23', () => {
		expect(Tempo.createStrict({ day: 19, hour: 23, month: 8, year: 2026 }).hour).toBe(23);
	});

	it('adds days across a month boundary', () => {
		expect(Tempo.createStrict({ day: 31, month: 8, year: 2026 }).addDays(1).toDateString()).toBe('2026-09-01');
	});

	it('still refuses a day the calendar does not have', () => {
		expect(() => Tempo.createStrict({ day: 31, month: 2, year: 2026 })).toThrow(RangeError);
	});
});

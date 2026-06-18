import { defaultTimeZone, normalizeTimeZone } from '../calendar';

import type { HumanDiffOptions, TempoInput, TempoOptions, TempoPolicy, TempoSettings } from '../types';

export const defaultHumanDiffOptions: HumanDiffOptions = {
	locale: 'en-US',
	numeric: 'always',
	style: 'long',
};

export const defaultTempoPolicy = (): TempoPolicy => ({
	fallbackLocale: 'en-US',
	humanDiffOptions: { ...defaultHumanDiffOptions },
	locale: 'en-US',
	midDayAt: 12,
	monthsOverflow: true,
	runtime: undefined,
	serializer: null,
	strictMode: true,
	testNow: null,
	timeZone: defaultTimeZone,
	toStringFormat: null,
	translator: undefined,
	weekendDays: [0, 6],
	yearsOverflow: true,
});

export const cloneTempoPolicy = (policy: TempoPolicy): TempoPolicy => ({
	...policy,
	humanDiffOptions: { ...policy.humanDiffOptions },
	testNow: policy.testNow === null ? null : new Date(policy.testNow.getTime()),
	weekendDays: [...policy.weekendDays],
});

export const normalizeWeekendDays = (days: readonly number[]): readonly number[] => days.map((day) => ((Math.trunc(day) % 7) + 7) % 7);

const dateFromInput = (input: TempoInput): Date => {
	if (input instanceof Date) {
		return new Date(input.getTime());
	}

	if (typeof input === 'object' && input !== null && 'toDate' in input && typeof (input as { readonly toDate?: unknown }).toDate === 'function') {
		return (input as { toDate: () => Date }).toDate();
	}

	const date = new Date(input as string | number);

	if (Number.isNaN(date.getTime())) {
		throw new RangeError(`Invalid Tempo testNow input: ${String(input)}`);
	}

	return date;
};

export const resolveTempoPolicy = (options?: TempoOptions, base: TempoPolicy = defaultTempoPolicy()): TempoPolicy => {
	const policy = cloneTempoPolicy(base);

	if (options === undefined) {
		return policy;
	}

	return {
		...policy,
		fallbackLocale: options.fallbackLocale ?? policy.fallbackLocale,
		humanDiffOptions: options.humanDiffOptions === undefined ? policy.humanDiffOptions : { ...policy.humanDiffOptions, ...options.humanDiffOptions },
		locale: options.locale ?? policy.locale,
		midDayAt: options.midDayAt ?? policy.midDayAt,
		monthsOverflow: options.monthsOverflow ?? policy.monthsOverflow,
		runtime: options.runtime ?? policy.runtime,
		serializer: options.serializer === undefined ? policy.serializer : options.serializer,
		strictMode: options.strictMode ?? policy.strictMode,
		testNow: options.testNow === undefined ? policy.testNow : options.testNow === null ? null : dateFromInput(options.testNow),
		timeZone: normalizeTimeZone(options.timeZone ?? policy.timeZone),
		toStringFormat: options.toStringFormat === undefined ? policy.toStringFormat : options.toStringFormat,
		translator: options.translator ?? policy.translator,
		weekendDays: options.weekendDays === undefined ? policy.weekendDays : normalizeWeekendDays(options.weekendDays),
		yearsOverflow: options.yearsOverflow ?? policy.yearsOverflow,
	};
};

export const policyToSettings = (policy: TempoPolicy): TempoSettings => ({
	fallbackLocale: policy.fallbackLocale,
	humanDiffOptions: { ...policy.humanDiffOptions },
	locale: policy.locale,
	midDayAt: policy.midDayAt,
	monthsOverflow: policy.monthsOverflow,
	strictMode: policy.strictMode,
	testNow: policy.testNow === null ? null : new Date(policy.testNow.getTime()),
	timeZone: policy.timeZone,
	weekendDays: [...policy.weekendDays],
	yearsOverflow: policy.yearsOverflow,
});

export class TempoClock {
	private readonly policy: TempoPolicy;

	constructor(policy: TempoPolicy = defaultTempoPolicy()) {
		this.policy = cloneTempoPolicy(policy);
	}

	now(): Date {
		return this.policy.testNow === null ? new Date() : new Date(this.policy.testNow.getTime());
	}
}

export class TempoSettingsStore {
	private readonly policy: TempoPolicy;

	constructor(policy: TempoPolicy = defaultTempoPolicy()) {
		this.policy = cloneTempoPolicy(policy);
	}

	snapshot(): TempoSettings {
		return policyToSettings(this.policy);
	}
}

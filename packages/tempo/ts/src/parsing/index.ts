import type { TempoComponents, TempoInput, TempoOptions } from '../types';
import { tempoConfig } from '../config';
import { TempoRuntime } from '../runtime';

import {
	assertFiniteNumber,
	dateFromPartsAsUTC,
	dateFromZonedComponents,
	isoDatePattern,
	isoLocalPattern,
	millisecondsPerMinute,
	monthNumberFromName,
	normalizeTimeZone,
	timezonePattern,
} from '../calendar';

type TempoDateCarrier = {
	readonly timeZone: string;
	toDate: () => Date;
};

type TempoRuntimeCarrier = TempoDateCarrier & {
	getRuntime: () => TempoRuntime;
};

const isTempoDateCarrier = (input: unknown): input is TempoDateCarrier =>
	typeof input === 'object' &&
	input !== null &&
	'timeZone' in input &&
	typeof (input as { readonly timeZone?: unknown }).timeZone === 'string' &&
	typeof (input as { readonly toDate?: unknown }).toDate === 'function';

const isTempoRuntimeCarrier = (input: unknown): input is TempoRuntimeCarrier => isTempoDateCarrier(input) && typeof (input as { readonly getRuntime?: unknown }).getRuntime === 'function';

export const runtimeFromOptions = (input: TempoInput | undefined, options?: TempoOptions): TempoRuntime => {
	const base =
		options?.runtime ??
		(isTempoRuntimeCarrier(input)
			? input.getRuntime()
			: new TempoRuntime({
					fallbackLocale: tempoConfig.fallbackLocale,
					locale: tempoConfig.locale,
				}));

	return options?.translator !== undefined || options?.locale !== undefined || options?.fallbackLocale !== undefined
		? base.with({
				fallbackLocale: options.fallbackLocale,
				locale: options.locale,
				translator: options.translator,
			})
		: base;
};

export const fromNumericTimestamp = (input: number): Date => {
	assertFiniteNumber(input, 'Timestamp');

	const magnitude = Math.abs(input);
	const milliseconds = magnitude < 10_000_000_000 ? input * 1000 : input;

	return new Date(milliseconds);
};

export const parseLocalText = (input: string, timeZone: string): Date | null => {
	const dateOnly = isoDatePattern.exec(input);

	if (dateOnly !== null) {
		return dateFromZonedComponents(
			{
				day: Number(dateOnly[3]),
				month: Number(dateOnly[2]),
				timeZone,
				year: Number(dateOnly[1]),
			},
			timeZone,
		);
	}

	if (timezonePattern.test(input)) {
		return null;
	}

	const local = isoLocalPattern.exec(input);

	if (local === null) {
		return null;
	}

	return dateFromZonedComponents(
		{
			day: Number(local[3]),
			hour: Number(local[4] ?? 0),
			millisecond: Number((local[7] ?? '0').slice(0, 3).padEnd(3, '0')),
			minute: Number(local[5] ?? 0),
			month: Number(local[2]),
			second: Number(local[6] ?? 0),
			timeZone,
			year: Number(local[1]),
		},
		timeZone,
	);
};

export const zoneFromInput = (input: TempoInput, options: TempoOptions | undefined): string => {
	if (isTempoDateCarrier(input)) {
		return normalizeTimeZone(options?.timeZone ?? input.timeZone);
	}

	return normalizeTimeZone(options?.timeZone);
};

export const asDate = (input: TempoInput, options?: TempoOptions): Date => {
	if (isTempoDateCarrier(input)) {
		return input.toDate();
	}

	if (input instanceof Date) {
		return new Date(input.getTime());
	}

	const timeZone = normalizeTimeZone(options?.timeZone);

	const date = typeof input === 'number' ? fromNumericTimestamp(input) : (parseLocalText(input, timeZone) ?? new Date(input));

	if (Number.isNaN(date.getTime())) {
		throw new RangeError(`Invalid Tempo input: ${String(input)}`);
	}

	return date;
};

export const escapePattern = (input: string): string => input.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');

export const parseOffsetMinutes = (input: string): number => {
	if (input === 'Z') {
		return 0;
	}

	const match = /^([+-])(\d{2}):?(\d{2})$/.exec(input);

	if (match === null) {
		throw new RangeError(`Invalid Tempo offset: ${input}`);
	}

	const minutes = Number(match[2]) * 60 + Number(match[3]);

	return match[1] === '-' ? -minutes : minutes;
};

export const parseFromPattern = (input: string, pattern: string, options?: TempoOptions): Date => {
	const tokens = ['YYYY', 'MMMM', 'dddd', 'MMM', 'ddd', 'SSS', 'Do', 'YY', 'ZZ', 'MM', 'DD', 'HH', 'hh', 'mm', 'ss', 'Z', 'M', 'D', 'H', 'h', 'm', 's', 'A', 'a'] as const;

	const groups: string[] = [];

	let expression = '^';

	for (let index = 0; index < pattern.length; ) {
		if (pattern[index] === '[') {
			const end = pattern.indexOf(']', index);

			if (end >= 0) {
				expression += escapePattern(pattern.slice(index + 1, end));
				index = end + 1;
				continue;
			}
		}

		const token = tokens.find((item) => pattern.slice(index).startsWith(item));

		if (token === undefined) {
			expression += escapePattern(pattern[index] ?? '');
			index += 1;
			continue;
		}

		groups.push(token);
		expression +=
			token === 'A' || token === 'a'
				? '(AM|PM|am|pm)'
				: token === 'MMM' || token === 'MMMM'
					? '([\\p{L}.]+)'
					: token === 'ddd' || token === 'dddd'
						? '([\\p{L}.]+)'
						: token === 'Do'
							? '(\\d{1,2})(?:st|nd|rd|th)'
							: token === 'Z'
								? '(Z|[+-]\\d{2}:\\d{2})'
								: token === 'ZZ'
									? '(Z|[+-]\\d{4})'
									: '(\\d{1,' + (token.length === 1 ? '4' : String(token.length)) + '})';
		index += token.length;
	}

	expression += '$';

	const match = new RegExp(expression, 'u').exec(input);

	if (match === null) {
		throw new RangeError(`Input does not match Tempo format: ${input}`);
	}

	const values = new Map<string, string>();

	groups.forEach((token, index) => values.set(token, match[index + 1] ?? ''));

	let year = values.has('YYYY') ? Number(values.get('YYYY')) : values.has('YY') ? 2000 + Number(values.get('YY')) : 1970;

	const meridiem = values.get('A') ?? values.get('a');

	let hour = Number(values.get('HH') ?? values.get('H') ?? values.get('hh') ?? values.get('h') ?? 0);

	if (meridiem !== undefined) {
		const lower = meridiem.toLowerCase();

		if (lower === 'pm' && hour < 12) {
			hour += 12;
		}

		if (lower === 'am' && hour === 12) {
			hour = 0;
		}
	}

	if (!Number.isFinite(year)) {
		year = 1970;
	}

	const components: TempoComponents = {
		day: Number(values.get('DD') ?? values.get('Do') ?? values.get('D') ?? 1),
		hour,
		millisecond: Number((values.get('SSS') ?? '0').slice(0, 3).padEnd(3, '0')),
		minute: Number(values.get('mm') ?? values.get('m') ?? 0),
		month: monthNumberFromName(values.get('MMMM') ?? values.get('MMM') ?? '') ?? Number(values.get('MM') ?? values.get('M') ?? 1),
		second: Number(values.get('ss') ?? values.get('s') ?? 0),
		timeZone: options?.timeZone,
		year,
	};

	const offset = values.get('Z') ?? values.get('ZZ');

	if (offset !== undefined && offset !== '') {
		const offsetMinutes = parseOffsetMinutes(offset);

		return new Date(dateFromPartsAsUTC(components) - offsetMinutes * millisecondsPerMinute);
	}

	return dateFromZonedComponents(components, options?.timeZone);
};

export class TempoParser {
	runtimeFromOptions(input: TempoInput | undefined, options?: TempoOptions): TempoRuntime {
		return runtimeFromOptions(input, options);
	}

	asDate(input: TempoInput, options?: TempoOptions): Date {
		return asDate(input, options);
	}

	zoneFromInput(input: TempoInput, options?: TempoOptions): string {
		return zoneFromInput(input, options);
	}

	fromNumericTimestamp(input: number): Date {
		return fromNumericTimestamp(input);
	}

	parseFromPattern(input: string, pattern: string, options?: TempoOptions): Date {
		return parseFromPattern(input, pattern, options);
	}
}

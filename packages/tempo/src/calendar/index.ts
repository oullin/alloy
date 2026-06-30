import type { TempoComponents, TimeUnit, TimeZoneNameStyle, WeekdayInput } from '#types';

export type ZonedParts = {
	readonly year: number;
	readonly month: number;
	readonly day: number;
	readonly hour: number;
	readonly minute: number;
	readonly second: number;
	readonly millisecond: number;
	readonly weekday: number;
};

export type DateTimePart = Intl.DateTimeFormatPart;

export const defaultTimeZone = 'UTC';

export const millisecondsPerSecond = 1000;

export const millisecondsPerMinute = 60 * millisecondsPerSecond;

export const millisecondsPerHour = 60 * millisecondsPerMinute;

export const millisecondsPerDay = 24 * millisecondsPerHour;

export const millisecondsPerWeek = 7 * millisecondsPerDay;

export const isoDatePattern = /^(\d{4})-(\d{2})-(\d{2})$/;

export const isoLocalPattern = /^(\d{4})-(\d{2})-(\d{2})(?:[T\s](\d{2})(?::?(\d{2}))?(?::?(\d{2})(?:\.(\d{1,9}))?)?)?$/;

export const isoDurationPattern =
	/^(-)?P(?:(\d+(?:\.\d+)?)Y)?(?:(\d+(?:\.\d+)?)M)?(?:(\d+(?:\.\d+)?)W)?(?:(\d+(?:\.\d+)?)D)?(?:T(?:(\d+(?:\.\d+)?)H)?(?:(\d+(?:\.\d+)?)M)?(?:(\d+(?:\.\d+)?)S)?)?$/i;

export const timezonePattern = /(?:Z|[+-]\d{2}:?\d{2})$/i;

const formatterCache = new Map<string, Intl.DateTimeFormat>();
const monthNameCache = new Map<string, readonly string[]>();
const weekdayNameCache = new Map<string, readonly string[]>();

export const pad = (value: number, length = 2): string => String(Math.trunc(Math.abs(value))).padStart(length, '0');

export const assertFiniteNumber = (value: number, label: string): void => {
	if (!Number.isFinite(value)) {
		throw new RangeError(`${label} must be a finite number`);
	}
};

export const normalizeTimeZone = (timeZone: string | undefined): string => {
	const zone = timeZone ?? defaultTimeZone;

	try {
		Intl.DateTimeFormat('en-US', { timeZone: zone }).format(new Date(0));
	} catch (error) {
		throw new RangeError(`Invalid Tempo time zone: ${zone}`, {
			cause: error,
		});
	}

	return zone;
};

export const getFormatter = (timeZone: string, timeZoneName?: TimeZoneNameStyle, locale = 'en-US-u-nu-latn'): Intl.DateTimeFormat => {
	const key = `${locale}|${timeZone}|${timeZoneName ?? 'parts'}`;
	const cached = formatterCache.get(key);

	if (cached !== undefined) {
		return cached;
	}

	const formatter = new Intl.DateTimeFormat(locale, {
		calendar: 'gregory',
		day: '2-digit',
		fractionalSecondDigits: 3,
		hour: '2-digit',
		hourCycle: 'h23',
		minute: '2-digit',
		month: '2-digit',
		second: '2-digit',
		timeZone,
		timeZoneName,
		weekday: 'short',
		year: 'numeric',
	});

	formatterCache.set(key, formatter);

	return formatter;
};

export const readPart = (parts: readonly DateTimePart[], type: string): string => {
	const part = parts.find((item) => item.type === type);

	return part?.value ?? '';
};

export const weekdayIndex = (weekday: string): number => {
	switch (weekday) {
		case 'Sun':
			return 0;

		case 'Mon':
			return 1;

		case 'Tue':
			return 2;

		case 'Wed':
			return 3;

		case 'Thu':
			return 4;

		case 'Fri':
			return 5;

		case 'Sat':
			return 6;

		default:
			return 0;
	}
};

export const getZonedParts = (date: Date, timeZone: string): ZonedParts => {
	const parts = getFormatter(timeZone).formatToParts(date);

	return {
		year: Number(readPart(parts, 'year')),
		month: Number(readPart(parts, 'month')),
		day: Number(readPart(parts, 'day')),
		hour: Number(readPart(parts, 'hour')),
		minute: Number(readPart(parts, 'minute')),
		second: Number(readPart(parts, 'second')),
		millisecond: Number(readPart(parts, 'fractionalSecond') || '0'),
		weekday: weekdayIndex(readPart(parts, 'weekday')),
	};
};

export const dateFromPartsAsUTC = (parts: TempoComponents): number =>
	Date.UTC(parts.year, (parts.month ?? 1) - 1, parts.day ?? 1, parts.hour ?? 0, parts.minute ?? 0, parts.second ?? 0, parts.millisecond ?? 0);

export const dateFromZonedComponents = (components: TempoComponents, fallbackTimeZone = defaultTimeZone): Date => {
	const timeZone = normalizeTimeZone(components.timeZone ?? fallbackTimeZone);

	let utc = dateFromPartsAsUTC(components);

	const target = dateFromPartsAsUTC(components);

	for (let index = 0; index < 4; index += 1) {
		const actual = getZonedParts(new Date(utc), timeZone);
		const actualAsUTC = dateFromPartsAsUTC(actual);
		const next = utc + target - actualAsUTC;

		if (next === utc) {
			break;
		}

		utc = next;
	}

	return new Date(utc);
};

export const safeComponentParts = (components: TempoComponents): Required<Omit<TempoComponents, 'timeZone'>> => ({
	day: components.day ?? 1,
	hour: components.hour ?? 0,
	millisecond: components.millisecond ?? 0,
	minute: components.minute ?? 0,
	month: components.month ?? 1,
	second: components.second ?? 0,
	year: components.year,
});

export const assertSafeZonedComponents = (components: TempoComponents, date: Date, timeZone: string): void => {
	const expected = safeComponentParts(components);
	const actual = getZonedParts(date, timeZone);

	if (
		actual.year !== expected.year ||
		actual.month !== expected.month ||
		actual.day !== expected.day ||
		actual.hour !== expected.hour ||
		actual.minute !== expected.minute ||
		actual.second !== expected.second ||
		actual.millisecond !== expected.millisecond
	) {
		throw new RangeError('Invalid Tempo local date/time components');
	}
};

export const daysInMonth = (year: number, month: number): number => new Date(Date.UTC(year, month, 0)).getUTCDate();

export const isoWeekData = (parts: Pick<ZonedParts, 'year' | 'month' | 'day'>): { readonly year: number; readonly week: number } => {
	const date = new Date(Date.UTC(parts.year, parts.month - 1, parts.day));
	const day = date.getUTCDay() || 7;

	date.setUTCDate(date.getUTCDate() + 4 - day);

	const year = date.getUTCFullYear();
	const yearStart = new Date(Date.UTC(year, 0, 1));

	const week = Math.ceil(((date.getTime() - yearStart.getTime()) / millisecondsPerDay + 1) / 7);

	return { week, year };
};

export const weeksInISOYear = (year: number): number => isoWeekData({ day: 28, month: 12, year }).week;

export const resolveWeekday = (weekday: WeekdayInput): number => {
	if (typeof weekday === 'number') {
		return ((weekday % 7) + 7) % 7;
	}

	switch (weekday.toLowerCase()) {
		case 'sunday':

		case 'sun':
			return 0;

		case 'monday':

		case 'mon':
			return 1;

		case 'tuesday':

		case 'tue':
			return 2;

		case 'wednesday':

		case 'wed':
			return 3;

		case 'thursday':

		case 'thu':
			return 4;

		case 'friday':

		case 'fri':
			return 5;

		case 'saturday':

		case 'sat':
			return 6;
	}

	throw new RangeError(`Invalid Tempo weekday: ${String(weekday)}`);
};

export const normalizeUnit = (unit: TimeUnit): Exclude<TimeUnit, `${string}s`> => {
	switch (unit) {
		case 'milliseconds':
			return 'millisecond';

		case 'seconds':
			return 'second';

		case 'minutes':
			return 'minute';

		case 'hours':
			return 'hour';

		case 'days':
			return 'day';

		case 'weeks':
			return 'week';

		case 'months':
			return 'month';

		case 'quarters':
			return 'quarter';

		case 'years':
			return 'year';

		case 'decades':
			return 'decade';

		case 'centuries':
			return 'century';

		case 'millenniums':

		case 'millennia':
			return 'millennium';

		default:
			return unit;
	}
};

export const fixedUnitMilliseconds = (unit: TimeUnit): number | null => {
	switch (normalizeUnit(unit)) {
		case 'millisecond':
			return 1;

		case 'second':
			return millisecondsPerSecond;

		case 'minute':
			return millisecondsPerMinute;

		case 'hour':
			return millisecondsPerHour;

		case 'day':
			return millisecondsPerDay;

		case 'week':
			return millisecondsPerWeek;

		default:
			return null;
	}
};

export const monthNames = (locale: string, width: 'short' | 'long'): readonly string[] => {
	const key = `${locale}|month|${width}`;
	const cached = monthNameCache.get(key);

	if (cached !== undefined) {
		return cached;
	}

	const names = Array.from({ length: 12 }, (_, index) =>
		new Intl.DateTimeFormat(locale, {
			month: width,
			timeZone: defaultTimeZone,
		}).format(new Date(Date.UTC(2024, index, 1))),
	);

	monthNameCache.set(key, names);

	return names;
};

export const weekdayNames = (locale: string, width: 'short' | 'long'): readonly string[] => {
	const key = `${locale}|weekday|${width}`;
	const cached = weekdayNameCache.get(key);

	if (cached !== undefined) {
		return cached;
	}

	const names = Array.from({ length: 7 }, (_, index) =>
		new Intl.DateTimeFormat(locale, {
			timeZone: defaultTimeZone,
			weekday: width,
		}).format(new Date(Date.UTC(2024, 0, 7 + index))),
	);

	weekdayNameCache.set(key, names);

	return names;
};

export const monthNumberFromName = (input: string, locale = 'en-US'): number | null => {
	const normalized = input.toLowerCase();

	const names = [...monthNames(locale, 'long'), ...monthNames(locale, 'short')].map((name, index) => ({
		name: name.toLowerCase(),
		month: (index % 12) + 1,
	}));

	const match = names.find((item) => item.name === normalized);

	return match?.month ?? null;
};

export class TempoCalendar {
	normalizeTimeZone(timeZone: string | undefined): string {
		return normalizeTimeZone(timeZone);
	}

	getZonedParts(date: Date, timeZone: string): ZonedParts {
		return getZonedParts(date, timeZone);
	}

	dateFromZonedComponents(components: TempoComponents, fallbackTimeZone = defaultTimeZone): Date {
		return dateFromZonedComponents(components, fallbackTimeZone);
	}
}

export class TempoArithmetic {
	fixedUnitMilliseconds(unit: TimeUnit): number | null {
		return fixedUnitMilliseconds(unit);
	}

	normalizeUnit(unit: TimeUnit): Exclude<TimeUnit, `${string}s`> {
		return normalizeUnit(unit);
	}
}

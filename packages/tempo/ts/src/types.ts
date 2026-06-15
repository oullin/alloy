import type { TempoDuration } from './duration';
import type { Tempo, TempoImmutable, TempoMutable } from './core';
import type { TempoRuntime } from './runtime';

export type TempoInput = Date | Tempo | TempoImmutable | TempoMutable | number | string;

export type TimeUnit =
	| 'millisecond'
	| 'milliseconds'
	| 'second'
	| 'seconds'
	| 'minute'
	| 'minutes'
	| 'hour'
	| 'hours'
	| 'day'
	| 'days'
	| 'week'
	| 'weeks'
	| 'month'
	| 'months'
	| 'quarter'
	| 'quarters'
	| 'year'
	| 'years'
	| 'decade'
	| 'decades'
	| 'century'
	| 'centuries'
	| 'millennium'
	| 'millenniums'
	| 'millennia';

export type BoundaryUnit = 'millisecond' | 'second' | 'minute' | 'hour' | 'day' | 'week' | 'month' | 'quarter' | 'year' | 'decade' | 'century' | 'millennium';

export type ComparisonUnit = 'millisecond' | BoundaryUnit;

export type TimeZoneNameStyle = NonNullable<Intl.DateTimeFormatOptions['timeZoneName']>;

export type WeekdayInput = 0 | 1 | 2 | 3 | 4 | 5 | 6 | 7 | 'sunday' | 'monday' | 'tuesday' | 'wednesday' | 'thursday' | 'friday' | 'saturday' | 'sun' | 'mon' | 'tue' | 'wed' | 'thu' | 'fri' | 'sat';

export type TempoOptions = {
	readonly fallbackLocale?: string;
	readonly locale?: string;
	readonly runtime?: TempoRuntime;
	readonly timeZone?: string;
	readonly translator?: TempoTranslator;
};

export type TempoSettings = {
	readonly fallbackLocale?: string;
	readonly humanDiffOptions?: HumanDiffOptions;
	readonly locale?: string;
	readonly midDayAt?: number;
	readonly monthsOverflow?: boolean;
	readonly strictMode?: boolean;
	readonly testNow?: TempoInput | null;
	readonly timeZone?: string;
	readonly weekendDays?: readonly number[];
	readonly yearsOverflow?: boolean;
};

export type TempoSerializer = (tempo: TempoImmutable) => string;

export type TempoTranslationValue = string | number | null;

export type TempoTranslator = {
	readonly fallbackLocale?: string;
	readonly locale?: string;
	translate?: (key: string, replacements?: Record<string, string>) => TempoTranslationValue;
	getMessage?: (key: string) => TempoTranslationValue;
};

export type TempoComponents = {
	readonly year: number;
	readonly month?: number;
	readonly day?: number;
	readonly hour?: number;
	readonly minute?: number;
	readonly second?: number;
	readonly millisecond?: number;
	readonly timeZone?: string;
};

export type TempoSettableComponents = Partial<Omit<TempoComponents, 'timeZone'>> & {
	readonly timeZone?: string;
};

export type TempoObject = Required<Omit<TempoComponents, 'timeZone'>> & {
	readonly offsetMinutes: number;
	readonly timeZone: string;
	readonly weekday: number;
};

export type DurationLike = {
	readonly years?: number;
	readonly quarters?: number;
	readonly months?: number;
	readonly weeks?: number;
	readonly days?: number;
	readonly hours?: number;
	readonly minutes?: number;
	readonly seconds?: number;
	readonly milliseconds?: number;
};

export type DurationInput = DurationLike | TempoDuration | string;

export type DurationObject = Required<DurationLike>;

export type DiffOptions = {
	readonly absolute?: boolean;
	readonly float?: boolean;
};

export type FormatOptions = {
	readonly locale?: string;
	readonly timeZone?: string;
};

export type CalendarFormatKey = 'sameDay' | 'nextDay' | 'nextWeek' | 'lastDay' | 'lastWeek' | 'sameElse';

export type CalendarFormats = Partial<Record<CalendarFormatKey, string>>;

export type TimeStringPrecision = 'second' | 'millisecond';

export type HumanDiffOptions = {
	readonly absolute?: boolean;
	readonly locale?: string;
	readonly numeric?: Intl.RelativeTimeFormatNumeric;
	readonly style?: Intl.RelativeTimeFormatStyle;
	readonly unit?: 'second' | 'minute' | 'hour' | 'day' | 'week' | 'month' | 'year';
};

export type StartOfWeekOptions = {
	readonly weekStartsOn?: number;
};

export type PeriodOptions = {
	readonly step?: DurationInput;
	readonly includeEnd?: boolean;
};

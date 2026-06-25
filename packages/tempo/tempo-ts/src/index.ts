import { Tempo } from '#core';
import { TempoDuration } from '#duration';
import { TempoFactory } from '#factory';

export type {
	BoundaryUnit,
	CalendarFormatKey,
	CalendarFormats,
	ComparisonUnit,
	DiffOptions,
	DurationInput,
	DurationLike,
	DurationObject,
	FormatOptions,
	HumanDiffOptions,
	PeriodOptions,
	StartOfWeekOptions,
	TempoComponents,
	TempoInput,
	TempoObject,
	TempoOptions,
	TempoPolicy,
	TempoSerializer,
	TempoSettableComponents,
	TempoSettings,
	TempoTranslationValue,
	TempoTranslator,
	TimeStringPrecision,
	TimeUnit,
	TimeZoneNameStyle,
	WeekdayInput,
} from '#types';
export { createTempoRuntime, TempoRuntime } from '#runtime';
export { TempoDuration } from '#duration';
export { Tempo, TempoImmutable } from '#core';
export { TempoInterval, TempoPeriod } from '#ranges';
export { TempoFactory } from '#factory';

export const now: typeof Tempo.now = (options) => Tempo.now(options);

export const today: typeof Tempo.today = (options) => Tempo.today(options);

export const tomorrow: typeof Tempo.tomorrow = (options) => Tempo.tomorrow(options);

export const yesterday: typeof Tempo.yesterday = (options) => Tempo.yesterday(options);

export const parse: typeof Tempo.parse = (input, options) => Tempo.parse(input, options);

export const fromJSON: typeof Tempo.fromJSON = (input, options) => Tempo.fromJSON(input, options);

export const min: typeof Tempo.min = (...items) => Tempo.min(...items);

export const max: typeof Tempo.max = (...items) => Tempo.max(...items);

export const average: typeof Tempo.average = (start, end) => Tempo.average(start, end);

export const fromFormat: typeof Tempo.fromFormat = (input, pattern, options) => Tempo.fromFormat(input, pattern, options);

export const create: typeof Tempo.create = (components, options) => Tempo.create(components, options);

export const createNormalized: typeof Tempo.createNormalized = (components, options) => Tempo.createNormalized(components, options);

export const fromDate: typeof Tempo.fromDate = (year, month, day, options) => Tempo.fromDate(year, month, day, options);

export const fromTime: typeof Tempo.fromTime = (hour, minute, second, millisecond, options) => Tempo.fromTime(hour, minute, second, millisecond, options);

export const fromTimeString: typeof Tempo.fromTimeString = (time, options) => Tempo.fromTimeString(time, options);

export const fromObject: typeof Tempo.fromObject = (components) => Tempo.fromObject(components);

export const fromTimestamp: typeof Tempo.fromTimestamp = (timestamp, options) => Tempo.fromTimestamp(timestamp, options);

export const fromTimestampMs: typeof Tempo.fromTimestampMs = (timestamp, options) => Tempo.fromTimestampMs(timestamp, options);

export const createDuration: typeof TempoDuration.from = (input) => TempoDuration.from(input);

export const parseDuration: typeof TempoDuration.parse = (input) => TempoDuration.parse(input);

export const createFactory: typeof TempoFactory.create = (options) => TempoFactory.create(options);

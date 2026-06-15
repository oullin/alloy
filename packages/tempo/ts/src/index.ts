import { Tempo } from './core';
import { TempoDuration } from './duration';
import { TempoFactory } from './factory';

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
	TempoSerializer,
	TempoSettableComponents,
	TempoSettings,
	TempoTranslationValue,
	TempoTranslator,
	TimeStringPrecision,
	TimeUnit,
	TimeZoneNameStyle,
	WeekdayInput,
} from './types';
export { createTempoRuntime, TempoRuntime } from './runtime';
export { TempoDuration } from './duration';
export { Tempo, TempoImmutable, TempoMutable } from './core';
export { TempoInterval, TempoPeriod } from './ranges';
export { TempoFactory } from './factory';

export const now = Tempo.now;

export const today = Tempo.today;

export const tomorrow = Tempo.tomorrow;

export const yesterday = Tempo.yesterday;

export const parse = Tempo.parse;

export const tryParse = Tempo.tryParse;

export const canParse = Tempo.canParse;

export const fromJSON = Tempo.fromJSON;

export const min = Tempo.min;

export const max = Tempo.max;

export const minimum = Tempo.minimum;

export const maximum = Tempo.maximum;

export const average = Tempo.average;

export const fromFormat = Tempo.fromFormat;

export const tryFromFormat = Tempo.tryFromFormat;

export const hasFormat = Tempo.hasFormat;

export const create = Tempo.create;

export const createSafe = Tempo.createSafe;

export const createFromDate = Tempo.createFromDate;

export const createFromTime = Tempo.createFromTime;

export const createMidnightDate = Tempo.createMidnightDate;

export const fromObject = Tempo.fromObject;

export const fromTimestamp = Tempo.fromTimestamp;

export const fromTimestampMs = Tempo.fromTimestampMs;

export const createDuration = TempoDuration.from;

export const parseDuration = TempoDuration.parse;

export const createFactory = TempoFactory.create;

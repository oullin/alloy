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
} from './types';
export { createTempoRuntime, TempoRuntime } from './runtime';
export { TempoDuration } from './duration';
export { Tempo, TempoImmutable } from './core';
export { TempoInterval, TempoPeriod } from './ranges';
export { TempoFactory } from './factory';

export const now = Tempo.now;

export const today = Tempo.today;

export const tomorrow = Tempo.tomorrow;

export const yesterday = Tempo.yesterday;

export const parse = Tempo.parse;

export const fromJSON = Tempo.fromJSON;

export const min = Tempo.min;

export const max = Tempo.max;

export const average = Tempo.average;

export const fromFormat = Tempo.fromFormat;

export const create = Tempo.create;

export const createNormalized = Tempo.createNormalized;

export const fromDate = Tempo.fromDate;

export const fromTime = Tempo.fromTime;

export const fromTimeString = Tempo.fromTimeString;

export const fromObject = Tempo.fromObject;

export const fromTimestamp = Tempo.fromTimestamp;

export const fromTimestampMs = Tempo.fromTimestampMs;

export const createDuration = TempoDuration.from;

export const parseDuration = TempoDuration.parse;

export const createFactory = TempoFactory.create;

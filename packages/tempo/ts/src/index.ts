export type TempoInput =
  | Date
  | Tempo
  | TempoImmutable
  | TempoMutable
  | number
  | string;

export type TimeUnit =
  | "millisecond"
  | "milliseconds"
  | "second"
  | "seconds"
  | "minute"
  | "minutes"
  | "hour"
  | "hours"
  | "day"
  | "days"
  | "week"
  | "weeks"
  | "month"
  | "months"
  | "quarter"
  | "quarters"
  | "year"
  | "years";

export type BoundaryUnit =
  | "second"
  | "minute"
  | "hour"
  | "day"
  | "week"
  | "month"
  | "quarter"
  | "year";

export type ComparisonUnit = "millisecond" | BoundaryUnit;

export type TimeZoneNameStyle = NonNullable<
  Intl.DateTimeFormatOptions["timeZoneName"]
>;

export type WeekdayInput =
  | 0
  | 1
  | 2
  | 3
  | 4
  | 5
  | 6
  | "sunday"
  | "monday"
  | "tuesday"
  | "wednesday"
  | "thursday"
  | "friday"
  | "saturday"
  | "sun"
  | "mon"
  | "tue"
  | "wed"
  | "thu"
  | "fri"
  | "sat";

export type TempoOptions = {
  readonly timeZone?: string;
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

export type TempoSettableComponents = Partial<
  Omit<TempoComponents, "timeZone">
> & {
  readonly timeZone?: string;
};

export type TempoObject = Required<Omit<TempoComponents, "timeZone">> & {
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

export type TimeStringPrecision = "second" | "millisecond";

export type HumanDiffOptions = {
  readonly absolute?: boolean;
  readonly locale?: string;
  readonly numeric?: Intl.RelativeTimeFormatNumeric;
  readonly style?: Intl.RelativeTimeFormatStyle;
  readonly unit?:
    | "second"
    | "minute"
    | "hour"
    | "day"
    | "week"
    | "month"
    | "year";
};

export type StartOfWeekOptions = {
  readonly weekStartsOn?: number;
};

export type PeriodOptions = {
  readonly step?: DurationInput;
  readonly includeEnd?: boolean;
};

type ZonedParts = {
  readonly year: number;
  readonly month: number;
  readonly day: number;
  readonly hour: number;
  readonly minute: number;
  readonly second: number;
  readonly millisecond: number;
  readonly weekday: number;
};

type DateTimePart = Intl.DateTimeFormatPart;

const defaultTimeZone = "UTC";
const millisecondsPerSecond = 1000;
const millisecondsPerMinute = 60 * millisecondsPerSecond;
const millisecondsPerHour = 60 * millisecondsPerMinute;
const millisecondsPerDay = 24 * millisecondsPerHour;
const millisecondsPerWeek = 7 * millisecondsPerDay;
const isoDatePattern = /^(\d{4})-(\d{2})-(\d{2})$/;
const isoLocalPattern =
  /^(\d{4})-(\d{2})-(\d{2})(?:[T\s](\d{2})(?::?(\d{2}))?(?::?(\d{2})(?:\.(\d{1,9}))?)?)?$/;
const isoDurationPattern =
  /^(-)?P(?:(\d+(?:\.\d+)?)Y)?(?:(\d+(?:\.\d+)?)M)?(?:(\d+(?:\.\d+)?)W)?(?:(\d+(?:\.\d+)?)D)?(?:T(?:(\d+(?:\.\d+)?)H)?(?:(\d+(?:\.\d+)?)M)?(?:(\d+(?:\.\d+)?)S)?)?$/i;
const timezonePattern = /(?:Z|[+-]\d{2}:?\d{2})$/i;
const formatterCache = new Map<string, Intl.DateTimeFormat>();
const monthNameCache = new Map<string, readonly string[]>();
const weekdayNameCache = new Map<string, readonly string[]>();

const pad = (value: number, length = 2): string =>
  String(Math.trunc(Math.abs(value))).padStart(length, "0");

const assertFiniteNumber = (value: number, label: string): void => {
  if (!Number.isFinite(value)) {
    throw new RangeError(`${label} must be a finite number`);
  }
};

const normalizeTimeZone = (timeZone: string | undefined): string => {
  const zone = timeZone ?? defaultTimeZone;

  try {
    Intl.DateTimeFormat("en-US", { timeZone: zone }).format(new Date(0));
  } catch (error) {
    throw new RangeError(`Invalid Tempo time zone: ${zone}`, {
      cause: error,
    });
  }

  return zone;
};

const getFormatter = (
  timeZone: string,
  timeZoneName?: TimeZoneNameStyle,
  locale = "en-US-u-nu-latn",
): Intl.DateTimeFormat => {
  const key = `${locale}|${timeZone}|${timeZoneName ?? "parts"}`;
  const cached = formatterCache.get(key);

  if (cached !== undefined) {
    return cached;
  }

  const formatter = new Intl.DateTimeFormat(locale, {
    calendar: "gregory",
    day: "2-digit",
    fractionalSecondDigits: 3,
    hour: "2-digit",
    hourCycle: "h23",
    minute: "2-digit",
    month: "2-digit",
    second: "2-digit",
    timeZone,
    timeZoneName,
    weekday: "short",
    year: "numeric",
  });

  formatterCache.set(key, formatter);
  return formatter;
};

const readPart = (parts: readonly DateTimePart[], type: string): string => {
  const part = parts.find((item) => item.type === type);
  return part?.value ?? "";
};

const weekdayIndex = (weekday: string): number => {
  switch (weekday) {
    case "Sun":
      return 0;
    case "Mon":
      return 1;
    case "Tue":
      return 2;
    case "Wed":
      return 3;
    case "Thu":
      return 4;
    case "Fri":
      return 5;
    case "Sat":
      return 6;
    default:
      return 0;
  }
};

const getZonedParts = (date: Date, timeZone: string): ZonedParts => {
  const parts = getFormatter(timeZone).formatToParts(date);

  return {
    year: Number(readPart(parts, "year")),
    month: Number(readPart(parts, "month")),
    day: Number(readPart(parts, "day")),
    hour: Number(readPart(parts, "hour")),
    minute: Number(readPart(parts, "minute")),
    second: Number(readPart(parts, "second")),
    millisecond: Number(readPart(parts, "fractionalSecond") || "0"),
    weekday: weekdayIndex(readPart(parts, "weekday")),
  };
};

const dateFromPartsAsUTC = (parts: TempoComponents): number =>
  Date.UTC(
    parts.year,
    (parts.month ?? 1) - 1,
    parts.day ?? 1,
    parts.hour ?? 0,
    parts.minute ?? 0,
    parts.second ?? 0,
    parts.millisecond ?? 0,
  );

const dateFromZonedComponents = (
  components: TempoComponents,
  fallbackTimeZone = defaultTimeZone,
): Date => {
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

const fromNumericTimestamp = (input: number): Date => {
  assertFiniteNumber(input, "Timestamp");
  const magnitude = Math.abs(input);
  const milliseconds = magnitude < 10_000_000_000 ? input * 1000 : input;

  return new Date(milliseconds);
};

const parseLocalText = (input: string, timeZone: string): Date | null => {
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
      millisecond: Number((local[7] ?? "0").slice(0, 3).padEnd(3, "0")),
      minute: Number(local[5] ?? 0),
      month: Number(local[2]),
      second: Number(local[6] ?? 0),
      timeZone,
      year: Number(local[1]),
    },
    timeZone,
  );
};

const zoneFromInput = (
  input: TempoInput,
  options: TempoOptions | undefined,
): string => {
  if (
    input instanceof Tempo ||
    input instanceof TempoImmutable ||
    input instanceof TempoMutable
  ) {
    return normalizeTimeZone(options?.timeZone ?? input.timeZone);
  }

  return normalizeTimeZone(options?.timeZone);
};

const asDate = (input: TempoInput, options?: TempoOptions): Date => {
  if (
    input instanceof Tempo ||
    input instanceof TempoImmutable ||
    input instanceof TempoMutable
  ) {
    return input.toDate();
  }

  if (input instanceof Date) {
    return new Date(input.getTime());
  }

  const timeZone = normalizeTimeZone(options?.timeZone);
  const date =
    typeof input === "number"
      ? fromNumericTimestamp(input)
      : (parseLocalText(input, timeZone) ?? new Date(input));

  if (Number.isNaN(date.getTime())) {
    throw new RangeError(`Invalid Tempo input: ${String(input)}`);
  }

  return date;
};

const daysInMonth = (year: number, month: number): number =>
  new Date(Date.UTC(year, month, 0)).getUTCDate();

const isoWeekData = (
  parts: Pick<ZonedParts, "year" | "month" | "day">,
): { readonly year: number; readonly week: number } => {
  const date = new Date(Date.UTC(parts.year, parts.month - 1, parts.day));
  const day = date.getUTCDay() || 7;
  date.setUTCDate(date.getUTCDate() + 4 - day);

  const year = date.getUTCFullYear();
  const yearStart = new Date(Date.UTC(year, 0, 1));
  const week = Math.ceil(
    ((date.getTime() - yearStart.getTime()) / millisecondsPerDay + 1) / 7,
  );

  return { week, year };
};

const weeksInISOYear = (year: number): number =>
  isoWeekData({ day: 28, month: 12, year }).week;

const resolveWeekday = (weekday: WeekdayInput): number => {
  if (typeof weekday === "number") {
    return ((weekday % 7) + 7) % 7;
  }

  switch (weekday.toLowerCase()) {
    case "sunday":
    case "sun":
      return 0;
    case "monday":
    case "mon":
      return 1;
    case "tuesday":
    case "tue":
      return 2;
    case "wednesday":
    case "wed":
      return 3;
    case "thursday":
    case "thu":
      return 4;
    case "friday":
    case "fri":
      return 5;
    case "saturday":
    case "sat":
      return 6;
  }

  throw new RangeError(`Invalid Tempo weekday: ${String(weekday)}`);
};

const normalizeUnit = (unit: TimeUnit): Exclude<TimeUnit, `${string}s`> => {
  switch (unit) {
    case "milliseconds":
      return "millisecond";
    case "seconds":
      return "second";
    case "minutes":
      return "minute";
    case "hours":
      return "hour";
    case "days":
      return "day";
    case "weeks":
      return "week";
    case "months":
      return "month";
    case "quarters":
      return "quarter";
    case "years":
      return "year";
    default:
      return unit;
  }
};

const fixedUnitMilliseconds = (unit: TimeUnit): number | null => {
  switch (normalizeUnit(unit)) {
    case "millisecond":
      return 1;
    case "second":
      return millisecondsPerSecond;
    case "minute":
      return millisecondsPerMinute;
    case "hour":
      return millisecondsPerHour;
    case "day":
      return millisecondsPerDay;
    case "week":
      return millisecondsPerWeek;
    default:
      return null;
  }
};

const monthNames = (
  locale: string,
  width: "short" | "long",
): readonly string[] => {
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

const weekdayNames = (
  locale: string,
  width: "short" | "long",
): readonly string[] => {
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

const monthNumberFromName = (
  input: string,
  locale = "en-US",
): number | null => {
  const normalized = input.toLowerCase();
  const names = [
    ...monthNames(locale, "long"),
    ...monthNames(locale, "short"),
  ].map((name, index) => ({
    name: name.toLowerCase(),
    month: (index % 12) + 1,
  }));
  const match = names.find((item) => item.name === normalized);

  return match?.month ?? null;
};

const timeZoneName = (
  date: Date,
  timeZone: string,
  style: TimeZoneNameStyle,
  locale: string,
): string => {
  const parts = getFormatter(timeZone, style, locale).formatToParts(date);

  return readPart(parts, "timeZoneName") || timeZone;
};

const formatOffset = (offsetMinutes: number, separator: ":" | ""): string => {
  const sign = offsetMinutes >= 0 ? "+" : "-";
  const absolute = Math.abs(offsetMinutes);
  const hours = Math.trunc(absolute / 60);
  const minutes = absolute % 60;

  return `${sign}${pad(hours)}${separator}${pad(minutes)}`;
};

const ordinal = (value: number): string => {
  const remainder = value % 100;

  if (remainder >= 11 && remainder <= 13) {
    return `${value}th`;
  }

  switch (value % 10) {
    case 1:
      return `${value}st`;
    case 2:
      return `${value}nd`;
    case 3:
      return `${value}rd`;
    default:
      return `${value}th`;
  }
};

const escapePattern = (input: string): string =>
  input.replace(/[.*+?^${}()|[\]\\]/g, "\\$&");

const parseOffsetMinutes = (input: string): number => {
  if (input === "Z") {
    return 0;
  }

  const match = /^([+-])(\d{2}):?(\d{2})$/.exec(input);
  if (match === null) {
    throw new RangeError(`Invalid Tempo offset: ${input}`);
  }

  const minutes = Number(match[2]) * 60 + Number(match[3]);
  return match[1] === "-" ? -minutes : minutes;
};

const parseFromPattern = (
  input: string,
  pattern: string,
  options?: TempoOptions,
): Date => {
  const tokens = [
    "YYYY",
    "MMMM",
    "dddd",
    "MMM",
    "ddd",
    "SSS",
    "Do",
    "YY",
    "ZZ",
    "MM",
    "DD",
    "HH",
    "hh",
    "mm",
    "ss",
    "Z",
    "M",
    "D",
    "H",
    "h",
    "m",
    "s",
    "A",
    "a",
  ] as const;
  const groups: string[] = [];
  let expression = "^";

  for (let index = 0; index < pattern.length; ) {
    if (pattern[index] === "[") {
      const end = pattern.indexOf("]", index);

      if (end >= 0) {
        expression += escapePattern(pattern.slice(index + 1, end));
        index = end + 1;
        continue;
      }
    }

    const token = tokens.find((item) => pattern.slice(index).startsWith(item));

    if (token === undefined) {
      expression += escapePattern(pattern[index] ?? "");
      index += 1;
      continue;
    }

    groups.push(token);
    expression +=
      token === "A" || token === "a"
        ? "(AM|PM|am|pm)"
        : token === "MMM" || token === "MMMM"
          ? "([\\p{L}.]+)"
          : token === "ddd" || token === "dddd"
            ? "([\\p{L}.]+)"
            : token === "Do"
              ? "(\\d{1,2})(?:st|nd|rd|th)"
              : token === "Z"
                ? "(Z|[+-]\\d{2}:\\d{2})"
                : token === "ZZ"
                  ? "(Z|[+-]\\d{4})"
                  : "(\\d{1," +
                    (token.length === 1 ? "4" : String(token.length)) +
                    "})";
    index += token.length;
  }

  expression += "$";
  const match = new RegExp(expression, "u").exec(input);

  if (match === null) {
    throw new RangeError(`Input does not match Tempo format: ${input}`);
  }

  const values = new Map<string, string>();
  groups.forEach((token, index) => values.set(token, match[index + 1] ?? ""));

  let year = values.has("YYYY")
    ? Number(values.get("YYYY"))
    : values.has("YY")
      ? 2000 + Number(values.get("YY"))
      : 1970;
  const meridiem = values.get("A") ?? values.get("a");
  let hour = Number(
    values.get("HH") ??
      values.get("H") ??
      values.get("hh") ??
      values.get("h") ??
      0,
  );

  if (meridiem !== undefined) {
    const lower = meridiem.toLowerCase();
    if (lower === "pm" && hour < 12) {
      hour += 12;
    }
    if (lower === "am" && hour === 12) {
      hour = 0;
    }
  }

  if (!Number.isFinite(year)) {
    year = 1970;
  }

  const components: TempoComponents = {
    day: Number(values.get("DD") ?? values.get("Do") ?? values.get("D") ?? 1),
    hour,
    millisecond: Number((values.get("SSS") ?? "0").slice(0, 3).padEnd(3, "0")),
    minute: Number(values.get("mm") ?? values.get("m") ?? 0),
    month:
      monthNumberFromName(values.get("MMMM") ?? values.get("MMM") ?? "") ??
      Number(values.get("MM") ?? values.get("M") ?? 1),
    second: Number(values.get("ss") ?? values.get("s") ?? 0),
    timeZone: options?.timeZone,
    year,
  };
  const offset = values.get("Z") ?? values.get("ZZ");

  if (offset !== undefined && offset !== "") {
    const offsetMinutes = parseOffsetMinutes(offset);
    return new Date(
      dateFromPartsAsUTC(components) - offsetMinutes * millisecondsPerMinute,
    );
  }

  return dateFromZonedComponents(components, options?.timeZone);
};

const bestRelativeUnit = (
  milliseconds: number,
): NonNullable<HumanDiffOptions["unit"]> => {
  const absolute = Math.abs(milliseconds);

  if (absolute < millisecondsPerMinute) {
    return "second";
  }
  if (absolute < millisecondsPerHour) {
    return "minute";
  }
  if (absolute < millisecondsPerDay) {
    return "hour";
  }
  if (absolute < millisecondsPerWeek) {
    return "day";
  }
  if (absolute < millisecondsPerDay * 30) {
    return "week";
  }
  if (absolute < millisecondsPerDay * 365) {
    return "month";
  }

  return "year";
};

const unitDivisor = (unit: NonNullable<HumanDiffOptions["unit"]>): number => {
  switch (unit) {
    case "second":
      return millisecondsPerSecond;
    case "minute":
      return millisecondsPerMinute;
    case "hour":
      return millisecondsPerHour;
    case "day":
      return millisecondsPerDay;
    case "week":
      return millisecondsPerWeek;
    case "month":
      return millisecondsPerDay * 30;
    case "year":
      return millisecondsPerDay * 365;
  }
};

const durationFromInput = (input: DurationInput): TempoDuration =>
  input instanceof TempoDuration
    ? input
    : typeof input === "string"
      ? TempoDuration.parse(input)
      : new TempoDuration(input);

export class TempoDuration {
  readonly years: number;
  readonly quarters: number;
  readonly months: number;
  readonly weeks: number;
  readonly days: number;
  readonly hours: number;
  readonly minutes: number;
  readonly seconds: number;
  readonly milliseconds: number;

  constructor(input: DurationLike = {}) {
    this.years = input.years ?? 0;
    this.quarters = input.quarters ?? 0;
    this.months = input.months ?? 0;
    this.weeks = input.weeks ?? 0;
    this.days = input.days ?? 0;
    this.hours = input.hours ?? 0;
    this.minutes = input.minutes ?? 0;
    this.seconds = input.seconds ?? 0;
    this.milliseconds = input.milliseconds ?? 0;

    for (const [key, value] of Object.entries(this.toObject())) {
      assertFiniteNumber(value, `Duration ${key}`);
    }
  }

  static from(input: DurationInput): TempoDuration {
    return durationFromInput(input);
  }

  static parse(input: string): TempoDuration {
    const match = isoDurationPattern.exec(input);

    if (match === null) {
      throw new RangeError(`Invalid Tempo duration: ${input}`);
    }

    const sign = match[1] === "-" ? -1 : 1;
    const seconds = Number(match[8] ?? 0);
    const wholeSeconds = Math.trunc(seconds);

    return new TempoDuration({
      days: sign * Number(match[5] ?? 0),
      hours: sign * Number(match[6] ?? 0),
      milliseconds: sign * Math.round((seconds - wholeSeconds) * 1000),
      minutes: sign * Number(match[7] ?? 0),
      months: sign * Number(match[3] ?? 0),
      seconds: sign * wholeSeconds,
      weeks: sign * Number(match[4] ?? 0),
      years: sign * Number(match[2] ?? 0),
    }).normalized();
  }

  plus(input: DurationInput): TempoDuration {
    const other = durationFromInput(input);

    return new TempoDuration({
      days: this.days + other.days,
      hours: this.hours + other.hours,
      milliseconds: this.milliseconds + other.milliseconds,
      minutes: this.minutes + other.minutes,
      months: this.months + other.months,
      quarters: this.quarters + other.quarters,
      seconds: this.seconds + other.seconds,
      weeks: this.weeks + other.weeks,
      years: this.years + other.years,
    });
  }

  minus(input: DurationInput): TempoDuration {
    return this.plus(durationFromInput(input).negated());
  }

  negated(): TempoDuration {
    return new TempoDuration({
      days: -this.days,
      hours: -this.hours,
      milliseconds: -this.milliseconds,
      minutes: -this.minutes,
      months: -this.months,
      quarters: -this.quarters,
      seconds: -this.seconds,
      weeks: -this.weeks,
      years: -this.years,
    });
  }

  abs(): TempoDuration {
    return new TempoDuration({
      days: Math.abs(this.days),
      hours: Math.abs(this.hours),
      milliseconds: Math.abs(this.milliseconds),
      minutes: Math.abs(this.minutes),
      months: Math.abs(this.months),
      quarters: Math.abs(this.quarters),
      seconds: Math.abs(this.seconds),
      weeks: Math.abs(this.weeks),
      years: Math.abs(this.years),
    });
  }

  normalized(): TempoDuration {
    const sign = this.direction();
    const absolute = this.abs();
    let milliseconds = absolute.milliseconds;
    let seconds = absolute.seconds + Math.trunc(milliseconds / 1000);
    milliseconds %= 1000;
    let minutes = absolute.minutes + Math.trunc(seconds / 60);
    seconds %= 60;
    let hours = absolute.hours + Math.trunc(minutes / 60);
    minutes %= 60;
    let days = absolute.days + Math.trunc(hours / 24);
    hours %= 24;
    const months = absolute.months + absolute.quarters * 3;
    const years = absolute.years + Math.trunc(months / 12);

    days += absolute.weeks * 7;

    return new TempoDuration({
      days: sign * days,
      hours: sign * hours,
      milliseconds: sign * milliseconds,
      minutes: sign * minutes,
      months: sign * (months % 12),
      seconds: sign * seconds,
      years: sign * years,
    });
  }

  total(unit: TimeUnit): number {
    const milliseconds = this.totalMilliseconds();
    const fixed = fixedUnitMilliseconds(unit);

    if (fixed !== null) {
      return milliseconds / fixed;
    }

    switch (normalizeUnit(unit)) {
      case "month":
        return (
          this.years * 12 +
          this.quarters * 3 +
          this.months +
          milliseconds / (millisecondsPerDay * 30)
        );
      case "quarter":
        return this.total("month") / 3;
      case "year":
        return this.total("month") / 12;
      default:
        return milliseconds;
    }
  }

  isZero(): boolean {
    return Object.values(this.toObject()).every((value) => value === 0);
  }

  toObject(): DurationObject {
    return {
      days: this.days,
      hours: this.hours,
      milliseconds: this.milliseconds,
      minutes: this.minutes,
      months: this.months,
      quarters: this.quarters,
      seconds: this.seconds,
      weeks: this.weeks,
      years: this.years,
    };
  }

  toISOString(): string {
    if (this.isZero()) {
      return "PT0S";
    }

    const normalized = this.normalized();
    const sign = normalized.direction() < 0 ? "-" : "";
    const value = normalized.abs();
    const dateParts = [
      value.years === 0 ? "" : `${value.years}Y`,
      value.months === 0 ? "" : `${value.months}M`,
      value.days === 0 ? "" : `${value.days}D`,
    ].join("");
    const secondValue =
      value.milliseconds === 0
        ? String(value.seconds)
        : `${value.seconds}.${pad(value.milliseconds, 3)}`;
    const timeParts = [
      value.hours === 0 ? "" : `${value.hours}H`,
      value.minutes === 0 ? "" : `${value.minutes}M`,
      value.seconds === 0 && value.milliseconds === 0 ? "" : `${secondValue}S`,
    ].join("");

    return `${sign}P${dateParts}${timeParts === "" ? "" : `T${timeParts}`}`;
  }

  toJSON(): string {
    return this.toISOString();
  }

  toString(): string {
    return this.toISOString();
  }

  private totalMilliseconds(): number {
    return (
      (this.weeks * 7 + this.days) * millisecondsPerDay +
      this.hours * millisecondsPerHour +
      this.minutes * millisecondsPerMinute +
      this.seconds * millisecondsPerSecond +
      this.milliseconds
    );
  }

  private direction(): 1 | -1 {
    const first = Object.values(this.toObject()).find((value) => value !== 0);

    return first !== undefined && first < 0 ? -1 : 1;
  }
}

export class TempoImmutable {
  protected value: Date;
  protected zone: string;

  constructor(input: TempoInput = new Date(), options?: TempoOptions) {
    this.value = asDate(input, options);
    this.zone = zoneFromInput(input, options);
  }

  static now(options?: TempoOptions): TempoImmutable {
    return new TempoImmutable(new Date(), options);
  }

  static today(options?: TempoOptions): TempoImmutable {
    return TempoImmutable.now(options).startOfDay();
  }

  static tomorrow(options?: TempoOptions): TempoImmutable {
    return TempoImmutable.today(options).addDays(1);
  }

  static yesterday(options?: TempoOptions): TempoImmutable {
    return TempoImmutable.today(options).subDays(1);
  }

  static parse(input: TempoInput, options?: TempoOptions): TempoImmutable {
    return new TempoImmutable(input, options);
  }

  static tryParse(
    input: TempoInput,
    options?: TempoOptions,
  ): TempoImmutable | null {
    try {
      return TempoImmutable.parse(input, options);
    } catch {
      return null;
    }
  }

  static canParse(input: TempoInput, options?: TempoOptions): boolean {
    return TempoImmutable.tryParse(input, options) !== null;
  }

  static fromFormat(
    input: string,
    pattern: string,
    options?: TempoOptions,
  ): TempoImmutable {
    return new TempoImmutable(
      parseFromPattern(input, pattern, options),
      options,
    );
  }

  static tryFromFormat(
    input: string,
    pattern: string,
    options?: TempoOptions,
  ): TempoImmutable | null {
    try {
      return TempoImmutable.fromFormat(input, pattern, options);
    } catch {
      return null;
    }
  }

  static hasFormat(
    input: string,
    pattern: string,
    options?: TempoOptions,
  ): boolean {
    return TempoImmutable.tryFromFormat(input, pattern, options) !== null;
  }

  static create(components: TempoComponents): TempoImmutable {
    return new TempoImmutable(dateFromZonedComponents(components), {
      timeZone: components.timeZone,
    });
  }

  static createFromDate(
    year: number,
    month = 1,
    day = 1,
    options?: TempoOptions,
  ): TempoImmutable {
    return TempoImmutable.create({
      day,
      month,
      timeZone: options?.timeZone,
      year,
    });
  }

  static createMidnightDate(
    year: number,
    month: number,
    day: number,
    options?: TempoOptions,
  ): TempoImmutable {
    return TempoImmutable.createFromDate(year, month, day, options);
  }

  static createFromTime(
    hour = 0,
    minute = 0,
    second = 0,
    millisecond = 0,
    options?: TempoOptions,
  ): TempoImmutable {
    return TempoImmutable.today(options).setTime(
      hour,
      minute,
      second,
      millisecond,
    );
  }

  static fromObject(components: TempoComponents): TempoImmutable {
    return TempoImmutable.create(components);
  }

  static fromTimestamp(
    timestamp: number,
    options?: TempoOptions,
  ): TempoImmutable {
    return new TempoImmutable(fromNumericTimestamp(timestamp), options);
  }

  static fromTimestampMs(
    timestamp: number,
    options?: TempoOptions,
  ): TempoImmutable {
    assertFiniteNumber(timestamp, "Timestamp");
    return new TempoImmutable(new Date(timestamp), options);
  }

  static min(...items: readonly TempoInput[]): TempoImmutable {
    if (items.length === 0) {
      throw new RangeError("Tempo.min requires at least one input");
    }

    return items
      .map((item) => TempoImmutable.parse(item))
      .reduce((earliest, item) => (item.isBefore(earliest) ? item : earliest));
  }

  static max(...items: readonly TempoInput[]): TempoImmutable {
    if (items.length === 0) {
      throw new RangeError("Tempo.max requires at least one input");
    }

    return items
      .map((item) => TempoImmutable.parse(item))
      .reduce((latest, item) => (item.isAfter(latest) ? item : latest));
  }

  static average(start: TempoInput, end: TempoInput): TempoImmutable {
    const startTempo = TempoImmutable.parse(start);
    const endTempo = TempoImmutable.parse(end, {
      timeZone: startTempo.timeZone,
    });

    return new TempoImmutable(
      new Date(Math.trunc((startTempo.timestampMs + endTempo.timestampMs) / 2)),
      { timeZone: startTempo.timeZone },
    );
  }

  protected make(value: Date, timeZone = this.zone): this {
    const Constructor = this.constructor as new (
      input: TempoInput,
      options?: TempoOptions,
    ) => this;

    return new Constructor(value, { timeZone });
  }

  get timeZone(): string {
    return this.zone;
  }

  get timestamp(): number {
    return Math.trunc(this.value.getTime() / millisecondsPerSecond);
  }

  get timestampMs(): number {
    return this.value.getTime();
  }

  get year(): number {
    return this.parts().year;
  }

  get month(): number {
    return this.parts().month;
  }

  get quarter(): number {
    return Math.trunc((this.month - 1) / 3) + 1;
  }

  get day(): number {
    return this.parts().day;
  }

  get dayOfWeek(): number {
    return this.parts().weekday;
  }

  get isoWeekday(): number {
    return this.dayOfWeek === 0 ? 7 : this.dayOfWeek;
  }

  get isoWeek(): number {
    return isoWeekData(this.parts()).week;
  }

  get isoWeekYear(): number {
    return isoWeekData(this.parts()).year;
  }

  get weeksInISOYear(): number {
    return weeksInISOYear(this.isoWeekYear);
  }

  get dayOfYear(): number {
    return this.diffInDays(this.startOf("year")) + 1;
  }

  get hour(): number {
    return this.parts().hour;
  }

  get minute(): number {
    return this.parts().minute;
  }

  get second(): number {
    return this.parts().second;
  }

  get millisecond(): number {
    return this.parts().millisecond;
  }

  get offsetMinutes(): number {
    const parts = this.parts();
    const localAsUTC = dateFromPartsAsUTC(parts);

    return Math.trunc(
      (localAsUTC - this.value.getTime()) / millisecondsPerMinute,
    );
  }

  offsetString(separator: ":" | "" = ":"): string {
    return formatOffset(this.offsetMinutes, separator);
  }

  timezoneName(style: TimeZoneNameStyle = "short", locale = "en-US"): string {
    return timeZoneName(this.value, this.zone, style, locale);
  }

  isUtc(): boolean {
    return this.zone === defaultTimeZone && this.offsetMinutes === 0;
  }

  isLocal(): boolean {
    return this.zone === Intl.DateTimeFormat().resolvedOptions().timeZone;
  }

  isDST(): boolean {
    const parts = this.parts();
    const januaryOffset = this.offsetForDate(
      dateFromZonedComponents({
        day: 1,
        month: 1,
        timeZone: this.zone,
        year: parts.year,
      }),
    );
    const julyOffset = this.offsetForDate(
      dateFromZonedComponents({
        day: 1,
        month: 7,
        timeZone: this.zone,
        year: parts.year,
      }),
    );
    const standardOffset = Math.min(januaryOffset, julyOffset);

    return this.offsetMinutes > standardOffset;
  }

  isLeapYear(): boolean {
    const year = this.year;
    return year % 4 === 0 && (year % 100 !== 0 || year % 400 === 0);
  }

  daysInMonth(): number {
    return daysInMonth(this.year, this.month);
  }

  isWeekend(): boolean {
    return this.dayOfWeek === 0 || this.dayOfWeek === 6;
  }

  isSunday(): boolean {
    return this.dayOfWeek === 0;
  }

  isMonday(): boolean {
    return this.dayOfWeek === 1;
  }

  isTuesday(): boolean {
    return this.dayOfWeek === 2;
  }

  isWednesday(): boolean {
    return this.dayOfWeek === 3;
  }

  isThursday(): boolean {
    return this.dayOfWeek === 4;
  }

  isFriday(): boolean {
    return this.dayOfWeek === 5;
  }

  isSaturday(): boolean {
    return this.dayOfWeek === 6;
  }

  isWeekday(): boolean {
    return !this.isWeekend();
  }

  isPast(reference: TempoInput = new Date()): boolean {
    return this.isBefore(reference);
  }

  isFuture(reference: TempoInput = new Date()): boolean {
    return this.isAfter(reference);
  }

  isToday(reference: TempoInput = new Date()): boolean {
    return this.isSame(reference, "day");
  }

  isTomorrow(reference: TempoInput = new Date()): boolean {
    return this.isSame(
      TempoImmutable.parse(reference, { timeZone: this.zone }).addDays(1),
      "day",
    );
  }

  isYesterday(reference: TempoInput = new Date()): boolean {
    return this.isSame(
      TempoImmutable.parse(reference, { timeZone: this.zone }).subDays(1),
      "day",
    );
  }

  clone(): this {
    return this.make(this.value);
  }

  copy(): this {
    return this.clone();
  }

  timezone(timeZone: string): this {
    return this.setTimezone(timeZone);
  }

  setTimezone(timeZone: string, keepLocalTime = false): this {
    const nextZone = normalizeTimeZone(timeZone);

    if (!keepLocalTime) {
      return this.make(this.value, nextZone);
    }

    return this.make(
      dateFromZonedComponents({ ...this.toObject(), timeZone: nextZone }),
      nextZone,
    );
  }

  utc(): this {
    return this.setTimezone(defaultTimeZone);
  }

  local(): this {
    return this.setTimezone(Intl.DateTimeFormat().resolvedOptions().timeZone);
  }

  set(components: TempoSettableComponents): this {
    const timeZone = normalizeTimeZone(components.timeZone ?? this.zone);

    return this.make(
      dateFromZonedComponents(
        {
          ...this.toObject(),
          ...components,
          timeZone,
        },
        timeZone,
      ),
      timeZone,
    );
  }

  yearTo(year: number): this {
    return this.set({ year });
  }

  monthTo(month: number): this {
    return this.set({ month });
  }

  dayTo(day: number): this {
    return this.set({ day });
  }

  hourTo(hour: number): this {
    return this.set({ hour });
  }

  minuteTo(minute: number): this {
    return this.set({ minute });
  }

  secondTo(second: number): this {
    return this.set({ second });
  }

  millisecondTo(millisecond: number): this {
    return this.set({ millisecond });
  }

  setYear(year: number): this {
    return this.set({ year });
  }

  setMonth(month: number): this {
    return this.set({ month });
  }

  setDay(day: number): this {
    return this.set({ day });
  }

  setDate(year: number, month: number, day: number): this {
    return this.set({ day, month, year });
  }

  setHour(hour: number): this {
    return this.set({ hour });
  }

  setMinute(minute: number): this {
    return this.set({ minute });
  }

  setSecond(second: number): this {
    return this.set({ second });
  }

  setMillisecond(millisecond: number): this {
    return this.set({ millisecond });
  }

  setTime(
    hour: number,
    minute = this.minute,
    second = this.second,
    millisecond = this.millisecond,
  ): this {
    return this.set({ hour, millisecond, minute, second });
  }

  add(value: number, unit: TimeUnit): this {
    assertFiniteNumber(value, "Amount");

    const fixed = fixedUnitMilliseconds(unit);

    if (fixed !== null) {
      return this.make(new Date(this.value.getTime() + value * fixed));
    }

    switch (normalizeUnit(unit)) {
      case "month":
        return this.addMonths(value);
      case "quarter":
        return this.addMonths(value * 3);
      case "year":
        return this.addYears(value);
      default:
        return this.make(this.value);
    }
  }

  sub(value: number, unit: TimeUnit): this {
    return this.add(-value, unit);
  }

  addDuration(duration: DurationInput): this {
    const value = durationFromInput(duration);

    return this.addYears(value.years)
      .addMonths(value.quarters * 3 + value.months)
      .addWeeks(value.weeks)
      .addDays(value.days)
      .addHours(value.hours)
      .addMinutes(value.minutes)
      .addSeconds(value.seconds)
      .addMilliseconds(value.milliseconds);
  }

  subDuration(duration: DurationInput): this {
    return this.addDuration(durationFromInput(duration).negated());
  }

  addMilliseconds(milliseconds: number): this {
    return this.add(milliseconds, "millisecond");
  }

  subMilliseconds(milliseconds: number): this {
    return this.sub(milliseconds, "millisecond");
  }

  addSeconds(seconds: number): this {
    return this.add(seconds, "second");
  }

  subSeconds(seconds: number): this {
    return this.sub(seconds, "second");
  }

  addMinutes(minutes: number): this {
    return this.add(minutes, "minute");
  }

  subMinutes(minutes: number): this {
    return this.sub(minutes, "minute");
  }

  addHours(hours: number): this {
    return this.add(hours, "hour");
  }

  subHours(hours: number): this {
    return this.sub(hours, "hour");
  }

  addDays(days: number): this {
    return this.add(days, "day");
  }

  subDays(days: number): this {
    return this.sub(days, "day");
  }

  addWeekdays(days: number): this {
    assertFiniteNumber(days, "Weekdays");

    if (days === 0) {
      return this.clone();
    }

    const direction = days < 0 ? -1 : 1;
    let remaining = Math.abs(Math.trunc(days));
    let current = this.clone();

    while (remaining > 0) {
      current = current.addDays(direction);

      if (current.isWeekday()) {
        remaining -= 1;
      }
    }

    return current;
  }

  subWeekdays(days: number): this {
    return this.addWeekdays(-days);
  }

  addWeeks(weeks: number): this {
    return this.add(weeks, "week");
  }

  subWeeks(weeks: number): this {
    return this.sub(weeks, "week");
  }

  addMonths(months: number): this {
    assertFiniteNumber(months, "Months");
    const parts = this.toObject();

    return this.make(
      dateFromZonedComponents(
        {
          ...parts,
          month: parts.month + months,
        },
        this.zone,
      ),
    );
  }

  addMonthsNoOverflow(months: number): this {
    assertFiniteNumber(months, "Months");
    const parts = this.toObject();
    const firstOfTarget = dateFromZonedComponents(
      {
        ...parts,
        day: 1,
        month: parts.month + months,
      },
      this.zone,
    );
    const target = getZonedParts(firstOfTarget, this.zone);
    const day = Math.min(parts.day, daysInMonth(target.year, target.month));

    return this.make(
      dateFromZonedComponents(
        {
          ...parts,
          day,
          month: parts.month + months,
        },
        this.zone,
      ),
    );
  }

  subMonths(months: number): this {
    return this.addMonths(-months);
  }

  subMonthsNoOverflow(months: number): this {
    return this.addMonthsNoOverflow(-months);
  }

  addQuarters(quarters: number): this {
    return this.addMonths(quarters * 3);
  }

  subQuarters(quarters: number): this {
    return this.addQuarters(-quarters);
  }

  addYears(years: number): this {
    assertFiniteNumber(years, "Years");
    const parts = this.toObject();

    return this.make(
      dateFromZonedComponents(
        {
          ...parts,
          year: parts.year + years,
        },
        this.zone,
      ),
    );
  }

  addYearsNoOverflow(years: number): this {
    const parts = this.toObject();
    const day = Math.min(
      parts.day,
      daysInMonth(parts.year + years, parts.month),
    );

    return this.set({ day, year: parts.year + years });
  }

  age(reference: TempoInput = new Date()): number {
    return TempoImmutable.parse(reference, { timeZone: this.zone }).diffInYears(
      this,
    );
  }

  subYears(years: number): this {
    return this.addYears(-years);
  }

  subYearsNoOverflow(years: number): this {
    return this.addYearsNoOverflow(-years);
  }

  startOf(unit: BoundaryUnit, options?: StartOfWeekOptions): this {
    const parts = this.toObject();

    switch (unit) {
      case "second":
        return this.set({ millisecond: 0 });
      case "minute":
        return this.set({ millisecond: 0, second: 0 });
      case "hour":
        return this.set({ millisecond: 0, minute: 0, second: 0 });
      case "day":
        return this.set({ hour: 0, millisecond: 0, minute: 0, second: 0 });
      case "week": {
        const weekStartsOn = options?.weekStartsOn ?? 1;
        const delta = (parts.weekday - weekStartsOn + 7) % 7;
        return this.startOf("day").subDays(delta);
      }
      case "month":
        return this.set({
          day: 1,
          hour: 0,
          millisecond: 0,
          minute: 0,
          second: 0,
        });
      case "quarter":
        return this.set({
          day: 1,
          hour: 0,
          millisecond: 0,
          minute: 0,
          month: (this.quarter - 1) * 3 + 1,
          second: 0,
        });
      case "year":
        return this.set({
          day: 1,
          hour: 0,
          millisecond: 0,
          minute: 0,
          month: 1,
          second: 0,
        });
    }
  }

  endOf(unit: BoundaryUnit, options?: StartOfWeekOptions): this {
    switch (unit) {
      case "second":
        return this.startOf("second").addSeconds(1).subMilliseconds(1);
      case "minute":
        return this.startOf("minute").addMinutes(1).subMilliseconds(1);
      case "hour":
        return this.startOf("hour").addHours(1).subMilliseconds(1);
      case "day":
        return this.startOf("day").addDays(1).subMilliseconds(1);
      case "week":
        return this.startOf("week", options).addWeeks(1).subMilliseconds(1);
      case "month":
        return this.startOf("month").addMonths(1).subMilliseconds(1);
      case "quarter":
        return this.startOf("quarter").addQuarters(1).subMilliseconds(1);
      case "year":
        return this.startOf("year").addYears(1).subMilliseconds(1);
    }
  }

  isStartOf(unit: BoundaryUnit, options?: StartOfWeekOptions): boolean {
    return this.isSame(this.startOf(unit, options));
  }

  isEndOf(unit: BoundaryUnit, options?: StartOfWeekOptions): boolean {
    return this.isSame(this.endOf(unit, options));
  }

  startOfDay(): this {
    return this.startOf("day");
  }

  endOfDay(): this {
    return this.endOf("day");
  }

  startOfWeek(options?: StartOfWeekOptions): this {
    return this.startOf("week", options);
  }

  endOfWeek(options?: StartOfWeekOptions): this {
    return this.endOf("week", options);
  }

  startOfMonth(): this {
    return this.startOf("month");
  }

  endOfMonth(): this {
    return this.endOf("month");
  }

  startOfQuarter(): this {
    return this.startOf("quarter");
  }

  endOfQuarter(): this {
    return this.endOf("quarter");
  }

  firstOfMonth(weekday?: WeekdayInput): this {
    const first = this.startOf("month");

    if (weekday === undefined) {
      return first;
    }

    const target = resolveWeekday(weekday);
    const delta = (target - first.dayOfWeek + 7) % 7;

    return first.addDays(delta);
  }

  lastOfMonth(weekday?: WeekdayInput): this {
    const last = this.endOf("month").startOf("day");

    if (weekday === undefined) {
      return last;
    }

    const target = resolveWeekday(weekday);
    const delta = (last.dayOfWeek - target + 7) % 7;

    return last.subDays(delta);
  }

  nthOfMonth(occurrence: number, weekday: WeekdayInput): this | null {
    if (!Number.isInteger(occurrence) || occurrence === 0) {
      throw new RangeError(
        "Tempo nthOfMonth occurrence must be a non-zero integer",
      );
    }

    const currentMonth = this.month;
    const candidate =
      occurrence > 0
        ? this.firstOfMonth(weekday).addWeeks(occurrence - 1)
        : this.lastOfMonth(weekday).subWeeks(Math.abs(occurrence) - 1);

    return candidate.month === currentMonth ? candidate : null;
  }

  firstOfQuarter(weekday?: WeekdayInput): this {
    const first = this.startOf("quarter");

    if (weekday === undefined) {
      return first;
    }

    const target = resolveWeekday(weekday);
    const delta = (target - first.dayOfWeek + 7) % 7;

    return first.addDays(delta);
  }

  lastOfQuarter(weekday?: WeekdayInput): this {
    const last = this.endOf("quarter").startOf("day");

    if (weekday === undefined) {
      return last;
    }

    const target = resolveWeekday(weekday);
    const delta = (last.dayOfWeek - target + 7) % 7;

    return last.subDays(delta);
  }

  firstOfYear(weekday?: WeekdayInput): this {
    const first = this.startOf("year");

    if (weekday === undefined) {
      return first;
    }

    const target = resolveWeekday(weekday);
    const delta = (target - first.dayOfWeek + 7) % 7;

    return first.addDays(delta);
  }

  lastOfYear(weekday?: WeekdayInput): this {
    const last = this.endOf("year").startOf("day");

    if (weekday === undefined) {
      return last;
    }

    const target = resolveWeekday(weekday);
    const delta = (last.dayOfWeek - target + 7) % 7;

    return last.subDays(delta);
  }

  startOfYear(): this {
    return this.startOf("year");
  }

  endOfYear(): this {
    return this.endOf("year");
  }

  floor(unit: TimeUnit): this {
    const fixed = fixedUnitMilliseconds(unit);

    if (fixed === null) {
      return this.startOf(normalizeUnit(unit) as BoundaryUnit);
    }

    return this.make(new Date(Math.floor(this.timestampMs / fixed) * fixed));
  }

  ceil(unit: TimeUnit): this {
    const floored = this.floor(unit);

    if (floored.isSame(this)) {
      return floored;
    }

    return floored.add(1, unit);
  }

  round(unit: TimeUnit): this {
    const fixed = fixedUnitMilliseconds(unit);

    if (fixed === null) {
      const start = this.startOf(normalizeUnit(unit) as BoundaryUnit);
      const end = this.endOf(normalizeUnit(unit) as BoundaryUnit);
      const midpoint =
        start.timestampMs + (end.timestampMs - start.timestampMs) / 2;

      return this.timestampMs >= midpoint ? this.ceil(unit) : start;
    }

    return this.make(new Date(Math.round(this.timestampMs / fixed) * fixed));
  }

  next(weekday: WeekdayInput): this {
    const target = resolveWeekday(weekday);
    const delta = (target - this.dayOfWeek + 7) % 7 || 7;

    return this.addDays(delta);
  }

  previous(weekday: WeekdayInput): this {
    const target = resolveWeekday(weekday);
    const delta = (this.dayOfWeek - target + 7) % 7 || 7;

    return this.subDays(delta);
  }

  nextWeekday(): this {
    let next = this.addDays(1);

    while (next.isWeekend()) {
      next = next.addDays(1);
    }

    return next;
  }

  previousWeekday(): this {
    let previous = this.subDays(1);

    while (previous.isWeekend()) {
      previous = previous.subDays(1);
    }

    return previous;
  }

  diff(
    other: TempoInput,
    unit: TimeUnit = "millisecond",
    options?: DiffOptions,
  ): number {
    const otherDate = asDate(other);
    const rawMilliseconds = this.timestampMs - otherDate.getTime();
    const milliseconds = options?.absolute
      ? Math.abs(rawMilliseconds)
      : rawMilliseconds;
    const fixed = fixedUnitMilliseconds(unit);

    if (fixed !== null) {
      const value = milliseconds / fixed;
      return options?.float ? value : Math.trunc(value);
    }

    const otherTempo = TempoImmutable.parse(other, { timeZone: this.zone });
    const sign = milliseconds < 0 ? -1 : 1;
    const start = sign < 0 ? this : otherTempo;
    const end = sign < 0 ? otherTempo : this;
    const startParts = start.toObject();
    const endParts = end.toObject();
    let months =
      (endParts.year - startParts.year) * 12 +
      (endParts.month - startParts.month);

    if (endParts.day < startParts.day) {
      months -= 1;
    }

    const result =
      normalizeUnit(unit) === "year"
        ? months / 12
        : normalizeUnit(unit) === "quarter"
          ? months / 3
          : months;
    const signed = options?.absolute ? Math.abs(result) : result * sign;

    return options?.float ? signed : Math.trunc(signed);
  }

  diffInMilliseconds(other: TempoInput, options?: DiffOptions): number {
    return this.diff(other, "millisecond", options);
  }

  diffInSeconds(other: TempoInput, options?: DiffOptions): number {
    return this.diff(other, "second", options);
  }

  diffInMinutes(other: TempoInput, options?: DiffOptions): number {
    return this.diff(other, "minute", options);
  }

  diffInHours(other: TempoInput, options?: DiffOptions): number {
    return this.diff(other, "hour", options);
  }

  diffInDays(other: TempoInput, options?: DiffOptions): number {
    return this.diff(other, "day", options);
  }

  diffInWeeks(other: TempoInput, options?: DiffOptions): number {
    return this.diff(other, "week", options);
  }

  diffInWeekdays(other: TempoInput, options?: DiffOptions): number {
    return this.diffFilteredDays(other, (item) => item.isWeekday(), options);
  }

  diffInWeekendDays(other: TempoInput, options?: DiffOptions): number {
    return this.diffFilteredDays(other, (item) => item.isWeekend(), options);
  }

  diffInMonths(other: TempoInput, options?: DiffOptions): number {
    return this.diff(other, "month", options);
  }

  diffInYears(other: TempoInput, options?: DiffOptions): number {
    return this.diff(other, "year", options);
  }

  diffForHumans(
    other: TempoInput = new Date(),
    options?: HumanDiffOptions,
  ): string {
    const rawMilliseconds = this.timestampMs - asDate(other).getTime();
    const unit = options?.unit ?? bestRelativeUnit(rawMilliseconds);
    const value = Math.round(rawMilliseconds / unitDivisor(unit));
    const formatter = new Intl.RelativeTimeFormat(options?.locale ?? "en-US", {
      numeric: options?.numeric ?? "always",
      style: options?.style ?? "long",
    });

    return formatter.format(options?.absolute ? Math.abs(value) : value, unit);
  }

  isBefore(other: TempoInput, unit: ComparisonUnit = "millisecond"): boolean {
    return (
      this.comparableValue(unit) <
      TempoImmutable.parse(other, { timeZone: this.zone }).comparableValue(unit)
    );
  }

  isAfter(other: TempoInput, unit: ComparisonUnit = "millisecond"): boolean {
    return (
      this.comparableValue(unit) >
      TempoImmutable.parse(other, { timeZone: this.zone }).comparableValue(unit)
    );
  }

  isSame(other: TempoInput, unit: ComparisonUnit = "millisecond"): boolean {
    return (
      this.comparableValue(unit) ===
      TempoImmutable.parse(other, { timeZone: this.zone }).comparableValue(unit)
    );
  }

  isSameSecond(other: TempoInput): boolean {
    return this.isSame(other, "second");
  }

  isSameMinute(other: TempoInput): boolean {
    return this.isSame(other, "minute");
  }

  isSameHour(other: TempoInput): boolean {
    return this.isSame(other, "hour");
  }

  isSameDay(other: TempoInput): boolean {
    return this.isSame(other, "day");
  }

  isSameWeek(other: TempoInput): boolean {
    return this.isSame(other, "week");
  }

  isSameMonth(other: TempoInput): boolean {
    return this.isSame(other, "month");
  }

  isSameQuarter(other: TempoInput): boolean {
    return this.isSame(other, "quarter");
  }

  isSameYear(other: TempoInput): boolean {
    return this.isSame(other, "year");
  }

  isBirthday(other: TempoInput = new Date()): boolean {
    const compare = TempoImmutable.parse(other, { timeZone: this.zone });

    return this.month === compare.month && this.day === compare.day;
  }

  isSameOrBefore(other: TempoInput, unit?: ComparisonUnit): boolean {
    return this.isSame(other, unit) || this.isBefore(other, unit);
  }

  isSameOrAfter(other: TempoInput, unit?: ComparisonUnit): boolean {
    return this.isSame(other, unit) || this.isAfter(other, unit);
  }

  isBetween(
    start: TempoInput,
    end: TempoInput,
    unit: ComparisonUnit = "millisecond",
    inclusivity: "()" | "[]" | "[)" | "(]" = "()",
  ): boolean {
    const afterStart = inclusivity.startsWith("[")
      ? this.isSameOrAfter(start, unit)
      : this.isAfter(start, unit);
    const beforeEnd = inclusivity.endsWith("]")
      ? this.isSameOrBefore(end, unit)
      : this.isBefore(end, unit);

    return afterStart && beforeEnd;
  }

  clamp(min: TempoInput, max: TempoInput): this {
    const minTempo = TempoImmutable.parse(min, { timeZone: this.zone });
    const maxTempo = TempoImmutable.parse(max, { timeZone: this.zone });

    if (minTempo.isAfter(maxTempo)) {
      throw new RangeError("Tempo clamp minimum must be before maximum");
    }

    if (this.isBefore(minTempo)) {
      return this.make(minTempo.toDate(), minTempo.timeZone);
    }

    if (this.isAfter(maxTempo)) {
      return this.make(maxTempo.toDate(), maxTempo.timeZone);
    }

    return this.clone();
  }

  average(other: TempoInput): this {
    const end = TempoImmutable.parse(other, { timeZone: this.zone });

    return this.make(
      new Date(Math.trunc((this.timestampMs + end.timestampMs) / 2)),
    );
  }

  closest(...items: readonly TempoInput[]): this {
    if (items.length === 0) {
      throw new RangeError("Tempo.closest requires at least one input");
    }

    const closest = items
      .map((item) => TempoImmutable.parse(item, { timeZone: this.zone }))
      .reduce((best, item) =>
        Math.abs(item.timestampMs - this.timestampMs) <
        Math.abs(best.timestampMs - this.timestampMs)
          ? item
          : best,
      );

    return this.make(closest.toDate(), closest.timeZone);
  }

  farthest(...items: readonly TempoInput[]): this {
    if (items.length === 0) {
      throw new RangeError("Tempo.farthest requires at least one input");
    }

    const farthest = items
      .map((item) => TempoImmutable.parse(item, { timeZone: this.zone }))
      .reduce((best, item) =>
        Math.abs(item.timestampMs - this.timestampMs) >
        Math.abs(best.timestampMs - this.timestampMs)
          ? item
          : best,
      );

    return this.make(farthest.toDate(), farthest.timeZone);
  }

  min(other: TempoInput): this {
    return this.isBefore(other)
      ? this
      : this.make(asDate(other), zoneFromInput(other, undefined));
  }

  max(other: TempoInput): this {
    return this.isAfter(other)
      ? this
      : this.make(asDate(other), zoneFromInput(other, undefined));
  }

  format(pattern: string, options?: FormatOptions): string {
    const locale = options?.locale ?? "en-US";
    const timeZone = normalizeTimeZone(options?.timeZone ?? this.zone);
    const parts = getZonedParts(this.value, timeZone);
    const offset = this.offsetFor(timeZone);
    const monthsShort = monthNames(locale, "short");
    const monthsLong = monthNames(locale, "long");
    const weekdaysShort = weekdayNames(locale, "short");
    const weekdaysLong = weekdayNames(locale, "long");
    const hour12 = parts.hour % 12 === 0 ? 12 : parts.hour % 12;
    const replacements: Record<string, string> = {
      A: parts.hour < 12 ? "AM" : "PM",
      a: parts.hour < 12 ? "am" : "pm",
      D: String(parts.day),
      DD: pad(parts.day),
      Do: ordinal(parts.day),
      d: String(parts.weekday),
      ddd: weekdaysShort[parts.weekday] ?? "",
      dddd: weekdaysLong[parts.weekday] ?? "",
      H: String(parts.hour),
      HH: pad(parts.hour),
      h: String(hour12),
      hh: pad(hour12),
      M: String(parts.month),
      MM: pad(parts.month),
      MMM: monthsShort[parts.month - 1] ?? "",
      MMMM: monthsLong[parts.month - 1] ?? "",
      m: String(parts.minute),
      mm: pad(parts.minute),
      SSS: pad(parts.millisecond, 3),
      s: String(parts.second),
      ss: pad(parts.second),
      X: String(this.timestamp),
      x: String(this.timestampMs),
      Y: String(parts.year),
      YY: pad(parts.year % 100),
      YYYY: pad(parts.year, 4),
      Z: formatOffset(offset, ":"),
      ZZ: formatOffset(offset, ""),
    };

    return pattern.replace(
      /\[[^\]]*]|YYYY|MMMM|dddd|MMM|ddd|SSS|Do|YY|ZZ|MM|DD|HH|hh|mm|ss|Z|X|x|Y|M|D|H|h|m|s|A|a|d/g,
      (token) =>
        token.startsWith("[") && token.endsWith("]")
          ? token.slice(1, -1)
          : (replacements[token] ?? token),
    );
  }

  formatIntl(
    options?: Intl.DateTimeFormatOptions & { readonly locale?: string },
  ): string {
    const { locale, ...dateTimeOptions } = options ?? {};

    return new Intl.DateTimeFormat(locale, {
      timeZone: this.zone,
      ...dateTimeOptions,
    }).format(this.value);
  }

  toDate(): Date {
    return new Date(this.value.getTime());
  }

  toDateString(): string {
    return this.format("YYYY-MM-DD");
  }

  toTimeString(precision: TimeStringPrecision = "second"): string {
    const base = this.format("HH:mm:ss");

    return precision === "millisecond"
      ? `${base}.${pad(this.millisecond, 3)}`
      : base;
  }

  toDateTimeString(): string {
    return this.format("YYYY-MM-DD HH:mm:ss");
  }

  toDateTimeLocalString(precision: TimeStringPrecision = "second"): string {
    return `${this.toDateString()}T${this.toTimeString(precision)}`;
  }

  toISOString(): string {
    return this.value.toISOString();
  }

  toIso8601String(): string {
    return this.format("YYYY-MM-DDTHH:mm:ssZ");
  }

  toRfc3339String(precision: TimeStringPrecision = "second"): string {
    return `${this.toDateTimeLocalString(precision)}${this.offsetString(":")}`;
  }

  toRfc7231String(): string {
    return this.utc().format("ddd, DD MMM YYYY HH:mm:ss [GMT]");
  }

  toCookieString(): string {
    return this.utc().format("ddd, DD-MMM-YYYY HH:mm:ss [GMT]");
  }

  toAtomString(): string {
    return this.toRfc3339String();
  }

  toRssString(): string {
    return this.format("ddd, DD MMM YYYY HH:mm:ss ZZ");
  }

  toUnixString(): string {
    return String(this.timestamp);
  }

  toJSON(): string {
    return this.toISOString();
  }

  toObject(): TempoObject {
    const parts = this.parts();

    return {
      ...parts,
      offsetMinutes: this.offsetMinutes,
      timeZone: this.zone,
      weekday: parts.weekday,
    };
  }

  toMap(): Map<keyof TempoObject, TempoObject[keyof TempoObject]> {
    return new Map(
      Object.entries(this.toObject()) as Array<
        [keyof TempoObject, TempoObject[keyof TempoObject]]
      >,
    );
  }

  toArray(): [number, number, number, number, number, number, number] {
    const parts = this.parts();

    return [
      parts.year,
      parts.month,
      parts.day,
      parts.hour,
      parts.minute,
      parts.second,
      parts.millisecond,
    ];
  }

  valueOf(): number {
    return this.timestampMs;
  }

  toString(): string {
    return this.toISOString();
  }

  intervalUntil(end: TempoInput): TempoInterval {
    return new TempoInterval(this, end);
  }

  periodUntil(end: TempoInput, options?: PeriodOptions): TempoPeriod {
    return new TempoPeriod(this, end, options);
  }

  private parts(): ZonedParts {
    return getZonedParts(this.value, this.zone);
  }

  private offsetFor(timeZone: string): number {
    const parts = getZonedParts(this.value, timeZone);
    const localAsUTC = dateFromPartsAsUTC(parts);

    return Math.trunc(
      (localAsUTC - this.value.getTime()) / millisecondsPerMinute,
    );
  }

  private offsetForDate(date: Date): number {
    const parts = getZonedParts(date, this.zone);
    const localAsUTC = dateFromPartsAsUTC(parts);

    return Math.trunc((localAsUTC - date.getTime()) / millisecondsPerMinute);
  }

  private comparableValue(unit: ComparisonUnit): number {
    return unit === "millisecond"
      ? this.timestampMs
      : this.startOf(unit).timestampMs;
  }

  private diffFilteredDays(
    other: TempoInput,
    predicate: (item: TempoImmutable) => boolean,
    options?: DiffOptions,
  ): number {
    const otherTempo = TempoImmutable.parse(other, { timeZone: this.zone });
    const sign = this.isBefore(otherTempo, "day") ? -1 : 1;
    const start = sign < 0 ? this.startOf("day") : otherTempo.startOf("day");
    const end = sign < 0 ? otherTempo.startOf("day") : this.startOf("day");
    let current = start;
    let count = 0;

    while (current.isBefore(end, "day")) {
      current = current.addDays(1);

      if (current.isSameOrBefore(end, "day") && predicate(current)) {
        count += 1;
      }
    }

    const result = options?.absolute ? count : count * sign;

    return options?.float ? result : Math.trunc(result);
  }
}

export class Tempo extends TempoImmutable {
  static override now(options?: TempoOptions): Tempo {
    return new Tempo(new Date(), options);
  }

  static override today(options?: TempoOptions): Tempo {
    return Tempo.now(options).startOfDay();
  }

  static override tomorrow(options?: TempoOptions): Tempo {
    return Tempo.today(options).addDays(1);
  }

  static override yesterday(options?: TempoOptions): Tempo {
    return Tempo.today(options).subDays(1);
  }

  static override parse(input: TempoInput, options?: TempoOptions): Tempo {
    return new Tempo(input, options);
  }

  static override tryParse(
    input: TempoInput,
    options?: TempoOptions,
  ): Tempo | null {
    try {
      return Tempo.parse(input, options);
    } catch {
      return null;
    }
  }

  static override fromFormat(
    input: string,
    pattern: string,
    options?: TempoOptions,
  ): Tempo {
    return new Tempo(parseFromPattern(input, pattern, options), options);
  }

  static override tryFromFormat(
    input: string,
    pattern: string,
    options?: TempoOptions,
  ): Tempo | null {
    try {
      return Tempo.fromFormat(input, pattern, options);
    } catch {
      return null;
    }
  }

  static override create(components: TempoComponents): Tempo {
    return new Tempo(dateFromZonedComponents(components), {
      timeZone: components.timeZone,
    });
  }

  static override createFromDate(
    year: number,
    month = 1,
    day = 1,
    options?: TempoOptions,
  ): Tempo {
    return Tempo.create({ day, month, timeZone: options?.timeZone, year });
  }

  static override createMidnightDate(
    year: number,
    month: number,
    day: number,
    options?: TempoOptions,
  ): Tempo {
    return Tempo.createFromDate(year, month, day, options);
  }

  static override createFromTime(
    hour = 0,
    minute = 0,
    second = 0,
    millisecond = 0,
    options?: TempoOptions,
  ): Tempo {
    return Tempo.today(options).setTime(hour, minute, second, millisecond);
  }

  static override fromObject(components: TempoComponents): Tempo {
    return Tempo.create(components);
  }

  static override fromTimestamp(
    timestamp: number,
    options?: TempoOptions,
  ): Tempo {
    return new Tempo(fromNumericTimestamp(timestamp), options);
  }

  static override fromTimestampMs(
    timestamp: number,
    options?: TempoOptions,
  ): Tempo {
    assertFiniteNumber(timestamp, "Timestamp");
    return new Tempo(new Date(timestamp), options);
  }
}

export class TempoMutable extends TempoImmutable {
  static override now(options?: TempoOptions): TempoMutable {
    return new TempoMutable(new Date(), options);
  }

  static override today(options?: TempoOptions): TempoMutable {
    return TempoMutable.now(options).startOfDay();
  }

  static override tomorrow(options?: TempoOptions): TempoMutable {
    return TempoMutable.today(options).addDays(1);
  }

  static override yesterday(options?: TempoOptions): TempoMutable {
    return TempoMutable.today(options).subDays(1);
  }

  static override parse(
    input: TempoInput,
    options?: TempoOptions,
  ): TempoMutable {
    return new TempoMutable(input, options);
  }

  static override tryParse(
    input: TempoInput,
    options?: TempoOptions,
  ): TempoMutable | null {
    try {
      return TempoMutable.parse(input, options);
    } catch {
      return null;
    }
  }

  static override fromFormat(
    input: string,
    pattern: string,
    options?: TempoOptions,
  ): TempoMutable {
    return new TempoMutable(parseFromPattern(input, pattern, options), options);
  }

  static override tryFromFormat(
    input: string,
    pattern: string,
    options?: TempoOptions,
  ): TempoMutable | null {
    try {
      return TempoMutable.fromFormat(input, pattern, options);
    } catch {
      return null;
    }
  }

  static override create(components: TempoComponents): TempoMutable {
    return new TempoMutable(dateFromZonedComponents(components), {
      timeZone: components.timeZone,
    });
  }

  static override createFromDate(
    year: number,
    month = 1,
    day = 1,
    options?: TempoOptions,
  ): TempoMutable {
    return TempoMutable.create({
      day,
      month,
      timeZone: options?.timeZone,
      year,
    });
  }

  static override createMidnightDate(
    year: number,
    month: number,
    day: number,
    options?: TempoOptions,
  ): TempoMutable {
    return TempoMutable.createFromDate(year, month, day, options);
  }

  static override createFromTime(
    hour = 0,
    minute = 0,
    second = 0,
    millisecond = 0,
    options?: TempoOptions,
  ): TempoMutable {
    return TempoMutable.today(options).setTime(
      hour,
      minute,
      second,
      millisecond,
    );
  }

  static override fromObject(components: TempoComponents): TempoMutable {
    return TempoMutable.create(components);
  }

  static override fromTimestamp(
    timestamp: number,
    options?: TempoOptions,
  ): TempoMutable {
    return new TempoMutable(fromNumericTimestamp(timestamp), options);
  }

  static override fromTimestampMs(
    timestamp: number,
    options?: TempoOptions,
  ): TempoMutable {
    assertFiniteNumber(timestamp, "Timestamp");
    return new TempoMutable(new Date(timestamp), options);
  }

  protected override make(value: Date, timeZone = this.zone): this {
    this.value = new Date(value.getTime());
    this.zone = normalizeTimeZone(timeZone);

    return this;
  }
}

export class TempoInterval {
  readonly start: TempoImmutable;
  readonly end: TempoImmutable;

  constructor(start: TempoInput, end: TempoInput) {
    this.start = TempoImmutable.parse(start);
    this.end = TempoImmutable.parse(end);
  }

  get isInverted(): boolean {
    return this.start.isAfter(this.end);
  }

  get milliseconds(): number {
    return this.end.diffInMilliseconds(this.start);
  }

  get seconds(): number {
    return this.end.diffInSeconds(this.start);
  }

  get minutes(): number {
    return this.end.diffInMinutes(this.start);
  }

  get hours(): number {
    return this.end.diffInHours(this.start);
  }

  get days(): number {
    return this.end.diffInDays(this.start);
  }

  get weeks(): number {
    return this.end.diffInWeeks(this.start);
  }

  get months(): number {
    return this.end.diffInMonths(this.start);
  }

  get years(): number {
    return this.end.diffInYears(this.start);
  }

  contains(
    input: TempoInput,
    inclusivity: "()" | "[]" | "[)" | "(]" = "[]",
  ): boolean {
    return TempoImmutable.parse(input).isBetween(
      this.start,
      this.end,
      "millisecond",
      inclusivity,
    );
  }

  overlaps(other: TempoInterval): boolean {
    return this.start.isBefore(other.end) && this.end.isAfter(other.start);
  }

  intersection(other: TempoInterval): TempoInterval | null {
    if (!this.overlaps(other)) {
      return null;
    }

    return new TempoInterval(
      this.start.isAfter(other.start) ? this.start : other.start,
      this.end.isBefore(other.end) ? this.end : other.end,
    );
  }

  union(other: TempoInterval): TempoInterval {
    return new TempoInterval(
      this.start.isBefore(other.start) ? this.start : other.start,
      this.end.isAfter(other.end) ? this.end : other.end,
    );
  }

  toDuration(): TempoDuration {
    return new TempoDuration({ milliseconds: this.milliseconds }).normalized();
  }
}

export class TempoPeriod implements Iterable<Tempo> {
  readonly start: Tempo;
  readonly end: Tempo;
  readonly step: DurationInput;
  readonly includeEnd: boolean;

  constructor(start: TempoInput, end: TempoInput, options?: PeriodOptions) {
    this.start = Tempo.parse(start);
    this.end = Tempo.parse(end);
    this.step = options?.step ?? { days: 1 };
    this.includeEnd = options?.includeEnd ?? true;
  }

  *[Symbol.iterator](): Iterator<Tempo> {
    let current = this.start;
    const forward = this.end.isSameOrAfter(this.start);

    while (
      this.includeEnd
        ? forward
          ? current.isSameOrBefore(this.end)
          : current.isSameOrAfter(this.end)
        : forward
          ? current.isBefore(this.end)
          : current.isAfter(this.end)
    ) {
      yield current;
      const next = current.addDuration(this.step);

      if (next.isSame(current)) {
        throw new RangeError("TempoPeriod step must advance the period");
      }
      if (forward ? next.isBefore(current) : next.isAfter(current)) {
        throw new RangeError("TempoPeriod step must advance toward the end");
      }

      current = next;
    }
  }

  first(): Tempo | null {
    const next = this[Symbol.iterator]().next();

    return next.done === true ? null : next.value;
  }

  last(): Tempo | null {
    let last: Tempo | null = null;

    for (const value of this) {
      last = value;
    }

    return last;
  }

  count(): number {
    let total = 0;

    for (const _value of this) {
      total += 1;
    }

    return total;
  }

  isEmpty(): boolean {
    return this.first() === null;
  }

  contains(input: TempoInput): boolean {
    const value = Tempo.parse(input, { timeZone: this.start.timeZone });
    const forward = this.end.isSameOrAfter(this.start);
    const afterStart = forward
      ? value.isSameOrAfter(this.start)
      : value.isSameOrBefore(this.start);
    const beforeEnd = forward
      ? this.includeEnd
        ? value.isSameOrBefore(this.end)
        : value.isBefore(this.end)
      : this.includeEnd
        ? value.isSameOrAfter(this.end)
        : value.isAfter(this.end);

    return afterStart && beforeEnd;
  }

  filter(predicate: (value: Tempo, index: number) => boolean): Tempo[] {
    const values: Tempo[] = [];
    let index = 0;

    for (const value of this) {
      if (predicate(value, index)) {
        values.push(value);
      }

      index += 1;
    }

    return values;
  }

  map<T>(mapper: (value: Tempo, index: number) => T): T[] {
    const values: T[] = [];
    let index = 0;

    for (const value of this) {
      values.push(mapper(value, index));
      index += 1;
    }

    return values;
  }

  every(step: DurationInput): TempoPeriod {
    return new TempoPeriod(this.start, this.end, {
      includeEnd: this.includeEnd,
      step,
    });
  }

  toDuration(): TempoDuration {
    return this.start.intervalUntil(this.end).toDuration();
  }

  toArray(): Tempo[] {
    return Array.from(this);
  }
}

export class TempoFactory {
  private readonly nowValue: Date | null;
  private readonly zone: string;

  private constructor(nowValue: Date | null, timeZone = defaultTimeZone) {
    this.nowValue = nowValue === null ? null : new Date(nowValue.getTime());
    this.zone = normalizeTimeZone(timeZone);
  }

  static create(options?: TempoOptions): TempoFactory {
    return new TempoFactory(null, options?.timeZone);
  }

  static withTestNow(input: TempoInput, options?: TempoOptions): TempoFactory {
    return new TempoFactory(
      asDate(input, options),
      zoneFromInput(input, options),
    );
  }

  now(): Tempo {
    return Tempo.parse(this.nowValue ?? new Date(), { timeZone: this.zone });
  }

  today(): Tempo {
    return this.now().startOfDay();
  }

  tomorrow(): Tempo {
    return this.today().addDays(1);
  }

  yesterday(): Tempo {
    return this.today().subDays(1);
  }

  immutableNow(): TempoImmutable {
    return TempoImmutable.parse(this.nowValue ?? new Date(), {
      timeZone: this.zone,
    });
  }

  mutableNow(): TempoMutable {
    return TempoMutable.parse(this.nowValue ?? new Date(), {
      timeZone: this.zone,
    });
  }

  parse(input: TempoInput, options?: TempoOptions): Tempo {
    return Tempo.parse(input, { timeZone: options?.timeZone ?? this.zone });
  }

  tryParse(input: TempoInput, options?: TempoOptions): Tempo | null {
    return Tempo.tryParse(input, { timeZone: options?.timeZone ?? this.zone });
  }

  canParse(input: TempoInput, options?: TempoOptions): boolean {
    return this.tryParse(input, options) !== null;
  }

  fromFormat(input: string, pattern: string, options?: TempoOptions): Tempo {
    return Tempo.fromFormat(input, pattern, {
      timeZone: options?.timeZone ?? this.zone,
    });
  }

  tryFromFormat(
    input: string,
    pattern: string,
    options?: TempoOptions,
  ): Tempo | null {
    return Tempo.tryFromFormat(input, pattern, {
      timeZone: options?.timeZone ?? this.zone,
    });
  }

  hasFormat(input: string, pattern: string, options?: TempoOptions): boolean {
    return this.tryFromFormat(input, pattern, options) !== null;
  }

  create(components: TempoComponents): Tempo {
    return Tempo.create({ timeZone: this.zone, ...components });
  }

  createFromDate(year: number, month = 1, day = 1): Tempo {
    return this.create({ day, month, year });
  }

  createMidnightDate(year: number, month: number, day: number): Tempo {
    return this.createFromDate(year, month, day);
  }

  createFromTime(hour = 0, minute = 0, second = 0, millisecond = 0): Tempo {
    return this.today().setTime(hour, minute, second, millisecond);
  }

  fromObject(components: TempoComponents): Tempo {
    return this.create(components);
  }

  fromTimestamp(timestamp: number, options?: TempoOptions): Tempo {
    return Tempo.fromTimestamp(timestamp, {
      timeZone: options?.timeZone ?? this.zone,
    });
  }

  fromTimestampMs(timestamp: number, options?: TempoOptions): Tempo {
    return Tempo.fromTimestampMs(timestamp, {
      timeZone: options?.timeZone ?? this.zone,
    });
  }
}

export const now = Tempo.now;
export const today = Tempo.today;
export const tomorrow = Tempo.tomorrow;
export const yesterday = Tempo.yesterday;
export const parse = Tempo.parse;
export const tryParse = Tempo.tryParse;
export const canParse = Tempo.canParse;
export const min = Tempo.min;
export const max = Tempo.max;
export const average = Tempo.average;
export const fromFormat = Tempo.fromFormat;
export const tryFromFormat = Tempo.tryFromFormat;
export const hasFormat = Tempo.hasFormat;
export const create = Tempo.create;
export const createFromDate = Tempo.createFromDate;
export const createFromTime = Tempo.createFromTime;
export const createMidnightDate = Tempo.createMidnightDate;
export const fromObject = Tempo.fromObject;
export const fromTimestamp = Tempo.fromTimestamp;
export const fromTimestampMs = Tempo.fromTimestampMs;
export const createDuration = TempoDuration.from;
export const parseDuration = TempoDuration.parse;
export const createFactory = TempoFactory.create;

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

export type DiffOptions = {
  readonly absolute?: boolean;
  readonly float?: boolean;
};

export type FormatOptions = {
  readonly locale?: string;
  readonly timeZone?: string;
};

export type StartOfWeekOptions = {
  readonly weekStartsOn?: number;
};

export type PeriodOptions = {
  readonly step?: DurationLike;
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
  timeZoneName?: "shortOffset",
): Intl.DateTimeFormat => {
  const key = `${timeZone}|${timeZoneName ?? "parts"}`;
  const cached = formatterCache.get(key);

  if (cached !== undefined) {
    return cached;
  }

  const formatter = new Intl.DateTimeFormat("en-US-u-nu-latn", {
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

  static parse(input: TempoInput, options?: TempoOptions): TempoImmutable {
    return new TempoImmutable(input, options);
  }

  static create(components: TempoComponents): TempoImmutable {
    return new TempoImmutable(dateFromZonedComponents(components), {
      timeZone: components.timeZone,
    });
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

  addDuration(duration: DurationLike): this {
    return this.addYears(duration.years ?? 0)
      .addMonths((duration.quarters ?? 0) * 3 + (duration.months ?? 0))
      .addWeeks(duration.weeks ?? 0)
      .addDays(duration.days ?? 0)
      .addHours(duration.hours ?? 0)
      .addMinutes(duration.minutes ?? 0)
      .addSeconds(duration.seconds ?? 0)
      .addMilliseconds(duration.milliseconds ?? 0);
  }

  subDuration(duration: DurationLike): this {
    return this.addDuration({
      days: -(duration.days ?? 0),
      hours: -(duration.hours ?? 0),
      milliseconds: -(duration.milliseconds ?? 0),
      minutes: -(duration.minutes ?? 0),
      months: -(duration.months ?? 0),
      quarters: -(duration.quarters ?? 0),
      seconds: -(duration.seconds ?? 0),
      weeks: -(duration.weeks ?? 0),
      years: -(duration.years ?? 0),
    });
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

  startOfDay(): this {
    return this.startOf("day");
  }

  endOfDay(): this {
    return this.endOf("day");
  }

  startOfMonth(): this {
    return this.startOf("month");
  }

  endOfMonth(): this {
    return this.endOf("month");
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

  diffInMonths(other: TempoInput, options?: DiffOptions): number {
    return this.diff(other, "month", options);
  }

  diffInYears(other: TempoInput, options?: DiffOptions): number {
    return this.diff(other, "year", options);
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

  toTimeString(): string {
    return this.format("HH:mm:ss");
  }

  toDateTimeString(): string {
    return this.format("YYYY-MM-DD HH:mm:ss");
  }

  toISOString(): string {
    return this.value.toISOString();
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

  private comparableValue(unit: ComparisonUnit): number {
    return unit === "millisecond"
      ? this.timestampMs
      : this.startOf(unit).timestampMs;
  }
}

export class Tempo extends TempoImmutable {
  static override now(options?: TempoOptions): Tempo {
    return new Tempo(new Date(), options);
  }

  static override parse(input: TempoInput, options?: TempoOptions): Tempo {
    return new Tempo(input, options);
  }

  static override create(components: TempoComponents): Tempo {
    return new Tempo(dateFromZonedComponents(components), {
      timeZone: components.timeZone,
    });
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

  static override parse(
    input: TempoInput,
    options?: TempoOptions,
  ): TempoMutable {
    return new TempoMutable(input, options);
  }

  static override create(components: TempoComponents): TempoMutable {
    return new TempoMutable(dateFromZonedComponents(components), {
      timeZone: components.timeZone,
    });
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

  toDuration(): DurationLike {
    return { milliseconds: this.milliseconds };
  }
}

export class TempoPeriod implements Iterable<Tempo> {
  readonly start: Tempo;
  readonly end: Tempo;
  readonly step: DurationLike;
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

      current = next;
    }
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

  create(components: TempoComponents): Tempo {
    return Tempo.create({ timeZone: this.zone, ...components });
  }

  fromTimestamp(timestamp: number, options?: TempoOptions): Tempo {
    return Tempo.fromTimestamp(timestamp, {
      timeZone: options?.timeZone ?? this.zone,
    });
  }
}

export const now = Tempo.now;
export const parse = Tempo.parse;
export const create = Tempo.create;
export const fromObject = Tempo.fromObject;
export const fromTimestamp = Tempo.fromTimestamp;
export const fromTimestampMs = Tempo.fromTimestampMs;
export const createFactory = TempoFactory.create;

import type { DurationInput, PeriodOptions, TempoInput } from "../types";
import { Tempo, TempoImmutable } from "../core";
import { TempoDuration } from "../duration";

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

  get quarters(): number {
    return this.end.diffInQuarters(this.start);
  }

  get years(): number {
    return this.end.diffInYears(this.start);
  }

  invert(): TempoInterval {
    return new TempoInterval(this.end, this.start);
  }

  absolute(): TempoInterval {
    return this.isInverted
      ? this.invert()
      : new TempoInterval(this.start, this.end);
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

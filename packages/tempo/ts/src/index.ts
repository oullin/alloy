export type TempoInput = Date | Tempo | TempoImmutable | number | string;

const millisecondsPerDay = 86_400_000;
let testNow: Date | null = null;

const asDate = (input: TempoInput): Date => {
  if (input instanceof Tempo || input instanceof TempoImmutable) {
    return input.toDate();
  }

  if (input instanceof Date) {
    return new Date(input.getTime());
  }

  const date =
    typeof input === "number" ? new Date(input * 1000) : new Date(input);

  if (Number.isNaN(date.getTime())) {
    throw new RangeError(`Invalid Tempo input: ${String(input)}`);
  }

  return date;
};

const formatDate = (date: Date): string => date.toISOString().slice(0, 10);

export class TempoImmutable {
  protected readonly value: Date;

  protected constructor(value: Date) {
    this.value = new Date(value.getTime());
  }

  static now(): TempoImmutable {
    return new TempoImmutable(testNow ?? new Date());
  }

  static parse(input: TempoInput): TempoImmutable {
    return new TempoImmutable(asDate(input));
  }

  static fromTimestamp(timestamp: number): TempoImmutable {
    return new TempoImmutable(asDate(timestamp));
  }

  static setTestNow(input: TempoInput | null): void {
    testNow = input === null ? null : asDate(input);
  }

  addDays(days: number): TempoImmutable {
    return new TempoImmutable(
      new Date(this.value.getTime() + days * millisecondsPerDay),
    );
  }

  addMonths(months: number): TempoImmutable {
    const next = this.toDate();
    next.setUTCMonth(next.getUTCMonth() + months);
    return new TempoImmutable(next);
  }

  diffInDays(other: TempoInput): number {
    return Math.trunc(
      (this.value.getTime() - asDate(other).getTime()) / millisecondsPerDay,
    );
  }

  isBefore(other: TempoInput): boolean {
    return this.value.getTime() < asDate(other).getTime();
  }

  isAfter(other: TempoInput): boolean {
    return this.value.getTime() > asDate(other).getTime();
  }

  toDate(): Date {
    return new Date(this.value.getTime());
  }

  toDateString(): string {
    return formatDate(this.value);
  }

  toISOString(): string {
    return this.value.toISOString();
  }

  toJSON(): string {
    return this.toISOString();
  }
}

export class Tempo extends TempoImmutable {
  static override now(): Tempo {
    return new Tempo(testNow ?? new Date());
  }

  static override parse(input: TempoInput): Tempo {
    return new Tempo(asDate(input));
  }

  static override fromTimestamp(timestamp: number): Tempo {
    return new Tempo(asDate(timestamp));
  }

  override addDays(days: number): Tempo {
    this.value.setTime(this.value.getTime() + days * millisecondsPerDay);
    return this;
  }

  override addMonths(months: number): Tempo {
    this.value.setUTCMonth(this.value.getUTCMonth() + months);
    return this;
  }
}

export const parse = TempoImmutable.parse;
export const fromTimestamp = TempoImmutable.fromTimestamp;
export const setTestNow = TempoImmutable.setTestNow;

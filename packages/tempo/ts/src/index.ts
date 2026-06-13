export type TempoInput =
  | Date
  | Tempo
  | TempoImmutable
  | TempoMutable
  | number
  | string;

const millisecondsPerDay = 86_400_000;

const fromNumericTimestamp = (input: number): Date => {
  const magnitude = Math.abs(input);
  const milliseconds = magnitude < 10_000_000_000 ? input * 1000 : input;

  return new Date(milliseconds);
};

const asDate = (input: TempoInput): Date => {
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

  const date =
    typeof input === "number" ? fromNumericTimestamp(input) : new Date(input);

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
    return new TempoImmutable(new Date());
  }

  static parse(input: TempoInput): TempoImmutable {
    return new TempoImmutable(asDate(input));
  }

  static fromTimestamp(timestamp: number): TempoImmutable {
    return new TempoImmutable(asDate(timestamp));
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
    return new Tempo(new Date());
  }

  static override parse(input: TempoInput): Tempo {
    return new Tempo(asDate(input));
  }

  static override fromTimestamp(timestamp: number): Tempo {
    return new Tempo(asDate(timestamp));
  }

  override addDays(days: number): Tempo {
    return new Tempo(
      new Date(this.value.getTime() + days * millisecondsPerDay),
    );
  }

  override addMonths(months: number): Tempo {
    const next = this.toDate();
    next.setUTCMonth(next.getUTCMonth() + months);
    return new Tempo(next);
  }
}

export class TempoMutable extends TempoImmutable {
  static override now(): TempoMutable {
    return new TempoMutable(new Date());
  }

  static override parse(input: TempoInput): TempoMutable {
    return new TempoMutable(asDate(input));
  }

  static override fromTimestamp(timestamp: number): TempoMutable {
    return new TempoMutable(asDate(timestamp));
  }

  override addDays(days: number): TempoMutable {
    this.value.setTime(this.value.getTime() + days * millisecondsPerDay);
    return this;
  }

  override addMonths(months: number): TempoMutable {
    this.value.setUTCMonth(this.value.getUTCMonth() + months);
    return this;
  }
}

export class TempoFactory {
  private readonly nowValue: Date | null;

  private constructor(nowValue: Date | null) {
    this.nowValue = nowValue === null ? null : new Date(nowValue.getTime());
  }

  static create(): TempoFactory {
    return new TempoFactory(null);
  }

  static withTestNow(input: TempoInput): TempoFactory {
    return new TempoFactory(asDate(input));
  }

  now(): Tempo {
    return Tempo.parse(this.nowValue ?? new Date());
  }

  immutableNow(): TempoImmutable {
    return TempoImmutable.parse(this.nowValue ?? new Date());
  }

  mutableNow(): TempoMutable {
    return TempoMutable.parse(this.nowValue ?? new Date());
  }

  parse(input: TempoInput): Tempo {
    return Tempo.parse(input);
  }
}

export const parse = Tempo.parse;
export const fromTimestamp = Tempo.fromTimestamp;
export const createFactory = TempoFactory.create;

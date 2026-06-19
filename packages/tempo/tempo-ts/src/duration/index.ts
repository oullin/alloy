import type { DurationInput, DurationLike, DurationObject, TimeUnit } from '#types';
import { assertFiniteNumber, fixedUnitMilliseconds, isoDurationPattern, millisecondsPerDay, millisecondsPerHour, millisecondsPerMinute, millisecondsPerSecond, normalizeUnit, pad } from '#calendar';

export const durationFromInput = (input: DurationInput): TempoDuration => (input instanceof TempoDuration ? input : typeof input === 'string' ? TempoDuration.parse(input) : new TempoDuration(input));

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

		const sign = match[1] === '-' ? -1 : 1;
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

	static fromJSON(input: string): TempoDuration {
		const value = JSON.parse(input) as unknown;

		if (typeof value !== 'string') {
			throw new RangeError('Tempo duration JSON must be a string');
		}

		return TempoDuration.parse(value);
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
			case 'month':
				return this.years * 12 + this.quarters * 3 + this.months + milliseconds / (millisecondsPerDay * 30);

			case 'quarter':
				return this.total('month') / 3;

			case 'year':
				return this.total('month') / 12;

			default:
				return milliseconds;
		}
	}

	isZero(): boolean {
		return Object.values(this.toObject()).every((value) => value === 0);
	}

	isPositive(): boolean {
		return !this.isZero() && this.direction() > 0;
	}

	isNegative(): boolean {
		return this.direction() < 0;
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

	toMap(): Map<keyof DurationObject, DurationObject[keyof DurationObject]> {
		return new Map(Object.entries(this.toObject()) as Array<[keyof DurationObject, DurationObject[keyof DurationObject]]>);
	}

	toArray(): [number, number, number, number, number, number, number, number, number] {
		return [this.years, this.quarters, this.months, this.weeks, this.days, this.hours, this.minutes, this.seconds, this.milliseconds];
	}

	toISOString(): string {
		if (this.isZero()) {
			return 'PT0S';
		}

		const normalized = this.normalized();
		const sign = normalized.direction() < 0 ? '-' : '';
		const value = normalized.abs();
		const dateParts = [value.years === 0 ? '' : `${value.years}Y`, value.months === 0 ? '' : `${value.months}M`, value.days === 0 ? '' : `${value.days}D`].join('');
		const secondValue = value.milliseconds === 0 ? String(value.seconds) : `${value.seconds}.${pad(value.milliseconds, 3)}`;

		const timeParts = [value.hours === 0 ? '' : `${value.hours}H`, value.minutes === 0 ? '' : `${value.minutes}M`, value.seconds === 0 && value.milliseconds === 0 ? '' : `${secondValue}S`].join(
			'',
		);

		return `${sign}P${dateParts}${timeParts === '' ? '' : `T${timeParts}`}`;
	}

	toJSON(): string {
		return this.toISOString();
	}

	toString(): string {
		return this.toISOString();
	}

	private totalMilliseconds(): number {
		return (this.weeks * 7 + this.days) * millisecondsPerDay + this.hours * millisecondsPerHour + this.minutes * millisecondsPerMinute + this.seconds * millisecondsPerSecond + this.milliseconds;
	}

	private direction(): 1 | -1 {
		const first = Object.values(this.toObject()).find((value) => value !== 0);

		return first !== undefined && first < 0 ? -1 : 1;
	}
}

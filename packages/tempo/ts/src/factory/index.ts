import { defaultTimeZone, normalizeTimeZone } from '../calendar';
import { tempoConfig } from '../config';
import { Tempo, TempoImmutable, TempoMutable } from '../core';
import { asDate, runtimeFromOptions, zoneFromInput } from '../parsing';
import { TempoRuntime } from '../runtime';

import type { TempoComponents, TempoInput, TempoOptions, TempoTranslator } from '../types';

export class TempoFactory {
	private readonly nowValue: Date | null;
	private readonly runtime: TempoRuntime;
	private readonly zone: string;

	private constructor(
		nowValue: Date | null,
		timeZone = defaultTimeZone,
		runtime = new TempoRuntime({
			fallbackLocale: tempoConfig.fallbackLocale,
			locale: tempoConfig.locale,
		}),
	) {
		this.nowValue = nowValue === null ? null : new Date(nowValue.getTime());
		this.runtime = runtime;
		this.zone = normalizeTimeZone(timeZone);
	}

	static create(options?: TempoOptions): TempoFactory {
		return new TempoFactory(null, options?.timeZone, runtimeFromOptions(undefined, options));
	}

	static withTestNow(input: TempoInput, options?: TempoOptions): TempoFactory {
		return new TempoFactory(asDate(input, options), zoneFromInput(input, options), runtimeFromOptions(input, options));
	}

	getRuntime(): TempoRuntime {
		return this.runtime;
	}

	withRuntime(runtime: TempoRuntime): TempoFactory {
		return new TempoFactory(this.nowValue, this.zone, runtime);
	}

	withTranslator(translator: TempoTranslator): TempoFactory {
		return this.withRuntime(
			this.runtime.with({
				fallbackLocale: translator.fallbackLocale,
				locale: translator.locale ?? this.runtime.locale,
				translator,
			}),
		);
	}

	private options(options?: TempoOptions): TempoOptions {
		return {
			fallbackLocale: options?.fallbackLocale,
			locale: options?.locale,
			runtime:
				options?.runtime ??
				this.runtime.with({
					fallbackLocale: options?.fallbackLocale,
					locale: options?.locale,
					translator: options?.translator,
				}),
			timeZone: options?.timeZone ?? this.zone,
			translator: options?.translator,
		};
	}

	now(): Tempo {
		return Tempo.parse(this.nowValue ?? new Date(), this.options());
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
		return TempoImmutable.parse(this.nowValue ?? new Date(), this.options());
	}

	mutableNow(): TempoMutable {
		return TempoMutable.parse(this.nowValue ?? new Date(), this.options());
	}

	parse(input: TempoInput, options?: TempoOptions): Tempo {
		return Tempo.parse(input, this.options(options));
	}

	tryParse(input: TempoInput, options?: TempoOptions): Tempo | null {
		return Tempo.tryParse(input, this.options(options));
	}

	canParse(input: TempoInput, options?: TempoOptions): boolean {
		return this.tryParse(input, options) !== null;
	}

	fromFormat(input: string, pattern: string, options?: TempoOptions): Tempo {
		return Tempo.fromFormat(input, pattern, this.options(options));
	}

	tryFromFormat(input: string, pattern: string, options?: TempoOptions): Tempo | null {
		return Tempo.tryFromFormat(input, pattern, this.options(options));
	}

	hasFormat(input: string, pattern: string, options?: TempoOptions): boolean {
		return this.tryFromFormat(input, pattern, options) !== null;
	}

	create(components: TempoComponents): Tempo {
		return Tempo.create({ timeZone: this.zone, ...components }).withRuntime(this.runtime) as Tempo;
	}

	createSafe(components: TempoComponents): Tempo {
		return Tempo.createSafe({ timeZone: this.zone, ...components }).withRuntime(this.runtime) as Tempo;
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
		return Tempo.fromTimestamp(timestamp, this.options(options));
	}

	fromTimestampMs(timestamp: number, options?: TempoOptions): Tempo {
		return Tempo.fromTimestampMs(timestamp, this.options(options));
	}
}

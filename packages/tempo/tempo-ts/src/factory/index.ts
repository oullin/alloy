import { cloneTempoPolicy, resolveTempoPolicy } from '#config';
import { Tempo, TempoImmutable } from '#core';
import { asDate, zoneFromInput } from '#parsing';
import { TempoRuntime } from '#runtime';

import type { TempoComponents, TempoInput, TempoOptions, TempoPolicy, TempoTranslator } from '#types';

export class TempoFactory {
	private readonly nowValue: Date | null;
	private readonly policy: TempoPolicy;
	private readonly runtime: TempoRuntime;
	private readonly zone: string;

	private constructor(policy: TempoPolicy, nowValue: Date | null = policy.testNow) {
		this.nowValue = nowValue === null ? null : new Date(nowValue.getTime());
		this.runtime =
			policy.runtime ??
			new TempoRuntime({
				fallbackLocale: policy.fallbackLocale,
				locale: policy.locale,
				translator: policy.translator,
			});
		this.policy = cloneTempoPolicy({
			...policy,
			fallbackLocale: this.runtime.fallbackLocale,
			locale: this.runtime.locale,
			runtime: this.runtime,
			testNow: this.nowValue,
		});
		this.zone = policy.timeZone;
	}

	static create(options?: TempoOptions): TempoFactory {
		const policy = resolveTempoPolicy(options);

		return new TempoFactory(policy, policy.testNow);
	}

	static withTestNow(input: TempoInput, options?: TempoOptions): TempoFactory {
		const policy = resolveTempoPolicy(options);
		const now = asDate(input, options, policy);

		return new TempoFactory(
			resolveTempoPolicy(
				{
					...options,
					timeZone: zoneFromInput(input, options, policy),
				},
				policy,
			),
			now,
		);
	}

	getRuntime(): TempoRuntime {
		return this.runtime;
	}

	withRuntime(runtime: TempoRuntime): TempoFactory {
		return new TempoFactory(
			resolveTempoPolicy(
				{
					runtime,
				},
				this.policy,
			),
			this.nowValue,
		);
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
		const policy = resolveTempoPolicy(options, this.policy);

		const runtime =
			options?.runtime ??
			this.runtime.with({
				fallbackLocale: options?.fallbackLocale,
				locale: options?.locale,
				translator: options?.translator,
			});

		return {
			fallbackLocale: policy.fallbackLocale,
			humanDiffOptions: policy.humanDiffOptions,
			locale: policy.locale,
			midDayAt: policy.midDayAt,
			monthsOverflow: policy.monthsOverflow,
			runtime,
			serializer: policy.serializer,
			strictMode: policy.strictMode,
			testNow: this.nowValue,
			timeZone: policy.timeZone,
			toStringFormat: policy.toStringFormat,
			translator: policy.translator,
			weekendDays: policy.weekendDays,
			yearsOverflow: policy.yearsOverflow,
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

	parse(input: TempoInput, options?: TempoOptions): Tempo {
		return Tempo.parse(input, this.options(options));
	}

	fromFormat(input: string, pattern: string, options?: TempoOptions): Tempo {
		return Tempo.fromFormat(input, pattern, this.options(options));
	}

	create(components: TempoComponents): Tempo {
		return Tempo.create({ timeZone: this.zone, ...components }, this.options());
	}

	createNormalized(components: TempoComponents): Tempo {
		return Tempo.createNormalized({ timeZone: this.zone, ...components }, this.options());
	}

	fromDate(year: number, month = 1, day = 1): Tempo {
		return this.create({ day, month, year });
	}

	fromTime(hour = 0, minute = 0, second = 0, millisecond = 0): Tempo {
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

import type { TempoTranslationValue, TempoTranslator } from '../types';

export class TempoRuntime {
	readonly fallbackLocale: string;
	readonly locale: string;
	readonly translator: TempoTranslator | null;

	constructor(
		options: {
			readonly fallbackLocale?: string;
			readonly locale?: string;
			readonly translator?: TempoTranslator | null;
		} = {},
	) {
		this.fallbackLocale = options.fallbackLocale ?? 'en-US';
		this.locale = options.locale ?? options.translator?.locale ?? 'en-US';
		this.translator = options.translator ?? null;
	}

	with(
		options: {
			readonly fallbackLocale?: string;
			readonly locale?: string;
			readonly translator?: TempoTranslator | null;
		} = {},
	): TempoRuntime {
		return new TempoRuntime({
			fallbackLocale: options.fallbackLocale ?? this.fallbackLocale,
			locale: options.locale ?? this.locale,
			translator: options.translator === undefined ? this.translator : options.translator,
		});
	}

	hasTranslator(): boolean {
		return this.translator !== null;
	}

	translatorState(): TempoTranslator {
		return (
			this.translator ?? {
				fallbackLocale: this.fallbackLocale,
				locale: this.locale,
			}
		);
	}

	getMessage(key: string): TempoTranslationValue {
		const translated = this.translator?.getMessage?.(key);

		if (translated !== undefined && translated !== null) {
			return translated;
		}

		switch (key) {
			case 'day_of_first_week_of_year':
				return 4;

			case 'first_day_of_week':
				return 1;

			case 'locale':
				return this.locale;

			default:
				return null;
		}
	}

	translate(key: string, replacements: Record<string, string> = {}): TempoTranslationValue {
		const translated = this.translator?.translate?.(key, replacements);

		const message = translated === undefined || translated === null ? this.getMessage(key) : translated;

		return typeof message === 'string' ? replaceTranslationTokens(message, replacements) : message;
	}
}

export const createTempoRuntime = (options?: ConstructorParameters<typeof TempoRuntime>[0]): TempoRuntime => new TempoRuntime(options);

export const replaceTranslationTokens = (message: string, replacements: Record<string, string>): string =>
	Object.entries(replacements).reduce((result, [key, value]) => result.replaceAll(`:${key}`, value), message);

import type { Amount } from '#money/calculator';
import { ERR_CURRENCY_CONVERSION_NOT_FOUND, ERR_INVALID_EXCHANGE_RATE } from '#money/errors';
import { roundAwayFromZero } from '#money/internal/rounding';

export class ExchangeRates {
	private readonly rates = new Map<string, Map<string, number>>();

	public static create(): ExchangeRates {
		return new ExchangeRates();
	}

	public isValid(): boolean {
		return this.rates.size > 0;
	}

	public isInvalid(): boolean {
		return !this.isValid();
	}

	public addRate(baseCurrency: string, counterCurrency: string, rate: number): this {
		if (rate <= 0 || !Number.isFinite(rate)) {
			throw ERR_INVALID_EXCHANGE_RATE;
		}

		const base = baseCurrency.toUpperCase();
		const counter = counterCurrency.toUpperCase();
		const inner = this.rates.get(base) ?? new Map<string, number>();

		inner.set(counter, rate);
		this.rates.set(base, inner);

		return this;
	}

	public getRate(baseCurrency: string, counterCurrency: string): number {
		const base = baseCurrency.toUpperCase();
		const counter = counterCurrency.toUpperCase();

		if (base === counter) {
			return 1;
		}

		const direct = this.rates.get(base)?.get(counter);

		if (direct !== undefined) {
			return direct;
		}

		const inverse = this.rates.get(counter)?.get(base);

		if (inverse !== undefined) {
			return 1 / inverse;
		}

		throw ERR_CURRENCY_CONVERSION_NOT_FOUND;
	}

	public convertAmount(amount: Amount, fromCurrencyCode: string, fromFraction: number, toCurrencyCode: string, toFraction: number): Amount {
		if (fromCurrencyCode.toUpperCase() === toCurrencyCode.toUpperCase()) {
			return amount;
		}

		return this.convertAmountWithRate(amount, fromFraction, toFraction, this.getRate(fromCurrencyCode, toCurrencyCode));
	}

	public convertAmountWithRate(amount: Amount, fromFraction: number, toFraction: number, rate: number): Amount {
		if (rate <= 0 || !Number.isFinite(rate)) {
			throw ERR_INVALID_EXCHANGE_RATE;
		}

		const majorUnits = Number(amount) / 10 ** fromFraction;
		const convertedMajorUnits = majorUnits * rate;

		return roundAwayFromZero(convertedMajorUnits * 10 ** toFraction);
	}
}

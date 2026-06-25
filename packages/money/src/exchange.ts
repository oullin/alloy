import type { Amount } from '#money/calculator';
import { ERR_CURRENCY_CONVERSION_NOT_FOUND, ERR_INVALID_EXCHANGE_RATE, ERR_NO_CONVERTER_PROVIDED } from '#money/errors';

const roundAwayFromZero = (value: number): bigint => {
	if (!Number.isFinite(value)) {
		throw ERR_INVALID_EXCHANGE_RATE;
	}

	return BigInt(Math.sign(value) * Math.round(Math.abs(value)));
};

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

export class ExchangeConverter {
	public constructor(private readonly exchange: ExchangeRates) {
		if (exchange === null || exchange === undefined) {
			throw ERR_NO_CONVERTER_PROVIDED;
		}
	}

	public static create(exchange: ExchangeRates): ExchangeConverter {
		return new ExchangeConverter(exchange);
	}

	public getExchange(): ExchangeRates {
		if (this.exchange === null || this.exchange === undefined) {
			throw ERR_INVALID_EXCHANGE_RATE;
		}

		return this.exchange;
	}
}

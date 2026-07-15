import { MAX_INT64, MIN_INT64, type Amount } from '#money/calculator';
import { ERR_CURRENCY_CONVERSION_NOT_FOUND, ERR_INVALID_EXCHANGE_RATE, ERR_OVERFLOW } from '#money/errors';

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

	/**
	 * Converts an amount using a scale-12 rate representation, preserving exact
	 * precision across the full int64 amount range (including values above 2^53).
	 * Rounds half away from zero and throws `ERR_OVERFLOW` when the result leaves
	 * the int64 range.
	 */
	public convertAmountWithRate(amount: Amount, fromFraction: number, toFraction: number, rate: number): Amount {
		if (rate <= 0 || !Number.isFinite(rate)) {
			throw ERR_INVALID_EXCHANGE_RATE;
		}

		const RATE_SCALE = 12;
		const scaledRate = Math.round(rate * 10 ** RATE_SCALE);

		if (!Number.isFinite(scaledRate) || scaledRate > Number(MAX_INT64)) {
			throw ERR_OVERFLOW;
		}

		const rateScaled = BigInt(scaledRate);
		const numerator = amount * rateScaled * 10n ** BigInt(toFraction);
		const denominator = 10n ** BigInt(RATE_SCALE) * 10n ** BigInt(fromFraction);
		const negative = numerator < 0n;
		const absoluteNumerator = negative ? -numerator : numerator;
		let quotient = absoluteNumerator / denominator;
		const remainder = absoluteNumerator % denominator;

		if (remainder * 2n >= denominator) {
			quotient += 1n;
		}

		const result = negative ? -quotient : quotient;

		if (result < MIN_INT64 || result > MAX_INT64) {
			throw ERR_OVERFLOW;
		}

		return result;
	}
}

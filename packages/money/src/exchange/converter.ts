import { ERR_INVALID_EXCHANGE_RATE, ERR_NO_CONVERTER_PROVIDED } from '#money/errors';
import type { ExchangeRates } from '#money/exchange/rates';

/**
 * Guarded holder for an {@link ExchangeRates} table, ensuring conversions are
 * only attempted with a valid rate source.
 */
export class ExchangeConverter {
	/** @throws MoneyError `ERR_NO_CONVERTER_PROVIDED` when no rates are given. */
	public constructor(private readonly exchange: ExchangeRates) {
		if (exchange === null || exchange === undefined) {
			throw ERR_NO_CONVERTER_PROVIDED;
		}
	}

	public static create(exchange: ExchangeRates): ExchangeConverter {
		return new ExchangeConverter(exchange);
	}

	/**
	 * Returns the underlying exchange-rate table.
	 *
	 * @throws MoneyError `ERR_INVALID_EXCHANGE_RATE` when the table is missing.
	 */
	public getExchange(): ExchangeRates {
		if (this.exchange === null || this.exchange === undefined) {
			throw ERR_INVALID_EXCHANGE_RATE;
		}

		return this.exchange;
	}
}

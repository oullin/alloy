import { ERR_INVALID_EXCHANGE_RATE, ERR_NO_CONVERTER_PROVIDED } from '#money/errors';
import type { ExchangeRates } from '#money/exchange/rates';

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

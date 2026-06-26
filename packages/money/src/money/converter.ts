import type { CurrencyManager } from '#money/currency/manager';
import { ERR_CURRENCY_NOT_FOUND, ERR_INVALID_EXCHANGE_RATE, ERR_NO_CURRENCY_MANAGER } from '#money/errors';
import type { ExchangeRates } from '#money/exchange/rates';
import { MoneyValueBase, requireMoney, type MoneyValue } from '#money/money/core';

export class MoneyConverter {
	public constructor(
		private readonly currencies: CurrencyManager,
		private readonly exchange: ExchangeRates,
	) {
		if (currencies === null || currencies === undefined) {
			throw ERR_NO_CURRENCY_MANAGER;
		}

		if (exchange === null || exchange === undefined || exchange.isInvalid()) {
			throw ERR_INVALID_EXCHANGE_RATE;
		}
	}

	public static create(currencies: CurrencyManager, exchange: ExchangeRates): MoneyConverter {
		return new MoneyConverter(currencies, exchange);
	}

	public convert(money: MoneyValue, toCurrency: string): MoneyValue {
		const source = requireMoney(money);
		const target = this.currencies.findByCode(toCurrency);

		if (target === null) {
			throw ERR_CURRENCY_NOT_FOUND;
		}

		return new MoneyValueBase(this.exchange.convertAmount(source.amount(), source.currency().code, source.currency().fraction, target.code, target.fraction), target) as MoneyValue;
	}

	public convertWithRate(money: MoneyValue, toCurrency: string, rate: number): MoneyValue {
		const source = requireMoney(money);
		const target = this.currencies.findByCode(toCurrency);

		if (target === null) {
			throw ERR_CURRENCY_NOT_FOUND;
		}

		return new MoneyValueBase(this.exchange.convertAmountWithRate(source.amount(), source.currency().fraction, target.fraction, rate), target) as MoneyValue;
	}
}

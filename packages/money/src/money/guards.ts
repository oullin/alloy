import { ERR_NO_CURRENCY_INSTANCE, ERR_NO_MONEY_PROVIDED } from '#money/errors';
import type { MoneyValue } from '#money/money/value';

export const requireMoney = (money: MoneyValue | null | undefined): MoneyValue => {
	if (money === null || money === undefined) {
		throw ERR_NO_MONEY_PROVIDED;
	}

	if (money.currency() === null) {
		throw ERR_NO_CURRENCY_INSTANCE;
	}

	return money;
};

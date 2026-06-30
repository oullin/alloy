import { ERR_INVALID_AGGREGATOR_PROVIDER, ERR_NO_MONEY_PROVIDED } from '#money/errors';
import type { MoneyManager, MoneyValue } from '#money/money/core';

export class MoneyAggregator {
	public constructor(private readonly manager: MoneyManager | null) {}

	public static create(manager: MoneyManager): MoneyAggregator {
		return new MoneyAggregator(manager);
	}

	public sum(...items: MoneyValue[]): MoneyValue {
		const manager = this.requireManager();

		if (items.length === 0) {
			throw ERR_NO_MONEY_PROVIDED;
		}

		return manager.add(items[0] as MoneyValue, ...items.slice(1));
	}

	public min(...items: MoneyValue[]): MoneyValue {
		this.requireManager();

		if (items.length === 0) {
			throw ERR_NO_MONEY_PROVIDED;
		}

		let money = items[0] as MoneyValue;

		for (const item of items.slice(1)) {
			money.assertSameCurrency(item);

			if (item.amount() < money.amount()) {
				money = item;
			}
		}

		return money;
	}

	public max(...items: MoneyValue[]): MoneyValue {
		this.requireManager();

		if (items.length === 0) {
			throw ERR_NO_MONEY_PROVIDED;
		}

		let money = items[0] as MoneyValue;

		for (const item of items.slice(1)) {
			money.assertSameCurrency(item);

			if (item.amount() > money.amount()) {
				money = item;
			}
		}

		return money;
	}

	public avg(...items: MoneyValue[]): MoneyValue {
		const manager = this.requireManager();

		if (items.length === 0) {
			throw ERR_NO_MONEY_PROVIDED;
		}

		const sum = this.sum(...items);

		return manager.create(sum.amount() / BigInt(items.length), sum.currency().code);
	}

	private requireManager(): MoneyManager {
		if (this.manager === null) {
			throw ERR_INVALID_AGGREGATOR_PROVIDER;
		}

		return this.manager;
	}
}

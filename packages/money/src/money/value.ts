import type { Amount } from '#money/calculator';
import { CURRENCY_CODES, type CurrencyCode } from '#money/currency-data';
import { CurrencyDefinition } from '#money/currency/definition';
import { ERR_CURRENCY_MISMATCH, ERR_NO_CURRENCY_INSTANCE } from '#money/errors';
import { getDbMoneyValueSeparator, setDbMoneyValueSeparator } from '#money/money/db-separator';
import { requireMoney } from '#money/money/guards';
import { MoneyManager } from '#money/money/manager';

export interface MoneyJsonValue {
	readonly amount: string;
	readonly currency: string;
}

export class MoneyValueBase {
	public constructor(
		private readonly valueAmount: Amount,
		private readonly valueCurrency: CurrencyDefinition | null,
	) {}

	public static fromCurrency(amount: Amount, code: string, manager: MoneyManager = MoneyManager.default()): MoneyValue {
		return manager.create(amount, code);
	}

	public static fromDbValue(input: string | Uint8Array): MoneyValue {
		const separator = getDbMoneyValueSeparator();
		const value = typeof input === 'string' ? input : new TextDecoder().decode(input);
		const parts = value.split(separator);

		if (parts.length !== 2 || parts[0] === '' || parts[1] === '') {
			throw new Error(`${JSON.stringify(input)} is not valid to scan into MoneyValue; update your query to return an "amount${separator}currency_code" pair`);
		}

		if (!/^-?\d+$/u.test(parts[0] ?? '')) {
			throw new Error(`scanning ${JSON.stringify(parts[0])} into an Amount`);
		}

		const currency = CurrencyDefinition.fromDbValue(parts[1] ?? '');

		return new MoneyValueBase(BigInt(parts[0] ?? '0'), currency) as MoneyValue;
	}

	public static getDbMoneyValueSeparator(): string {
		return getDbMoneyValueSeparator();
	}

	public static setDbMoneyValueSeparator(separator: string): void {
		setDbMoneyValueSeparator(separator);
	}

	public amount(): Amount {
		requireMoney(this as MoneyValue);

		return this.valueAmount;
	}

	public currency(): CurrencyDefinition {
		if (this.valueCurrency === null) {
			throw ERR_NO_CURRENCY_INSTANCE;
		}

		return this.valueCurrency;
	}

	public assertSameCurrency(other: MoneyValue | null | undefined): void {
		const value = requireMoney(this as MoneyValue);
		const otherValue = requireMoney(other);

		if (!value.currency().equals(otherValue.currency())) {
			throw ERR_CURRENCY_MISMATCH;
		}
	}

	public sameCurrency(other: MoneyValue | null | undefined): boolean {
		const value = requireMoney(this as MoneyValue);
		const otherValue = requireMoney(other);

		return value.currency().equals(otherValue.currency());
	}

	public compareAmount(other: MoneyValue): number {
		this.assertSameCurrency(other);

		if (this.valueAmount > other.amount()) {
			return 1;
		}

		if (this.valueAmount < other.amount()) {
			return -1;
		}

		return 0;
	}

	public equals(other: MoneyValue): boolean {
		return this.compareAmount(other) === 0;
	}

	public greaterThan(other: MoneyValue): boolean {
		return this.compareAmount(other) === 1;
	}

	public greaterThanOrEqual(other: MoneyValue): boolean {
		return this.compareAmount(other) >= 0;
	}

	public lessThan(other: MoneyValue): boolean {
		return this.compareAmount(other) === -1;
	}

	public lessThanOrEqual(other: MoneyValue): boolean {
		return this.compareAmount(other) <= 0;
	}

	public isZero(): boolean {
		requireMoney(this as MoneyValue);

		return this.valueAmount === 0n;
	}

	public isPositive(): boolean {
		requireMoney(this as MoneyValue);

		return this.valueAmount > 0n;
	}

	public isNegative(): boolean {
		requireMoney(this as MoneyValue);

		return this.valueAmount < 0n;
	}

	public display(): string {
		return this.currency().formatter().format(this.valueAmount);
	}

	public asMajorUnits(): number {
		return this.currency().formatter().toMajorUnits(this.valueAmount);
	}

	public compare(other: MoneyValue): number {
		return this.compareAmount(other);
	}

	public dbValue(): string {
		return `${this.valueAmount.toString()}${getDbMoneyValueSeparator()}${this.currency().code}`;
	}

	public toJSON(): MoneyJsonValue {
		return {
			amount: this.valueAmount.toString(),
			currency: this.currency().code,
		};
	}
}

type MoneyValueFactories = {
	[Code in CurrencyCode as `from${Code}`]: (amount: Amount) => MoneyValue;
};

const moneyValueFactories = Object.fromEntries(
	CURRENCY_CODES.map((code) => [`from${code}`, (amount: Amount): MoneyValue => MoneyManager.default().create(amount, code)]),
) as MoneyValueFactories;

export const MoneyValue = Object.assign(MoneyValueBase, moneyValueFactories) as typeof MoneyValueBase & MoneyValueFactories;

export type MoneyValue = MoneyValueBase;

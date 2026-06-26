import { MoneyCalculator, type Amount } from '#money/calculator';
import { CURRENCY_CODES, type CurrencyCode } from '#money/currency-data';
import { CurrencyDefinition } from '#money/currency/model';
import { CurrencyManager } from '#money/currency/manager';
import { roundAwayFromZero } from '#money/internal/rounding';
import { MoneyParser } from '#money/parser/money-parser';

import {
	ERR_CURRENCY_MISMATCH,
	ERR_EMPTY_AMOUNT_STRING,
	ERR_INVALID_SPLIT,
	ERR_NEGATIVE_RATIOS,
	ERR_NO_CURRENCY_INSTANCE,
	ERR_NO_CURRENCY_MANAGER,
	ERR_NO_MONEY_PROVIDED,
	ERR_NO_MULTIPLIERS_PROVIDED,
	ERR_NO_RATIOS_PROVIDED,
	ERR_RATIOS_EXCEED_MAX_INT,
} from '#money/errors';

export type { Amount };

export interface MoneyJsonValue {
	readonly amount: string;
	readonly currency: string;
}

let dbMoneyValueSeparator = '|';

const getDbMoneyValueSeparator = (): string => dbMoneyValueSeparator;

const setDbMoneyValueSeparator = (separator: string): void => {
	if (separator.trim() === '') {
		throw new Error(`separator [${separator}] cannot be empty`);
	}

	dbMoneyValueSeparator = separator;
};

export const requireMoney = (money: MoneyValue | null | undefined): MoneyValue => {
	if (money === null || money === undefined) {
		throw ERR_NO_MONEY_PROVIDED;
	}

	if (money.currency() === null) {
		throw ERR_NO_CURRENCY_INSTANCE;
	}

	return money;
};

const ensureAmountString = (amount: string): string => {
	const trimmed = amount.trim();

	if (trimmed === '') {
		throw ERR_EMPTY_AMOUNT_STRING;
	}

	return trimmed;
};

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

export class MoneyManager {
	private readonly currencyManager: CurrencyManager;
	private readonly calculator: MoneyCalculator;
	private readonly parser: MoneyParser;

	public constructor(currencyManager: CurrencyManager = CurrencyManager.default(), calculator: MoneyCalculator = MoneyCalculator.create(), parser: MoneyParser = MoneyParser.create()) {
		this.currencyManager = currencyManager;
		this.calculator = calculator;
		this.parser = parser;
	}

	public static default(): MoneyManager {
		return new MoneyManager();
	}

	public static withCurrencyManager(currencyManager: CurrencyManager | null): MoneyManager {
		if (currencyManager === null) {
			throw ERR_NO_CURRENCY_MANAGER;
		}

		return new MoneyManager(currencyManager);
	}

	public create(amount: Amount, code: string): MoneyValue {
		return new MoneyValueBase(amount, this.currencyManager.resolve(code)) as MoneyValue;
	}

	public createFromFloat(amount: number, code: string): MoneyValue {
		const currency = this.currencyManager.resolve(code);
		const scaled = amount * 10 ** currency.fraction;

		return this.create(roundAwayFromZero(scaled), code);
	}

	public createFromString(amount: string, code: string): MoneyValue {
		const trimmed = ensureAmountString(amount);
		const currency = this.currencyManager.resolve(code);
		const signed = this.parser.parseStringSign(trimmed);
		const value = this.parser.parseAmountString(signed.amount, currency.fraction, signed.negative);

		return this.create(value, code);
	}

	public getCurrencyManager(): CurrencyManager {
		return this.currencyManager;
	}

	public add(money: MoneyValue, ...items: MoneyValue[]): MoneyValue {
		const base = requireMoney(money);

		let amount = base.amount();

		for (const item of items) {
			base.assertSameCurrency(item);
			amount += item.amount();
		}

		return this.create(amount, base.currency().code);
	}

	public subtract(money: MoneyValue, ...items: MoneyValue[]): MoneyValue {
		const base = requireMoney(money);

		let amount = base.amount();

		for (const item of items) {
			base.assertSameCurrency(item);
			amount -= item.amount();
		}

		return this.create(amount, base.currency().code);
	}

	public multiply(money: MoneyValue, ...values: bigint[]): MoneyValue {
		const base = requireMoney(money);

		if (values.length === 0) {
			throw ERR_NO_MULTIPLIERS_PROVIDED;
		}

		return this.create(this.calculator.safeMultiply(base.amount(), ...values), base.currency().code);
	}

	public absolute(money: MoneyValue): MoneyValue {
		const base = requireMoney(money);

		return this.create(this.calculator.absolute(base.amount()), base.currency().code);
	}

	public negative(money: MoneyValue): MoneyValue {
		const base = requireMoney(money);

		return this.create(-base.amount(), base.currency().code);
	}

	public round(money: MoneyValue): MoneyValue {
		const base = requireMoney(money);

		return this.create(this.calculator.round(base.amount(), base.currency().fraction), base.currency().code);
	}

	public split(money: MoneyValue, count: number): MoneyValue[] {
		const base = requireMoney(money);

		if (count <= 0) {
			throw ERR_INVALID_SPLIT;
		}

		const divisor = BigInt(count);
		const quotient = this.calculator.divide(base.amount(), divisor);
		const remainder = this.calculator.modulus(base.amount(), divisor);
		const parts = Array.from({ length: count }, () => this.create(quotient, base.currency().code));
		const increment = base.amount() < 0n ? -1n : 1n;

		for (let index = 0; index < Number(this.calculator.absolute(remainder)); index++) {
			parts[index] = this.create((parts[index]?.amount() ?? 0n) + increment, base.currency().code);
		}

		return parts;
	}

	public allocate(money: MoneyValue, ...ratios: number[]): MoneyValue[] {
		const base = requireMoney(money);

		if (ratios.length === 0) {
			throw ERR_NO_RATIOS_PROVIDED;
		}

		let sum = 0n;

		for (const ratio of ratios) {
			if (ratio < 0) {
				throw ERR_NEGATIVE_RATIOS;
			}

			sum += BigInt(ratio);

			if (sum > BigInt(Number.MAX_SAFE_INTEGER)) {
				throw ERR_RATIOS_EXCEED_MAX_INT;
			}
		}

		let total = 0n;

		const parts = ratios.map((ratio) => {
			const amount = this.calculator.allocate(base.amount(), BigInt(ratio), sum);

			total = this.calculator.add(total, amount);

			return this.create(amount, base.currency().code);
		});

		if (sum === 0n) {
			return parts;
		}

		let leftover = this.calculator.subtract(base.amount(), total);

		const increment = leftover < 0n ? -1n : 1n;

		for (let index = 0; leftover !== 0n && index < parts.length; index++) {
			parts[index] = this.create((parts[index]?.amount() ?? 0n) + increment, base.currency().code);
			leftover = this.calculator.subtract(leftover, increment);
		}

		return parts;
	}
}

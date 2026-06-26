import { MoneyCalculator, type Amount } from '#money/calculator';
import { CurrencyManager } from '#money/currency/manager';
import { ERR_INVALID_SPLIT, ERR_NEGATIVE_RATIOS, ERR_NO_CURRENCY_MANAGER, ERR_NO_MULTIPLIERS_PROVIDED, ERR_NO_RATIOS_PROVIDED, ERR_RATIOS_EXCEED_MAX_INT } from '#money/errors';
import { roundAwayFromZero } from '#money/internal/rounding';
import { ensureAmountString } from '#money/money/amount-string';
import { requireMoney } from '#money/money/guards';
import { MoneyValueBase, type MoneyValue } from '#money/money/value';
import { MoneyParser } from '#money/parser/money-parser';

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

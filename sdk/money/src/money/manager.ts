import { MoneyCalculator, type Amount } from '#money/calculator';
import { CurrencyManager } from '#money/currency/manager';
import { ERR_INVALID_SPLIT, ERR_NEGATIVE_RATIOS, ERR_NO_CURRENCY_MANAGER, ERR_NO_MULTIPLIERS_PROVIDED, ERR_NO_RATIOS_PROVIDED, ERR_RATIOS_EXCEED_MAX_INT } from '#money/errors';
import { roundAwayFromZero } from '#money/internal/rounding';
import { ensureAmountString } from '#money/money/amount-string';
import { requireMoney } from '#money/money/guards';
import { MoneyValueBase, type MoneyValue } from '#money/money/value';
import { MoneyParser } from '#money/parser/money-parser';

/**
 * Factory and arithmetic hub for {@link MoneyValue} instances. Combines a
 * currency manager (code resolution), a calculator (overflow-safe bigint
 * math), and a parser (string amounts) behind a single entry point.
 */
export class MoneyManager {
	private readonly currencyManager: CurrencyManager;
	private readonly calculator: MoneyCalculator;
	private readonly parser: MoneyParser;

	public constructor(currencyManager: CurrencyManager = CurrencyManager.default(), calculator: MoneyCalculator = MoneyCalculator.create(), parser: MoneyParser = MoneyParser.create()) {
		this.currencyManager = currencyManager;
		this.calculator = calculator;
		this.parser = parser;
	}

	/** Returns a manager wired with the default currency data, calculator, and parser. */
	public static default(): MoneyManager {
		return new MoneyManager();
	}

	/**
	 * Returns a manager backed by the given currency manager.
	 *
	 * @throws MoneyError `ERR_NO_CURRENCY_MANAGER` when the manager is null.
	 */
	public static withCurrencyManager(currencyManager: CurrencyManager | null): MoneyManager {
		if (currencyManager === null) {
			throw ERR_NO_CURRENCY_MANAGER;
		}

		return new MoneyManager(currencyManager);
	}

	/** Creates a money value from a minor-unit amount and a currency code. */
	public create(amount: Amount, code: string): MoneyValue {
		return new MoneyValueBase(amount, this.currencyManager.resolve(code)) as MoneyValue;
	}

	/** Creates a money value from a major-unit float, rounding half away from zero. */
	public createFromFloat(amount: number, code: string): MoneyValue {
		const currency = this.currencyManager.resolve(code);
		const scaled = amount * 10 ** currency.fraction;

		return this.create(roundAwayFromZero(scaled), code);
	}

	/**
	 * Creates a money value from a decimal amount string such as `"-12.34"`.
	 *
	 * @throws MoneyError `ERR_EMPTY_AMOUNT_STRING`, `ERR_INVALID_AMOUNT`, `ERR_INVALID_AMOUNT_MULTIPLE`, or `ERR_INVALID_AMOUNT_FRACTION` for malformed input.
	 */
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

	/**
	 * Returns the sum of the given money values.
	 *
	 * @throws MoneyError `ERR_CURRENCY_MISMATCH` when the currencies differ, `ERR_OVERFLOW` when the result leaves the int64 range.
	 */
	public add(money: MoneyValue, ...items: MoneyValue[]): MoneyValue {
		const base = requireMoney(money);

		let amount = base.amount();

		for (const item of items) {
			base.assertSameCurrency(item);
			amount = this.calculator.add(amount, item.amount());
		}

		return this.create(amount, base.currency().code);
	}

	/**
	 * Subtracts the given money values from the first one.
	 *
	 * @throws MoneyError `ERR_CURRENCY_MISMATCH` when the currencies differ, `ERR_OVERFLOW` when the result leaves the int64 range.
	 */
	public subtract(money: MoneyValue, ...items: MoneyValue[]): MoneyValue {
		const base = requireMoney(money);

		let amount = base.amount();

		for (const item of items) {
			base.assertSameCurrency(item);
			amount = this.calculator.subtract(amount, item.amount());
		}

		return this.create(amount, base.currency().code);
	}

	/**
	 * Multiplies a money value by one or more bigint factors.
	 *
	 * @throws MoneyError `ERR_NO_MULTIPLIERS_PROVIDED` when no factors are given, `ERR_OVERFLOW` when the result leaves the int64 range.
	 */
	public multiply(money: MoneyValue, ...values: bigint[]): MoneyValue {
		const base = requireMoney(money);

		if (values.length === 0) {
			throw ERR_NO_MULTIPLIERS_PROVIDED;
		}

		return this.create(this.calculator.safeMultiply(base.amount(), ...values), base.currency().code);
	}

	/** Returns the absolute value of the given money. */
	public absolute(money: MoneyValue): MoneyValue {
		const base = requireMoney(money);

		return this.create(this.calculator.absolute(base.amount()), base.currency().code);
	}

	/** Returns the given money with its sign flipped. */
	public negative(money: MoneyValue): MoneyValue {
		const base = requireMoney(money);

		return this.create(-base.amount(), base.currency().code);
	}

	/** Rounds the amount to the currency's fraction, half away from zero. */
	public round(money: MoneyValue): MoneyValue {
		const base = requireMoney(money);

		return this.create(this.calculator.round(base.amount(), base.currency().fraction), base.currency().code);
	}

	/**
	 * Splits a money value into `count` parts, distributing the remainder one
	 * minor unit at a time from the first part.
	 *
	 * @throws MoneyError `ERR_INVALID_SPLIT` when `count` is not positive.
	 */
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

	/**
	 * Allocates a money value proportionally across the given ratios, spreading
	 * any leftover minor units from the first part.
	 *
	 * @throws MoneyError `ERR_NO_RATIOS_PROVIDED`, `ERR_NEGATIVE_RATIOS`, or `ERR_RATIOS_EXCEED_MAX_INT` for invalid ratios.
	 */
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

import { ERR_OVERFLOW } from '#money/errors';

export type Amount = bigint;

export const MIN_INT64 = -(2n ** 63n);

export const MAX_INT64 = 2n ** 63n - 1n;

const inInt64Range = (value: bigint): boolean => value >= MIN_INT64 && value <= MAX_INT64;

/**
 * Overflow-safe bigint arithmetic for monetary amounts. Every operation that
 * could leave the signed int64 range throws `ERR_OVERFLOW` instead of
 * silently wrapping or losing precision.
 */
export class MoneyCalculator {
	public static create(): MoneyCalculator {
		return new MoneyCalculator();
	}

	/**
	 * Adds two amounts.
	 *
	 * @throws MoneyError `ERR_OVERFLOW` when the result leaves the int64 range.
	 */
	public add(a: Amount, b: Amount): Amount {
		return MoneyCalculator.safeAdd(a, b);
	}

	/**
	 * Subtracts `b` from `a`.
	 *
	 * @throws MoneyError `ERR_OVERFLOW` when the result leaves the int64 range.
	 */
	public subtract(a: Amount, b: Amount): Amount {
		return MoneyCalculator.safeSubtract(a, b);
	}

	/**
	 * Multiplies an amount by a factor.
	 *
	 * @throws MoneyError `ERR_OVERFLOW` when the result leaves the int64 range.
	 */
	public multiply(amount: Amount, seed: bigint): Amount {
		return MoneyCalculator.ration(amount, seed);
	}

	public safeMultiply(initial: bigint, ...multipliers: bigint[]): bigint {
		return MoneyCalculator.safeMultiply(initial, ...multipliers);
	}

	/** Divides an amount, truncating toward zero; division by zero yields 0n. */
	public divide(amount: Amount, seed: bigint): Amount {
		if (seed === 0n) {
			return 0n;
		}

		return amount / seed;
	}

	/** Returns the remainder of the division; modulus by zero yields 0n. */
	public modulus(amount: Amount, seed: bigint): Amount {
		if (seed === 0n) {
			return 0n;
		}

		return amount % seed;
	}

	/** Returns the portion of `amount` represented by `ration` out of `scale`. */
	public allocate(amount: Amount, ration: bigint, scale: bigint): Amount {
		if (amount === 0n || scale === 0n) {
			return 0n;
		}

		return MoneyCalculator.ration(amount, ration) / scale;
	}

	public absolute(amount: Amount): Amount {
		if (amount < 0n) {
			return -amount;
		}

		return amount;
	}

	public negative(amount: Amount): Amount {
		if (amount > 0n) {
			return -amount;
		}

		return amount;
	}

	/** Rounds an amount to a power-of-ten boundary, half away from zero. */
	public round(amount: Amount, exponent: number): Amount {
		if (amount === 0n || exponent <= 0 || exponent > 18) {
			return amount;
		}

		let absolute = this.absolute(amount);

		const reminder = 10n ** BigInt(exponent);
		const module = absolute % reminder;

		if (module > reminder / 2n) {
			absolute += reminder;
		}

		absolute = (absolute / reminder) * reminder;

		return amount < 0n ? -absolute : absolute;
	}

	public static safeAdd(a: Amount, b: Amount): Amount {
		const result = a + b;

		if (!inInt64Range(result)) {
			throw ERR_OVERFLOW;
		}

		return result;
	}

	public static safeSubtract(a: Amount, b: Amount): Amount {
		const result = a - b;

		if (!inInt64Range(result)) {
			throw ERR_OVERFLOW;
		}

		return result;
	}

	public static ration(amount: Amount, ration: bigint): bigint {
		if (ration === 0n || amount === 0n) {
			return 0n;
		}

		const result = amount * ration;

		if (!inInt64Range(result)) {
			throw ERR_OVERFLOW;
		}

		return result;
	}

	/**
	 * Multiplies an amount by each factor in turn.
	 *
	 * @throws MoneyError `ERR_OVERFLOW` when an intermediate result leaves the int64 range.
	 */
	public static safeMultiply(initial: bigint, ...multipliers: bigint[]): bigint {
		let result = initial;

		for (const value of multipliers) {
			const next = result * value;

			if (!inInt64Range(next)) {
				throw ERR_OVERFLOW;
			}

			result = next;
		}

		return result;
	}
}

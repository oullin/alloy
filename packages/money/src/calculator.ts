import { ERR_OVERFLOW } from '#money/errors';

export type Amount = bigint;

export const MIN_INT64 = -(2n ** 63n);

export const MAX_INT64 = 2n ** 63n - 1n;

const inInt64Range = (value: bigint): boolean => value >= MIN_INT64 && value <= MAX_INT64;

export class MoneyCalculator {
	public static create(): MoneyCalculator {
		return new MoneyCalculator();
	}

	public add(a: Amount, b: Amount): Amount {
		return MoneyCalculator.safeAdd(a, b);
	}

	public subtract(a: Amount, b: Amount): Amount {
		return MoneyCalculator.safeSubtract(a, b);
	}

	public multiply(amount: Amount, seed: bigint): Amount {
		return MoneyCalculator.ration(amount, seed);
	}

	public safeMultiply(initial: bigint, ...multipliers: bigint[]): bigint {
		return MoneyCalculator.safeMultiply(initial, ...multipliers);
	}

	public divide(amount: Amount, seed: bigint): Amount {
		if (seed === 0n) {
			return 0n;
		}

		return amount / seed;
	}

	public modulus(amount: Amount, seed: bigint): Amount {
		if (seed === 0n) {
			return 0n;
		}

		return amount % seed;
	}

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

		return inInt64Range(result) ? result : 0n;
	}

	public static safeSubtract(a: Amount, b: Amount): Amount {
		const result = a - b;

		return inInt64Range(result) ? result : 0n;
	}

	public static ration(amount: Amount, ration: bigint): bigint {
		if (ration === 0n || amount === 0n) {
			return 0n;
		}

		const result = amount * ration;

		return inInt64Range(result) ? result : 0n;
	}

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

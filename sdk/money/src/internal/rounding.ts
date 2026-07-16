import { ERR_INVALID_EXCHANGE_RATE } from '#money/errors';

export const roundAwayFromZero = (value: number): bigint => {
	if (!Number.isFinite(value)) {
		throw ERR_INVALID_EXCHANGE_RATE;
	}

	return BigInt(Math.sign(value) * Math.round(Math.abs(value)));
};

/** Divides by a positive divisor, rounding half away from zero. */
export const divideRoundAwayFromZero = (dividend: bigint, divisor: bigint): bigint => {
	const quotient = dividend / divisor;
	const remainder = dividend % divisor;
	const absoluteRemainder = remainder < 0n ? -remainder : remainder;

	if (absoluteRemainder * 2n < divisor) {
		return quotient;
	}

	return dividend < 0n ? quotient - 1n : quotient + 1n;
};

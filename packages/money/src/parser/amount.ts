import { MAX_INT64, MIN_INT64 } from '#money/calculator';
import { ERR_INVALID_AMOUNT, ERR_INVALID_AMOUNT_FRACTION, ERR_INVALID_AMOUNT_MULTIPLE } from '#money/errors';

export const parseStringSign = (amount: string): { amount: string; negative: boolean } => {
	if (amount.startsWith('-')) {
		return { amount: amount.slice(1), negative: true };
	}

	if (amount.startsWith('+')) {
		return { amount: amount.slice(1), negative: false };
	}

	return { amount, negative: false };
};

export const parseDecimalParts = (amount: string): { integerPart: string; decimalPart: string } => {
	const parts = amount.split('.');

	if (parts.length > 2) {
		throw ERR_INVALID_AMOUNT_MULTIPLE;
	}

	return {
		integerPart: parts[0] === '' ? '0' : (parts[0] ?? '0'),
		decimalPart: parts[1] ?? '',
	};
};

export const validateAndPadDecimal = (decimalPart: string, fraction: number): string => {
	if (decimalPart.length > fraction) {
		throw ERR_INVALID_AMOUNT_FRACTION;
	}

	return decimalPart.padEnd(fraction, '0');
};

export const parseAmountString = (amount: string, fraction: number, negative: boolean): bigint => {
	const { integerPart, decimalPart } = parseDecimalParts(amount);
	const paddedDecimal = validateAndPadDecimal(decimalPart, fraction);
	const combined = `${integerPart}${paddedDecimal}`;

	if (!/^\d+$/u.test(combined)) {
		throw ERR_INVALID_AMOUNT;
	}

	let value = BigInt(combined);

	if (negative) {
		value = -value;
	}

	if (value < MIN_INT64 || value > MAX_INT64) {
		throw ERR_INVALID_AMOUNT;
	}

	return value;
};

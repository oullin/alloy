import { ERR_INVALID_EXCHANGE_RATE } from '#money/errors';

export const roundAwayFromZero = (value: number): bigint => {
	if (!Number.isFinite(value)) {
		throw ERR_INVALID_EXCHANGE_RATE;
	}

	return BigInt(Math.sign(value) * Math.round(Math.abs(value)));
};

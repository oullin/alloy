import { ERR_EMPTY_AMOUNT_STRING } from '#money/errors';

export const ensureAmountString = (amount: string): string => {
	const trimmed = amount.trim();

	if (trimmed === '') {
		throw ERR_EMPTY_AMOUNT_STRING;
	}

	return trimmed;
};

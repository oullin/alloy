import { ERR_INVALID_MONEY_STRING } from '#money/errors';

export const parseNumericString = (rawInput: string, useDecimalComma: boolean): number => {
	let input = rawInput.trim().replaceAll(' ', '');

	const hasDot = input.includes('.');
	const hasComma = input.includes(',');

	if (hasDot && hasComma) {
		const lastDot = input.lastIndexOf('.');
		const lastComma = input.lastIndexOf(',');

		if (lastDot < lastComma) {
			if (!validThousandsGrouping(input, '.', ',')) {
				throw ERR_INVALID_MONEY_STRING;
			}

			input = input.replaceAll('.', '').replaceAll(',', '.');
		} else {
			if (!validThousandsGrouping(input, ',', '.')) {
				throw ERR_INVALID_MONEY_STRING;
			}

			input = input.replaceAll(',', '');
		}
	} else if (hasComma) {
		input = useDecimalComma ? input.replaceAll(',', '.') : input.replaceAll(',', '');
	}

	const amount = Number.parseFloat(input);

	if (!Number.isFinite(amount) || !/^[-+]?\d*(?:\.\d*)?$/u.test(input) || input === '' || input === '-' || input === '+') {
		throw ERR_INVALID_MONEY_STRING;
	}

	return amount;
};

const validThousandsGrouping = (input: string, thousandsSeparator: string, decimalSeparator: string): boolean => {
	const decimalIndex = input.lastIndexOf(decimalSeparator);

	let integerPart = decimalIndex === -1 ? input : input.slice(0, decimalIndex);

	if (integerPart.length === 0) {
		return false;
	}

	if (integerPart[0] === '-' || integerPart[0] === '+') {
		integerPart = integerPart.slice(1);
	}

	const groups = integerPart.split(thousandsSeparator);

	if (groups.length === 1) {
		return true;
	}

	if ((groups[0]?.length ?? 0) === 0 || (groups[0]?.length ?? 0) > 3) {
		return false;
	}

	return groups.slice(1).every((group) => group.length === 3);
};

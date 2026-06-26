let dbMoneyValueSeparator = '|';

export const getDbMoneyValueSeparator = (): string => dbMoneyValueSeparator;

export const setDbMoneyValueSeparator = (separator: string): void => {
	if (separator.trim() === '') {
		throw new Error(`separator [${separator}] cannot be empty`);
	}

	dbMoneyValueSeparator = separator;
};

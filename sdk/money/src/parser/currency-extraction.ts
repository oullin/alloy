import { CurrencyManager } from '#money/currency/manager';
import type { ISOCodePattern } from '#money/currency/iso-code-pattern';
import { ERR_CURRENCY_NOT_SPECIFIED, ERR_PARSER_INVALID_STATE } from '#money/errors';

export const extractCurrency = (iso: ISOCodePattern, input: string, defaultCurrency?: string): { input: string; currency: string } => {
	if (iso === null) {
		throw ERR_PARSER_INVALID_STATE;
	}

	let currency = defaultCurrency ?? '';
	let value = input;

	for (const symbol of iso.getSymbolsLongestFirst()) {
		if (value.includes(symbol.id)) {
			currency = symbol.currency;
			value = value.replaceAll(symbol.id, '');
			break;
		}
	}

	value = value.trim();

	const pattern = iso.getPattern();
	const match = pattern.exec(value);

	if (match?.[1] !== undefined) {
		currency = match[1];
		value = `${value.slice(0, match.index)}${value.slice(match.index + match[0].length)}`;
	}

	if (currency === '') {
		throw ERR_CURRENCY_NOT_SPECIFIED;
	}

	if (CurrencyManager.default().findByCode(currency) === null) {
		throw ERR_CURRENCY_NOT_SPECIFIED;
	}

	return { input: value, currency };
};

import { CURRENCY_DATA, CURRENCY_SYMBOLS, type CurrencyCode, type CurrencySymbolData } from '#money/currency-data';
import { CurrencyDefinition } from '#money/currency/model';

export interface CurrencyProvider {
	get(): CurrencyDefinition;
	getCode(): CurrencyCode;
	getSymbols(): CurrencySymbolData[];
}

export class DefaultCurrencyProvider implements CurrencyProvider {
	public getCode(): CurrencyCode {
		return 'SGD';
	}

	public get(): CurrencyDefinition {
		return new CurrencyDefinition(CURRENCY_DATA.SGD);
	}

	public getSymbols(): CurrencySymbolData[] {
		return CURRENCY_SYMBOLS.map((symbol) => ({ id: symbol.id, currency: symbol.currency }));
	}
}

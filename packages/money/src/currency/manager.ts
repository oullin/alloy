import type { CurrencyCode, CurrencySymbolData } from '#money/currency-data';
import { CurrencyDefinition } from '#money/currency/definition';
import { CurrencyMap } from '#money/currency/map';
import { DefaultCurrencyProvider, type CurrencyProvider } from '#money/currency/provider';

export class CurrencyManager {
	private readonly currencies: CurrencyMap;
	private readonly symbols: CurrencySymbolData[];
	private readonly defaultCurrency: CurrencyDefinition;

	public constructor(currencies: CurrencyMap = CurrencyMap.default(), provider: CurrencyProvider = new DefaultCurrencyProvider()) {
		this.currencies = currencies;
		this.symbols = provider.getSymbols();
		this.defaultCurrency = provider.get();
	}

	public static default(): CurrencyManager {
		return new CurrencyManager();
	}

	public static withProvider(provider: CurrencyProvider): CurrencyManager {
		return new CurrencyManager(CurrencyMap.default(), provider);
	}

	public static for(provider: CurrencyProvider | null, dataset: ReadonlyMap<string, CurrencyDefinition> | Record<string, CurrencyDefinition>): CurrencyManager {
		const resolvedProvider = provider ?? new DefaultCurrencyProvider();
		const map = CurrencyMap.from(dataset);

		map.assertValid();

		return new CurrencyManager(map, resolvedProvider);
	}

	public findByCode(code: string): CurrencyDefinition | null {
		return this.currencies.findByCode(code);
	}

	public findByNumericCode(code: string): CurrencyDefinition | null {
		const lookup = code.trim().toUpperCase();

		return this.currencies.values().find((currency) => currency.numericCode === lookup) ?? null;
	}

	public add(currency: CurrencyDefinition | null): CurrencyDefinition | null {
		if (currency === null) {
			return null;
		}

		this.currencies.set(currency);

		return currency;
	}

	public addFrom(code: CurrencyCode, grapheme: string, template: string, decimal: string, thousand: string, numericCode: string, fraction: number): CurrencyDefinition | null {
		return this.add(new CurrencyDefinition({ code, decimal, fraction, grapheme, numericCode: numericCode.toUpperCase(), template, thousand }));
	}

	public getDefault(): CurrencyDefinition {
		return this.defaultCurrency;
	}

	public getSymbols(): CurrencySymbolData[] {
		return [...this.symbols];
	}

	public resolve(code: string): CurrencyDefinition {
		return this.findByCode(code) ?? this.defaultCurrency;
	}
}

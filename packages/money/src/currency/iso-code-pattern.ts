import type { CurrencySymbolData } from '#money/currency-data';
import type { CurrencyDefinition } from '#money/currency/definition';
import { CurrencyMap } from '#money/currency/map';
import { DefaultCurrencyProvider, type CurrencyProvider } from '#money/currency/provider';
import { ERR_NO_CONVERTER_PROVIDED } from '#money/errors';

export class ISOCodePattern {
	private readonly currencies: CurrencyMap;
	private readonly symbols: CurrencySymbolData[];
	private symbolsLongestFirstValue: CurrencySymbolData[] | null = null;
	private patternValue: RegExp | null = null;

	public constructor(currencies: CurrencyMap = CurrencyMap.default(), provider: CurrencyProvider = new DefaultCurrencyProvider()) {
		this.currencies = currencies;
		this.symbols = provider.getSymbols();
	}

	public static default(): ISOCodePattern {
		return new ISOCodePattern();
	}

	public static with(provider: CurrencyProvider | null, dataset: ReadonlyMap<string, CurrencyDefinition> | Record<string, CurrencyDefinition>): ISOCodePattern {
		if (provider === null) {
			throw ERR_NO_CONVERTER_PROVIDED;
		}

		return new ISOCodePattern(CurrencyMap.from(dataset), provider);
	}

	public getSymbolsLongestFirst(): CurrencySymbolData[] {
		this.symbolsLongestFirstValue ??= [...this.symbols].sort((left, right) => {
			const byLength = right.id.length - left.id.length;

			return byLength === 0 ? left.id.localeCompare(right.id) : byLength;
		});

		return [...this.symbolsLongestFirstValue];
	}

	public getPattern(): RegExp {
		this.patternValue ??= new RegExp(`\\b(${this.currencies.getCodes().join('|')})\\b`);

		return this.patternValue;
	}
}

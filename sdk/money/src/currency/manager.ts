import type { CurrencyCode, CurrencySymbolData } from '#money/currency-data';
import { CurrencyDefinition } from '#money/currency/definition';
import { CurrencyMap } from '#money/currency/map';
import { DefaultCurrencyProvider, type CurrencyProvider } from '#money/currency/provider';
import { ERR_CURRENCY_NOT_FOUND } from '#money/errors';

/**
 * Registry of {@link CurrencyDefinition} entries keyed by ISO code, with a
 * provider-supplied default currency used when a code cannot be resolved.
 */
export class CurrencyManager {
	private readonly currencies: CurrencyMap;
	private readonly symbols: CurrencySymbolData[];
	private defaultCurrency: CurrencyDefinition;

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

	/** Looks up a currency by ISO alpha code, returning null when unknown. */
	public findByCode(code: string): CurrencyDefinition | null {
		return this.currencies.findByCode(code);
	}

	/** Looks up a currency by ISO numeric code, returning null when unknown. */
	public findByNumericCode(code: string): CurrencyDefinition | null {
		const lookup = code.trim().toUpperCase();

		return this.currencies.values().find((currency) => currency.numericCode === lookup) ?? null;
	}

	/** Registers (or replaces) a currency definition; null input is ignored. */
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

	/** Returns the provider-supplied default currency. */
	public getDefault(): CurrencyDefinition {
		return this.defaultCurrency;
	}

	public getSymbols(): CurrencySymbolData[] {
		return [...this.symbols];
	}

	/** Resolves a code to its definition, falling back to the default currency. */
	public resolve(code: string): CurrencyDefinition {
		return this.findByCode(code) ?? this.defaultCurrency;
	}

	/**
	 * Changes the currency this manager falls back to when {@link resolve} cannot
	 * match a code. The code is trimmed and upper-cased before lookup, so `'usd'`
	 * and `' USD '` both select USD.
	 *
	 * The default belongs to the manager, not the module. The constructor takes
	 * its starting value from the provider and copies it, so this takes effect
	 * immediately on this instance — including one built before the call — and
	 * leaves every other manager alone.
	 *
	 * @throws MoneyError `ERR_CURRENCY_NOT_FOUND` when this manager does not know
	 *   the code, leaving the existing default in place.
	 */
	public setDefault(code: string): void {
		const definition = this.findByCode(code);

		if (definition === null) {
			throw ERR_CURRENCY_NOT_FOUND;
		}

		this.defaultCurrency = definition;
	}
}

import { MoneyFormatter } from '#money/format';
import { ERR_CURRENCY_NOT_FOUND, ERR_NO_CONVERTER_PROVIDED, ERR_NO_CURRENCY_MAP_DATASET } from '#money/errors';
import { CURRENCY_CODES, CURRENCY_DATA, CURRENCY_SYMBOLS, type CurrencyCode, type CurrencyDefinitionData, type CurrencySymbolData } from '#money/currency-data';

export { CURRENCY_CODES, CURRENCY_SYMBOLS, type CurrencyCode, type CurrencyDefinitionData, type CurrencySymbolData };

export class CurrencyDefinition {
	public readonly decimal: string;
	public readonly thousand: string;
	public readonly code: CurrencyCode;
	public readonly fraction: number;
	public readonly numericCode: string;
	public readonly grapheme: string;
	public readonly template: string;

	public constructor(data: CurrencyDefinitionData) {
		this.decimal = data.decimal;
		this.thousand = data.thousand;
		this.code = data.code;
		this.fraction = data.fraction;
		this.numericCode = data.numericCode;
		this.grapheme = data.grapheme;
		this.template = data.template;
	}

	public formatter(): MoneyFormatter {
		return MoneyFormatter.create(this.fraction, this.decimal, this.thousand, this.grapheme, this.template);
	}

	public equals(other: CurrencyDefinition | null | undefined): boolean {
		return other instanceof CurrencyDefinition && this.code === other.code;
	}

	public get(): CurrencyDefinition {
		return this;
	}

	public dbValue(): string {
		return this.code;
	}

	public static fromDbValue(input: string | Uint8Array): CurrencyDefinition {
		const code = typeof input === 'string' ? input : new TextDecoder().decode(input);
		const currency = CurrencyManager.default().findByCode(code);

		if (currency === null) {
			throw ERR_CURRENCY_NOT_FOUND;
		}

		return currency;
	}
}

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

export class CurrencyMap {
	private static defaultDefinitions: CurrencyDefinition[] | null = null;
	private readonly dataset: Map<string, CurrencyDefinition>;

	public constructor(dataset: Iterable<CurrencyDefinition>) {
		this.dataset = new Map();

		for (const currency of dataset) {
			this.dataset.set(currency.code, currency);
		}
	}

	public static default(): CurrencyMap {
		CurrencyMap.defaultDefinitions ??= Object.values(CURRENCY_DATA).map((data) => new CurrencyDefinition(data));

		return new CurrencyMap(CurrencyMap.defaultDefinitions);
	}

	public static from(dataset: ReadonlyMap<string, CurrencyDefinition> | Record<string, CurrencyDefinition>): CurrencyMap {
		const values = dataset instanceof Map ? [...dataset.values()] : Object.values(dataset);

		if (values.length === 0) {
			throw ERR_NO_CURRENCY_MAP_DATASET;
		}

		return new CurrencyMap(values);
	}

	public get(code: string): CurrencyDefinition | null {
		return this.dataset.get(code) ?? null;
	}

	public set(currency: CurrencyDefinition): void {
		this.dataset.set(currency.code, currency);
	}

	public isEmpty(): boolean {
		return this.dataset.size === 0;
	}

	public isNotEmpty(): boolean {
		return !this.isEmpty();
	}

	public assertValid(): void {
		if (this.dataset.size === 0) {
			throw ERR_NO_CURRENCY_MAP_DATASET;
		}
	}

	public findByCode(code: string): CurrencyDefinition | null {
		return this.dataset.get(code.trim().toUpperCase()) ?? null;
	}

	public getCodes(): CurrencyCode[] {
		return [...this.dataset.keys()] as CurrencyCode[];
	}

	public values(): CurrencyDefinition[] {
		return [...this.dataset.values()];
	}
}

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

import { CURRENCY_DATA, type CurrencyCode, type CurrencyDefinitionData } from '#money/currency-data';
import { CurrencyManager } from '#money/currency/manager';
import { ERR_CURRENCY_NOT_FOUND, ERR_NO_CURRENCY_MAP_DATASET } from '#money/errors';
import { MoneyFormatter } from '#money/format';

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

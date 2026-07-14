import { CURRENCY_DATA, type CurrencyCode } from '#money/currency-data';
import { CurrencyDefinition } from '#money/currency/definition';
import { ERR_NO_CURRENCY_MAP_DATASET } from '#money/errors';

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

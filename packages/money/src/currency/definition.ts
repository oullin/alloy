import type { CurrencyCode, CurrencyDefinitionData } from '#money/currency-data';
import { CurrencyManager } from '#money/currency/manager';
import { ERR_CURRENCY_NOT_FOUND } from '#money/errors';
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

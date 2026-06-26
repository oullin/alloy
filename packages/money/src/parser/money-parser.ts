import { ISOCodePattern } from '#money/currency/iso-code-pattern';
import { ERR_INVALID_MONEY_STRING } from '#money/errors';
import { parseAmountString, parseDecimalParts, parseStringSign, validateAndPadDecimal } from '#money/parser/amount';
import { extractCurrency } from '#money/parser/currency-extraction';
import { parseNumericString } from '#money/parser/numeric';
import type { ParsedMoneyAmount } from '#money/parser/types';

export class MoneyParser {
	private readonly iso: ISOCodePattern;

	public constructor(iso: ISOCodePattern = ISOCodePattern.default()) {
		this.iso = iso;
	}

	public static create(): MoneyParser {
		return new MoneyParser();
	}

	public static with(iso: ISOCodePattern): MoneyParser {
		return new MoneyParser(iso);
	}

	public parseAmount(input: string, defaultCurrency?: string): ParsedMoneyAmount {
		const value = input.trim();

		if (value === '') {
			throw ERR_INVALID_MONEY_STRING;
		}

		const extracted = extractCurrency(this.iso, value, defaultCurrency);

		return {
			amount: this.parseNumericString(extracted.input, false),
			currency: extracted.currency,
		};
	}

	public parseAmountWithDecimalComma(input: string, defaultCurrency?: string): ParsedMoneyAmount {
		const value = input.trim();

		if (value === '') {
			throw ERR_INVALID_MONEY_STRING;
		}

		const extracted = extractCurrency(this.iso, value, defaultCurrency);

		return {
			amount: this.parseNumericString(extracted.input, true),
			currency: extracted.currency,
		};
	}

	public parseDecimal(amount: string): number {
		return this.parseNumericString(amount, false);
	}

	public parseDecimalWithComma(amount: string): number {
		return this.parseNumericString(amount, true);
	}

	public parseStringSign(amount: string): { amount: string; negative: boolean } {
		return parseStringSign(amount);
	}

	public parseDecimalParts(amount: string): { integerPart: string; decimalPart: string } {
		return parseDecimalParts(amount);
	}

	public validateAndPadDecimal(decimalPart: string, fraction: number): string {
		return validateAndPadDecimal(decimalPart, fraction);
	}

	public parseAmountString(amount: string, fraction: number, negative: boolean): bigint {
		return parseAmountString(amount, fraction, negative);
	}

	private parseNumericString(rawInput: string, useDecimalComma: boolean): number {
		return parseNumericString(rawInput, useDecimalComma);
	}
}

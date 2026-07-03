import { ISOCodePattern } from '#money/currency/iso-code-pattern';
import { ERR_INVALID_MONEY_STRING } from '#money/errors';
import { parseAmountString, parseDecimalParts, parseStringSign, validateAndPadDecimal } from '#money/parser/amount';
import { extractCurrency } from '#money/parser/currency-extraction';
import { parseNumericString } from '#money/parser/numeric';
import type { ParsedMoneyAmount } from '#money/parser/types';

/**
 * Parses human-entered money strings such as `"USD 12.34"`, `"$12.34"`, or
 * `"12,34 EUR"` into a numeric amount plus a detected ISO currency code.
 */
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

	/**
	 * Parses a money string into an amount and currency, treating `.` as the
	 * decimal separator.
	 *
	 * @throws MoneyError `ERR_INVALID_MONEY_STRING` for empty input, `ERR_CURRENCY_NOT_SPECIFIED` when no currency is detected or defaulted.
	 */
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

	/** Same as {@link parseAmount} but treats `,` as the decimal separator. */
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

	/**
	 * Parses a plain numeric string (`.` decimal separator) into a number.
	 *
	 * @throws MoneyError `ERR_INVALID_MONEY_STRING` for malformed input.
	 */
	public parseDecimal(amount: string): number {
		return this.parseNumericString(amount, false);
	}

	/** Same as {@link parseDecimal} but treats `,` as the decimal separator. */
	public parseDecimalWithComma(amount: string): number {
		return this.parseNumericString(amount, true);
	}

	/** Strips a leading sign from an amount string and reports whether it was negative. */
	public parseStringSign(amount: string): { amount: string; negative: boolean } {
		return parseStringSign(amount);
	}

	public parseDecimalParts(amount: string): { integerPart: string; decimalPart: string } {
		return parseDecimalParts(amount);
	}

	public validateAndPadDecimal(decimalPart: string, fraction: number): string {
		return validateAndPadDecimal(decimalPart, fraction);
	}

	/**
	 * Converts an unsigned decimal amount string to minor units at the given fraction.
	 *
	 * @throws MoneyError `ERR_INVALID_AMOUNT`, `ERR_INVALID_AMOUNT_MULTIPLE`, or `ERR_INVALID_AMOUNT_FRACTION` for malformed input.
	 */
	public parseAmountString(amount: string, fraction: number, negative: boolean): bigint {
		return parseAmountString(amount, fraction, negative);
	}

	private parseNumericString(rawInput: string, useDecimalComma: boolean): number {
		return parseNumericString(rawInput, useDecimalComma);
	}
}

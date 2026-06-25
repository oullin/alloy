import { MAX_INT64, MIN_INT64 } from '#money/calculator';
import { CurrencyManager, ISOCodePattern } from '#money/currency';

import { ERR_CURRENCY_NOT_SPECIFIED, ERR_INVALID_AMOUNT, ERR_INVALID_AMOUNT_FRACTION, ERR_INVALID_AMOUNT_MULTIPLE, ERR_INVALID_MONEY_STRING, ERR_PARSER_INVALID_STATE } from '#money/errors';

export interface ParsedMoneyAmount {
	readonly amount: number;
	readonly currency: string;
}

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

		const extracted = this.extractCurrency(value, defaultCurrency);

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

		const extracted = this.extractCurrency(value, defaultCurrency);

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
		if (amount.startsWith('-')) {
			return { amount: amount.slice(1), negative: true };
		}

		if (amount.startsWith('+')) {
			return { amount: amount.slice(1), negative: false };
		}

		return { amount, negative: false };
	}

	public parseDecimalParts(amount: string): { integerPart: string; decimalPart: string } {
		const parts = amount.split('.');

		if (parts.length > 2) {
			throw ERR_INVALID_AMOUNT_MULTIPLE;
		}

		return {
			integerPart: parts[0] === '' ? '0' : (parts[0] ?? '0'),
			decimalPart: parts[1] ?? '',
		};
	}

	public validateAndPadDecimal(decimalPart: string, fraction: number): string {
		if (decimalPart.length > fraction) {
			throw ERR_INVALID_AMOUNT_FRACTION;
		}

		return decimalPart.padEnd(fraction, '0');
	}

	public parseAmountString(amount: string, fraction: number, negative: boolean): bigint {
		const { integerPart, decimalPart } = this.parseDecimalParts(amount);
		const paddedDecimal = this.validateAndPadDecimal(decimalPart, fraction);
		const combined = `${integerPart}${paddedDecimal}`;

		if (!/^\d+$/u.test(combined)) {
			throw ERR_INVALID_AMOUNT;
		}

		let value = BigInt(combined);

		if (negative) {
			value = -value;
		}

		if (value < MIN_INT64 || value > MAX_INT64) {
			throw ERR_INVALID_AMOUNT;
		}

		return value;
	}

	private parseNumericString(rawInput: string, useDecimalComma: boolean): number {
		let input = rawInput.trim().replaceAll(' ', '');

		const hasDot = input.includes('.');
		const hasComma = input.includes(',');

		if (hasDot && hasComma) {
			const lastDot = input.lastIndexOf('.');
			const lastComma = input.lastIndexOf(',');

			if (lastDot < lastComma) {
				if (!this.validThousandsGrouping(input, '.', ',')) {
					throw ERR_INVALID_MONEY_STRING;
				}

				input = input.replaceAll('.', '').replaceAll(',', '.');
			} else {
				if (!this.validThousandsGrouping(input, ',', '.')) {
					throw ERR_INVALID_MONEY_STRING;
				}

				input = input.replaceAll(',', '');
			}
		} else if (hasComma) {
			input = useDecimalComma ? input.replaceAll(',', '.') : input.replaceAll(',', '');
		}

		const amount = Number.parseFloat(input);

		if (!Number.isFinite(amount) || !/^[-+]?\d*(?:\.\d*)?$/u.test(input) || input === '' || input === '-' || input === '+') {
			throw ERR_INVALID_MONEY_STRING;
		}

		return amount;
	}

	private validThousandsGrouping(input: string, thousandsSeparator: string, decimalSeparator: string): boolean {
		const decimalIndex = input.lastIndexOf(decimalSeparator);

		let integerPart = decimalIndex === -1 ? input : input.slice(0, decimalIndex);

		if (integerPart.length === 0) {
			return false;
		}

		if (integerPart[0] === '-' || integerPart[0] === '+') {
			integerPart = integerPart.slice(1);
		}

		const groups = integerPart.split(thousandsSeparator);

		if (groups.length === 1) {
			return true;
		}

		if ((groups[0]?.length ?? 0) === 0 || (groups[0]?.length ?? 0) > 3) {
			return false;
		}

		return groups.slice(1).every((group) => group.length === 3);
	}

	private extractCurrency(input: string, defaultCurrency?: string): { input: string; currency: string } {
		if (this.iso === null) {
			throw ERR_PARSER_INVALID_STATE;
		}

		let currency = defaultCurrency ?? '';
		let value = input;

		for (const symbol of this.iso.getSymbolsLongestFirst()) {
			if (value.includes(symbol.id)) {
				currency = symbol.currency;
				value = value.replaceAll(symbol.id, '');
				break;
			}
		}

		value = value.trim();

		const pattern = this.iso.getPattern();
		const match = pattern.exec(value);

		if (match?.[1] !== undefined) {
			currency = match[1];
			value = `${value.slice(0, match.index)}${value.slice(match.index + match[0].length)}`;
		}

		if (currency === '') {
			throw ERR_CURRENCY_NOT_SPECIFIED;
		}

		if (CurrencyManager.default().findByCode(currency) === null) {
			throw ERR_CURRENCY_NOT_SPECIFIED;
		}

		return { input: value, currency };
	}
}

import { describe, expect, it } from 'vite-plus/test';

import {
	CurrencyManager,
	ERR_CURRENCY_CONVERSION_NOT_FOUND,
	ERR_CURRENCY_NOT_SPECIFIED,
	ERR_INVALID_AMOUNT,
	ERR_INVALID_AMOUNT_FRACTION,
	ERR_INVALID_AMOUNT_MULTIPLE,
	ERR_INVALID_EXCHANGE_RATE,
	ERR_INVALID_MONEY_STRING,
	ExchangeRates,
	MoneyConverter,
	MoneyFormatter,
	MoneyManager,
	MoneyParser,
} from '#money/index';

describe('money parser, formatter, and exchange', () => {
	it('parses localized symbol and code forms with signed amounts', () => {
		const parser = MoneyParser.create();

		expect(parser.parseAmount('$1,234.56')).toEqual({ amount: 1234.56, currency: 'USD' });
		expect(parser.parseAmount('-S$1,234.56')).toEqual({ amount: -1234.56, currency: 'SGD' });
		expect(parser.parseAmount('1,234.56 USD')).toEqual({ amount: 1234.56, currency: 'USD' });
		expect(parser.parseAmount('EUR -1234.56')).toEqual({ amount: -1234.56, currency: 'EUR' });
		expect(parser.parseAmountWithDecimalComma('1.234,56 EUR')).toEqual({ amount: 1234.56, currency: 'EUR' });
		expect(parser.parseAmount('123.45', 'USD')).toEqual({ amount: 123.45, currency: 'USD' });
	});

	it('throws coded parser errors for malformed input', () => {
		const parser = MoneyParser.create();

		expect(() => parser.parseAmount('')).toThrow(ERR_INVALID_MONEY_STRING.message);
		expect(() => parser.parseAmount('123.45')).toThrow(ERR_CURRENCY_NOT_SPECIFIED.message);
		expect(() => parser.parseAmount('123.45 XYZ')).toThrow(ERR_CURRENCY_NOT_SPECIFIED.message);
		expect(() => parser.parseDecimal('1.2.3')).toThrow(ERR_INVALID_MONEY_STRING.message);
		expect(() => parser.parseAmountString('1.2.3', 2, false)).toThrow(ERR_INVALID_AMOUNT_MULTIPLE.message);
		expect(() => parser.parseAmountString('12.345', 2, false)).toThrow(ERR_INVALID_AMOUNT_FRACTION.message);
		expect(() => parser.parseAmountString('12.x', 2, false)).toThrow(ERR_INVALID_AMOUNT.message);
	});

	it('formats locale-aware currencies including zero and three decimal fractions', () => {
		const manager = MoneyManager.default();

		expect(manager.create(123456n, 'USD').display()).toBe('$1,234.56');
		expect(manager.create(123456n, 'EUR').display()).toBe('€1,234.56');
		expect(manager.create(123456n, 'JPY').display()).toBe('¥123,456');
		expect(manager.create(123456n, 'BHD').display()).toBe('123.456 .د.ب');
		expect(MoneyFormatter.create(2, ',', '.', '€', '1 $').format(-123456n)).toBe('-1.234,56 €');
	});

	it('converts direct, inverse, identity, and explicit rates', () => {
		const rates = ExchangeRates.create().addRate('USD', 'EUR', 0.8).addRate('USD', 'JPY', 150);
		const converter = MoneyConverter.create(CurrencyManager.default(), rates);
		const manager = MoneyManager.default();
		const usd = manager.create(12345n, 'USD');

		expect(converter.convert(usd, 'USD').amount()).toBe(12345n);
		expect(converter.convert(usd, 'EUR').amount()).toBe(9876n);
		expect(converter.convert(manager.create(9876n, 'EUR'), 'USD').amount()).toBe(12345n);
		expect(converter.convert(usd, 'JPY').amount()).toBe(18518n);
		expect(converter.convertWithRate(usd, 'BHD', 0.377).amount()).toBe(46541n);
		expect(() => converter.convert(usd, 'GBP')).toThrow(ERR_CURRENCY_CONVERSION_NOT_FOUND.message);
		expect(() => rates.addRate('USD', 'EUR', Number.POSITIVE_INFINITY)).toThrow(ERR_INVALID_EXCHANGE_RATE.message);
	});
});

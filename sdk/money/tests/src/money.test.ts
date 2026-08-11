import { describe, expect, it } from 'vite-plus/test';
import type { Amount, CurrencyCode, MoneyValueInstance } from '#money/index';

import {
	CURRENCY_CODES,
	CurrencyDefinition,
	CurrencyMap,
	CurrencyManager,
	ERR_CURRENCY_CONVERSION_NOT_FOUND,
	ERR_CURRENCY_MISMATCH,
	ERR_INVALID_AMOUNT_FRACTION,
	ERR_INVALID_EXCHANGE_RATE,
	ERR_INVALID_JSON_UNMARSHAL,
	ERR_INVALID_MONEY_STRING,
	ERR_NO_MULTIPLIERS_PROVIDED,
	ERR_OVERFLOW,
	ExchangeRates,
	MAX_INT64,
	MIN_INT64,
	MoneyAggregator,
	MoneyCalculator,
	MoneyConverter,
	MoneyFormatter,
	MoneyJson,
	MoneyManager,
	MoneyParser,
	MoneyValue,
} from '#money/index';

type MoneyValueFactories = {
	[Code in CurrencyCode as `from${Code}`]: (amount: Amount) => MoneyValueInstance;
};

describe('@hara/sdk-money OOP port', () => {
	it('keeps the root package class-first without top-level DB helper functions', async () => {
		const api = await import('#money/index');

		expect('getDbMoneyValueSeparator' in api).toBe(false);
		expect('setDbMoneyValueSeparator' in api).toBe(false);
		expect(typeof api.MoneyValue.getDbMoneyValueSeparator).toBe('function');
		expect(typeof api.MoneyValue.setDbMoneyValueSeparator).toBe('function');
	});

	it('creates immutable money values through managers and static currency factories', () => {
		const manager = MoneyManager.default();
		const sgd = manager.create(1500n, 'SGD');

		expect(sgd.amount()).toBe(1500n);
		expect(sgd.currency().code).toBe('SGD');
		expect(sgd.display()).toBe('S$15.00');
		expect(sgd.asMajorUnits()).toBe(15);
		expect(sgd.isPositive()).toBe(true);

		const factories = MoneyValue as typeof MoneyValue & MoneyValueFactories;

		for (const code of CURRENCY_CODES) {
			const money = factories[`from${code}`](42n);

			expect(money.amount()).toBe(42n);
			expect(money.currency().code).toBe(code);
		}
	});

	it('performs manager arithmetic without mutating the source value', () => {
		const manager = MoneyManager.default();
		const left = manager.createFromString('12.34', 'USD');
		const right = manager.createFromFloat(1.25, 'USD');

		expect(left.amount()).toBe(1234n);
		expect(right.amount()).toBe(125n);

		const total = manager.add(left, right);
		const remaining = manager.subtract(total, right);
		const multiplied = manager.multiply(remaining, 2n, -3n);

		expect(total.amount()).toBe(1359n);
		expect(remaining.amount()).toBe(1234n);
		expect(multiplied.amount()).toBe(-7404n);
		expect(left.amount()).toBe(1234n);
		expect(() => manager.multiply(left)).toThrow(ERR_NO_MULTIPLIERS_PROVIDED.message);
	});

	it('preserves manager int64 overflow behavior', () => {
		const manager = MoneyManager.default();

		expect(() => manager.add(manager.create(MAX_INT64, 'SGD'), manager.create(1n, 'SGD'))).toThrow(ERR_OVERFLOW.message);
		expect(() => manager.subtract(manager.create(MIN_INT64, 'SGD'), manager.create(1n, 'SGD'))).toThrow(ERR_OVERFLOW.message);
	});

	it('compares and aggregates same-currency values', () => {
		const manager = MoneyManager.default();
		const low = manager.create(100n, 'EUR');
		const mid = manager.create(200n, 'EUR');
		const high = manager.create(300n, 'EUR');
		const usd = manager.create(100n, 'USD');
		const aggregator = MoneyAggregator.create(manager);

		expect(high.greaterThan(low)).toBe(true);
		expect(mid.equals(manager.create(200n, 'EUR'))).toBe(true);
		expect(aggregator.sum(low, mid, high).amount()).toBe(600n);
		expect(aggregator.min(high, low, mid)).toBe(low);
		expect(aggregator.max(low, high, mid)).toBe(high);
		expect(aggregator.avg(low, mid, high).amount()).toBe(200n);
		expect(aggregator.avg(manager.create(100n, 'EUR'), manager.create(101n, 'EUR')).amount()).toBe(101n);
		expect(aggregator.avg(manager.create(-100n, 'EUR'), manager.create(-101n, 'EUR')).amount()).toBe(-101n);
		expect(() => low.assertSameCurrency(usd)).toThrow(ERR_CURRENCY_MISMATCH.message);
	});

	it('splits and allocates remainders in input order', () => {
		const manager = MoneyManager.default();

		expect(manager.split(manager.create(5n, 'USD'), 2).map((money) => money.amount())).toEqual([3n, 2n]);
		expect(manager.split(manager.create(-5n, 'USD'), 2).map((money) => money.amount())).toEqual([-3n, -2n]);
		expect(manager.allocate(manager.create(5n, 'USD'), 1, 1, 1).map((money) => money.amount())).toEqual([2n, 2n, 1n]);
		expect(manager.allocate(manager.create(2n, 'USD'), 1, 3).map((money) => money.amount())).toEqual([1n, 1n]);
	});

	it('formats and parses money strings with currency symbols and decimal policies', () => {
		const formatter = MoneyFormatter.create(2, '.', ',', '$', '$1');
		const parser = MoneyParser.create();

		expect(formatter.format(123456789n)).toBe('$1,234,567.89');
		expect(formatter.format(-1n)).toBe('-$0.01');
		expect(parser.parseAmount('S$1,234.56')).toEqual({ amount: 1234.56, currency: 'SGD' });
		expect(parser.parseAmount('€1.234,56')).toEqual({ amount: 1234.56, currency: 'EUR' });
		expect(parser.parseAmountWithDecimalComma('10,50 EUR')).toEqual({ amount: 10.5, currency: 'EUR' });
		expect(() => parser.parseDecimal('1.2,3')).toThrow(ERR_INVALID_MONEY_STRING.message);
		expect(() => managerCreateFromString('12.345', 'USD')).toThrow(ERR_INVALID_AMOUNT_FRACTION.message);
	});

	it('looks up the full currency dataset by code and numeric code', () => {
		const manager = CurrencyManager.default();

		expect(CURRENCY_CODES.length).toBeGreaterThan(150);
		expect(manager.findByCode('sgd')?.code).toBe('SGD');
		expect(manager.findByNumericCode('840')?.code).toBe('USD');
		expect(manager.resolve('UNKNOWN').code).toBe('SGD');
		expect(CurrencyDefinition.fromDbValue('eur').code).toBe('EUR');
	});

	it('returns independent default currency maps from cached definitions', () => {
		const first = CurrencyMap.default();
		const second = CurrencyMap.default();

		first.set(new CurrencyDefinition({ code: 'SGD', numericCode: '998', fraction: 2, grapheme: 'T$', template: '$1', decimal: '.', thousand: ',' }));

		expect(first.findByCode('SGD')?.numericCode).toBe('998');
		expect(second.findByCode('SGD')?.numericCode).toBe('702');
	});

	it('converts amounts with direct rates, inverse rates, explicit rates, and fraction changes', () => {
		const rates = ExchangeRates.create().addRate('USD', 'EUR', 0.85).addRate('USD', 'JPY', 150);

		const converter = MoneyConverter.create(CurrencyManager.default(), rates);
		const usd = MoneyManager.default().create(500n, 'USD');

		expect(converter.convert(usd, 'EUR').amount()).toBe(425n);
		expect(converter.convert(usd, 'JPY').amount()).toBe(750n);
		expect(converter.convertWithRate(usd, 'EUR', 2).amount()).toBe(1000n);
		expect(rates.convertAmount(425n, 'EUR', 2, 'USD', 2)).toBe(500n);
		expect(() => converter.convert(usd, 'GBP')).toThrow(ERR_CURRENCY_CONVERSION_NOT_FOUND.message);
		expect(() => rates.addRate('USD', 'EUR', 0)).toThrow(ERR_INVALID_EXCHANGE_RATE.message);
	});

	it('marshals JSON explicitly and scans DB values without bigint precision loss', () => {
		const json = MoneyJson.default();
		const money = MoneyManager.default().create(9_223_372_036_854_775_807n, 'USD');

		expect(json.marshal(money)).toBe('{"amount":9223372036854775807,"currency":"USD"}');
		expect(json.unmarshal('{"amount":9223372036854775807,"currency":"USD"}').amount()).toBe(9_223_372_036_854_775_807n);
		expect(json.unmarshal('{"amount":12.5,"currency":"USD"}').amount()).toBe(13n);
		expect(json.unmarshal('{"description":"The amount is 50","amount":100,"currency":"USD"}').amount()).toBe(100n);
		expect(json.unmarshal('{"metadata":{"amount":50},"amount":100,"currency":"USD"}').amount()).toBe(100n);

		// A quoted amount decodes exactly as a bare one, matching Go's json.Number
		// field. This assertion previously required a throw, which pinned a
		// divergence: Go accepted the quoted form, and `toJSON()` emits it, so
		// unmarshalling this runtime's own stringified output failed.
		expect(json.unmarshal('{"amount":"100","currency":"USD"}').amount()).toBe(100n);
		expect(json.unmarshal(JSON.stringify(money)).amount()).toBe(9_223_372_036_854_775_807n);

		// Quoted still means a number, and the grammar is anchored, so no whitespace
		// padding and no numeric prefix of a longer string.
		expect(() => json.unmarshal('{"amount":"abc","currency":"USD"}')).toThrow(ERR_INVALID_JSON_UNMARSHAL.message);
		expect(() => json.unmarshal('{"amount":" 100 ","currency":"USD"}')).toThrow(ERR_INVALID_JSON_UNMARSHAL.message);
		expect(() => json.unmarshal('{"amount":"100abc","currency":"USD"}')).toThrow(ERR_INVALID_JSON_UNMARSHAL.message);
		expect(money.dbValue()).toBe('9223372036854775807|USD');
		expect(MoneyValue.fromDbValue('1234|SGD').display()).toBe('S$12.34');

		const original = MoneyValue.getDbMoneyValueSeparator();

		MoneyValue.setDbMoneyValueSeparator('::');

		try {
			expect(MoneyManager.default().create(10n, 'EUR').dbValue()).toBe('10::EUR');
			expect(MoneyValue.fromDbValue('10::EUR').currency().code).toBe('EUR');
		} finally {
			MoneyValue.setDbMoneyValueSeparator(original);
		}
	});

	it('preserves calculator int64 overflow behavior', () => {
		const calculator = MoneyCalculator.create();

		expect(() => calculator.add(MAX_INT64, 1n)).toThrow(ERR_OVERFLOW.message);
		expect(() => calculator.subtract(MIN_INT64, 1n)).toThrow(ERR_OVERFLOW.message);
		expect(() => calculator.multiply(MAX_INT64, 2n)).toThrow(ERR_OVERFLOW.message);
		expect(() => calculator.safeMultiply(MAX_INT64, 2n)).toThrow(ERR_OVERFLOW.message);
	});
});

const managerCreateFromString = (amount: string, code: string): MoneyValueInstance => MoneyManager.default().createFromString(amount, code);

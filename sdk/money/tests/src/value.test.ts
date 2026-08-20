import { describe, expect, it } from 'vite-plus/test';

import { ERR_CURRENCY_MISMATCH, ERR_NO_MONEY_PROVIDED, MoneyAggregator, MoneyJson, MoneyManager, MoneyValue } from '#money/index';

describe('money value behavior', () => {
	it('reads the minor-unit amount as a number for APIs that take minor units', () => {
		expect(MoneyValue.fromUSD(1500n).asMinorUnits()).toBe(1500);
		expect(MoneyValue.fromUSD(-4200n).asMinorUnits()).toBe(-4200);
		expect(MoneyValue.fromUSD(0n).asMinorUnits()).toBe(0);
	});

	it('reports whether an amount survives conversion to a number', () => {
		// The amount is a bigint because a portfolio total can exceed the
		// float-safe range. An inexact conversion returns an ordinary-looking
		// number, so the only way to know is to ask before converting.
		expect(MoneyValue.fromUSD(1500n).isSafeAsNumber()).toBe(true);
		expect(MoneyValue.fromUSD(BigInt(Number.MAX_SAFE_INTEGER)).isSafeAsNumber()).toBe(true);
		expect(MoneyValue.fromUSD(BigInt(Number.MIN_SAFE_INTEGER)).isSafeAsNumber()).toBe(true);
		expect(MoneyValue.fromUSD(BigInt(Number.MAX_SAFE_INTEGER) + 1n).isSafeAsNumber()).toBe(false);
		expect(MoneyValue.fromUSD(BigInt(Number.MIN_SAFE_INTEGER) - 1n).isSafeAsNumber()).toBe(false);
	});

	it('keeps an amount past the float-safe range exact on the value itself', () => {
		const beyond = BigInt(Number.MAX_SAFE_INTEGER) + 1n;

		expect(MoneyValue.fromUSD(beyond).amount()).toBe(beyond);
		expect(MoneyValue.fromUSD(beyond).isSafeAsNumber()).toBe(false);
	});

	it('performs immutable arithmetic and comparisons on MoneyValue instances', () => {
		const manager = MoneyManager.default();
		const base = MoneyValue.fromUSD(1000n);
		const fee = MoneyValue.fromUSD(250n);

		expect(manager.add(base, fee).amount()).toBe(1250n);
		expect(manager.subtract(base, fee).amount()).toBe(750n);
		expect(manager.multiply(fee, 3n).amount()).toBe(750n);
		expect(manager.negative(base).amount()).toBe(-1000n);
		expect(manager.absolute(manager.negative(base)).amount()).toBe(1000n);
		expect(base.amount()).toBe(1000n);
		expect(base.greaterThan(fee)).toBe(true);
	});

	it('rejects arithmetic and comparison across currencies', () => {
		const manager = MoneyManager.default();
		const usd = MoneyValue.fromUSD(100n);
		const eur = MoneyValue.fromEUR(100n);

		expect(() => usd.assertSameCurrency(eur)).toThrow(ERR_CURRENCY_MISMATCH.message);
		expect(() => manager.add(usd, eur)).toThrow(ERR_CURRENCY_MISMATCH.message);
		expect(() => usd.compare(eur)).toThrow(ERR_CURRENCY_MISMATCH.message);
	});

	it('round-trips JSON payloads and object JSON values', () => {
		const json = MoneyJson.default();
		const money = MoneyValue.fromBHD(123456n);
		const payload = json.marshal(money);
		const parsed = json.unmarshal(payload);

		expect(payload).toBe('{"amount":123456,"currency":"BHD"}');
		expect(parsed.amount()).toBe(123456n);
		expect(parsed.currency().code).toBe('BHD');
		expect(money.toJSON()).toEqual({ amount: '123456', currency: 'BHD' });
		expect(JSON.parse(JSON.stringify(money))).toEqual({ amount: '123456', currency: 'BHD' });
	});

	it('aggregates money values and rejects empty input', () => {
		const aggregator = MoneyAggregator.create(MoneyManager.default());
		const low = MoneyValue.fromUSD(100n);
		const mid = MoneyValue.fromUSD(200n);
		const high = MoneyValue.fromUSD(301n);

		expect(aggregator.sum(low, mid, high).amount()).toBe(601n);
		expect(aggregator.avg(low, mid, high).amount()).toBe(200n);
		expect(aggregator.min(high, low, mid)).toBe(low);
		expect(aggregator.max(low, high, mid)).toBe(high);
		expect(() => aggregator.sum()).toThrow(ERR_NO_MONEY_PROVIDED.message);
	});
});

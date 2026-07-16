import { describe, expect, it } from 'vite-plus/test';

import { ERR_OVERFLOW, MAX_INT64, MIN_INT64, MoneyCalculator } from '#money/index';

describe('money calculator', () => {
	it('adds, subtracts, multiplies, and divides integer amounts exactly', () => {
		const calculator = MoneyCalculator.create();

		expect(calculator.add(9007199254740991n, 9n)).toBe(9007199254741000n);
		expect(calculator.subtract(10000000000000000n, 1n)).toBe(9999999999999999n);
		expect(calculator.multiply(123456789n, 100000n)).toBe(12345678900000n);
		expect(calculator.divide(100n, 3n)).toBe(33n);
		expect(calculator.modulus(100n, 3n)).toBe(1n);
	});

	it('guards int64 overflow boundaries', () => {
		const calculator = MoneyCalculator.create();

		expect(calculator.add(MAX_INT64, 0n)).toBe(MAX_INT64);
		expect(calculator.subtract(MIN_INT64, 0n)).toBe(MIN_INT64);
		expect(() => calculator.add(MAX_INT64, 1n)).toThrow(ERR_OVERFLOW.message);
		expect(() => calculator.subtract(MIN_INT64, 1n)).toThrow(ERR_OVERFLOW.message);
		expect(() => calculator.multiply(MIN_INT64, -1n)).toThrow(ERR_OVERFLOW.message);
		expect(() => calculator.safeMultiply(3037000500n, 3037000500n)).toThrow(ERR_OVERFLOW.message);
	});

	it('returns int64-safe absolute values', () => {
		const calculator = MoneyCalculator.create();

		expect(calculator.absolute(MIN_INT64)).toBe(0n);
		expect(calculator.absolute(MIN_INT64 + 1n)).toBe(MAX_INT64);
		expect(calculator.absolute(-5n)).toBe(5n);
		expect(calculator.absolute(5n)).toBe(5n);
	});

	it('rounds positive and negative amounts to decimal exponents', () => {
		const calculator = MoneyCalculator.create();

		expect(calculator.round(1549n, 2)).toBe(1500n);
		expect(calculator.round(1551n, 2)).toBe(1600n);
		expect(calculator.round(-1551n, 2)).toBe(-1600n);
		expect(calculator.round(1550n, 2)).toBe(1600n);
		expect(calculator.round(1250n, 2)).toBe(1300n);
		expect(calculator.round(-1250n, 2)).toBe(-1300n);
		expect(calculator.round(250n, 2)).toBe(300n);
		expect(calculator.round(123n, 0)).toBe(123n);
		expect(calculator.round(123n, 19)).toBe(123n);
	});

	it('returns zero for division and modulus by zero', () => {
		const calculator = MoneyCalculator.create();

		expect(calculator.divide(100n, 0n)).toBe(0n);
		expect(calculator.modulus(100n, 0n)).toBe(0n);
	});
});

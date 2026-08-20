import { readFileSync } from 'node:fs';

import { describe, expect, it } from 'vite-plus/test';

import { CurrencyManager, ExchangeRates, MoneyAggregator, MoneyCalculator, MoneyJson, MoneyManager } from '#money/index';

// Shared Go<->TS golden vectors. This is the TS half of the cross-runtime drift
// guard (plan 008); the Go half lives in
// pkg/hub/money/money/conformance_test.go and reads the same JSON. Numeric
// inputs/outputs are strings so JSON float precision never enters the
// comparison; error cases are matched by MoneyError.code identity, not message.

interface MoneyConformanceCase {
	op: string;
	args: string[];
	expected?: string;
	error?: string;
	note: string;
}

interface MoneyConformanceFile {
	cases: MoneyConformanceCase[];
}

const fixturePath = new URL('../../../../conformance/money.json', import.meta.url);
const fixture = JSON.parse(readFileSync(fixturePath, 'utf8')) as MoneyConformanceFile;
const calculator = MoneyCalculator.create();
const manager = MoneyManager.default();
const rates = ExchangeRates.create();
const aggregator = MoneyAggregator.create(manager);
const json = MoneyJson.default();

const runMoneyOp = (testCase: MoneyConformanceCase): string => {
	const { args, op } = testCase;

	switch (op) {
		case 'round':
			return calculator.round(BigInt(args[0] as string), Number(args[1])).toString();

		case 'isSafeAsNumber':
			return manager.create(BigInt(args[0] as string), 'USD').isSafeAsNumber() ? 'true' : 'false';

		case 'absolute':
			return calculator.absolute(BigInt(args[0] as string)).toString();

		case 'add':
			return calculator.add(BigInt(args[0] as string), BigInt(args[1] as string)).toString();

		case 'subtract':
			return calculator.subtract(BigInt(args[0] as string), BigInt(args[1] as string)).toString();

		case 'multiply':
			return calculator.multiply(BigInt(args[0] as string), BigInt(args[1] as string)).toString();

		case 'createFromFloat':
			return manager
				.createFromFloat(Number(args[0]), args[1] as string)
				.amount()
				.toString();

		case 'convertWithRate':
			return rates.convertAmountWithRate(BigInt(args[0] as string), Number(args[1]), Number(args[2]), Number(args[3])).toString();

		case 'avg': {
			const values = args.slice(1).map((raw) => manager.create(BigInt(raw), args[0] as string));

			return aggregator
				.avg(...values)
				.amount()
				.toString();
		}

		case 'unmarshalAmount':
			return json
				.unmarshal(args[0] as string)
				.amount()
				.toString();

		case 'unmarshalCurrency':
			return json.unmarshal(args[0] as string).currency().code;

		case 'resolveWithDefault': {
			// Per manager, so nothing leaks between cases.
			const currencies = CurrencyManager.default();

			currencies.setDefault(args[0] as string);

			return currencies.resolve(args[1] as string).code;
		}

		default:
			throw new Error(`unknown money conformance op: ${op}`);
	}
};

describe('money cross-runtime conformance', () => {
	expect(fixture.cases.length).toBeGreaterThan(0);

	for (const testCase of fixture.cases) {
		it(`${testCase.op}(${testCase.args.join(',')})`, () => {
			if (testCase.error) {
				let thrown: unknown;

				try {
					runMoneyOp(testCase);
				} catch (error) {
					thrown = error;
				}

				expect(thrown, `expected ${testCase.error}: ${testCase.note}`).toBeDefined();
				expect((thrown as { code?: string })?.code, testCase.note).toBe(testCase.error);

				return;
			}

			expect(runMoneyOp(testCase), testCase.note).toBe(testCase.expected);
		});
	}
});

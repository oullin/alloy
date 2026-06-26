import type { CurrencyDefinition } from '#money/currency/definition';
import { CurrencyManager } from '#money/currency/manager';
import { roundAwayFromZero } from '#money/internal/rounding';
import { MoneyValueBase, requireMoney, type MoneyValue } from '#money/money/core';
import { findTopLevelJsonNumber } from '#money/money/json-scanner';

import {
	ERR_CURRENCY_NOT_FOUND,
	ERR_INVALID_JSON_UNMARSHAL,
	ERR_JSON_MARSHAL_FUNC_NIL,
	ERR_JSON_UNMARSHAL_FUNC_NIL,
	ERR_NO_CURRENCY_INSTANCE,
	invalidJsonUnmarshalFrom,
} from '#money/errors';

type JsonMarshal = (money: MoneyValue) => string;

type JsonUnmarshal = (payload: string) => MoneyValue;

type JsonCurrency = () => CurrencyDefinition;

export class MoneyJson {
	public constructor(
		private marshalHandler: JsonMarshal | null = null,
		private unmarshalHandler: JsonUnmarshal | null = null,
		private currencyHandler: JsonCurrency | null = null,
	) {}

	public static default(): MoneyJson {
		return new MoneyJson();
	}

	public static withParser(unmarshal?: JsonUnmarshal | null, marshal?: JsonMarshal | null, currency?: JsonCurrency | null): MoneyJson {
		return new MoneyJson(marshal ?? null, unmarshal ?? null, currency ?? null);
	}

	public marshal(money: MoneyValue): string {
		if (this.marshalHandler !== null) {
			return this.marshalHandler(money);
		}

		return this.defaultMarshal(money);
	}

	public unmarshal(payload: string): MoneyValue {
		if (this.unmarshalHandler !== null) {
			return this.unmarshalHandler(payload);
		}

		return this.defaultUnmarshal(payload);
	}

	public setMarshal(handler: JsonMarshal | null): void {
		if (handler === null) {
			throw ERR_JSON_MARSHAL_FUNC_NIL;
		}

		this.marshalHandler = handler;
	}

	public setUnmarshal(handler: JsonUnmarshal | null): void {
		if (handler === null) {
			throw ERR_JSON_UNMARSHAL_FUNC_NIL;
		}

		this.unmarshalHandler = handler;
	}

	public setCurrency(handler: JsonCurrency | null): void {
		if (handler === null) {
			throw ERR_NO_CURRENCY_INSTANCE;
		}

		this.currencyHandler = handler;
	}

	private defaultMarshal(money: MoneyValue): string {
		const value = requireMoney(money);

		return `{"amount":${value.amount().toString()},"currency":"${value.currency().code}"}`;
	}

	private defaultUnmarshal(payload: string): MoneyValue {
		let parsed: unknown;

		try {
			parsed = JSON.parse(payload);
		} catch (error) {
			throw invalidJsonUnmarshalFrom(error);
		}

		if (parsed === null || typeof parsed !== 'object' || Array.isArray(parsed)) {
			throw ERR_INVALID_JSON_UNMARSHAL;
		}

		const record = parsed as Record<string, unknown>;
		const amount = this.extractJsonAmount(payload);

		let currency: CurrencyDefinition;

		if (record.currency === undefined || record.currency === '') {
			currency = this.defaultJsonCurrency();
		} else if (typeof record.currency === 'string') {
			const found = CurrencyManager.default().findByCode(record.currency);

			if (found === null) {
				throw ERR_CURRENCY_NOT_FOUND;
			}

			currency = found;
		} else {
			throw ERR_INVALID_JSON_UNMARSHAL;
		}

		return new MoneyValueBase(amount, currency) as MoneyValue;
	}

	private defaultJsonCurrency(): CurrencyDefinition {
		return (this.currencyHandler ?? (() => CurrencyManager.default().resolve('SGD')))();
	}

	private extractJsonAmount(payload: string): bigint {
		const raw = findTopLevelJsonNumber(payload, 'amount');

		if (raw === null) {
			return 0n;
		}

		if (/^-?\d+$/u.test(raw)) {
			return BigInt(raw);
		}

		const numeric = Number(raw);

		if (!Number.isFinite(numeric)) {
			throw ERR_INVALID_JSON_UNMARSHAL;
		}

		return roundAwayFromZero(numeric);
	}
}

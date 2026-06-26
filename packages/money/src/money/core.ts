export type { Amount } from '#money/calculator';
import { MoneyValue as MoneyValueFactory } from '#money/money/value';
import type { MoneyValue as MoneyValueInstance } from '#money/money/value';

export { MoneyManager } from '#money/money/manager';
export { MoneyValueBase } from '#money/money/value';
export type { MoneyJsonValue } from '#money/money/value';
export { requireMoney } from '#money/money/guards';

export const MoneyValue = MoneyValueFactory;

export type MoneyValue = MoneyValueInstance;

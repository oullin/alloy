export { MoneyAggregator } from '#money/money/aggregator';
export { MoneyConverter } from '#money/money/converter';
export { MoneyJson } from '#money/money/json';
import { MoneyValue as MoneyValueFactory } from '#money/money/core';
import type { MoneyValue as MoneyValueInstance } from '#money/money/core';

export { MoneyManager, MoneyValueBase } from '#money/money/core';
export type { MoneyJsonValue } from '#money/money/core';

export const MoneyValue = MoneyValueFactory;

export type MoneyValue = MoneyValueInstance;

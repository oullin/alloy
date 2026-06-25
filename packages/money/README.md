# @alloy/money

Object-oriented TypeScript port of the Alloy Go `api/money` module.

The package stores monetary amounts as exact minor-unit `bigint` values and
keeps creation, conversion, parsing, formatting, and aggregation behavior behind
classes such as `MoneyValue`, `MoneyManager`, `CurrencyManager`, and
`ExchangeRates`.

```ts
import { MoneyManager, MoneyValue } from '@alloy/money';

const manager = MoneyManager.default();
const subtotal = MoneyValue.fromUSD(1250n);
const tax = manager.createFromString('1.25', 'USD');
const total = manager.add(subtotal, tax);

total.display(); // "$13.75"
```

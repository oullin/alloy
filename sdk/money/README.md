# Money TypeScript

The TypeScript package lives in `sdk/money` and exposes `@alloy/sdk/money`.
It models monetary amounts as immutable values: a bigint amount in minor
units (cents) paired with an ISO currency definition. Arithmetic is
overflow-checked against the signed int64 range, and splitting/allocation
distributes remainders deterministically. Failures throw coded `MoneyError`
instances (e.g. `ERR_CURRENCY_MISMATCH`, `ERR_OVERFLOW`).

This is a private workspace package: it is consumed by sibling packages via
`workspace:*` and is never published to npm.

Money values are created through a `MoneyManager`, which also owns all
arithmetic:

```ts
import { MoneyManager, MoneyValue } from '@alloy/sdk/money';

const manager = MoneyManager.default();

const price = manager.create(1500n, 'SGD'); // S$15.00 from minor units
const fromString = manager.createFromString('12.34', 'USD');
const fromFactory = MoneyValue.fromUSD(500n); // static per-currency factory

const total = manager.add(fromString, fromFactory);

total.display(); // "$17.34"
total.amount(); // 1734n
total.asMajorUnits(); // 17.34

manager.split(price, 3).map((part) => part.amount()); // [500n, 500n, 500n]
manager.allocate(price, 70, 30).map((part) => part.amount()); // [1050n, 450n]
```

Currency conversion composes exchange rates with a converter:

```ts
import { ExchangeRates, MoneyConverter, MoneyManager } from '@alloy/sdk/money';

const manager = MoneyManager.default();
const rates = ExchangeRates.create().addRate('USD', 'EUR', 0.9);
const converter = MoneyConverter.create(manager.getCurrencyManager(), rates);

converter.convert(manager.create(1000n, 'USD'), 'EUR').display(); // "€9.00"
```

Parsing human input detects the currency from symbols or ISO codes:

```ts
import { MoneyParser } from '@alloy/sdk/money';

MoneyParser.create().parseAmount('$12.34'); // { amount: 12.34, currency: 'USD' }
```

## API overview

| Entry point | Main exports | Purpose |
| --- | --- | --- |
| `@alloy/sdk/money` | everything below | root export |
| `@alloy/sdk/money` | `MoneyValue`, `MoneyManager`, `MoneyAggregator`, `MoneyConverter`, `MoneyJson` | value type, arithmetic, aggregation, conversion, JSON codecs |
| `@alloy/sdk/money/calculator` | `MoneyCalculator`, `MAX_INT64`, `MIN_INT64` | overflow-safe bigint arithmetic |
| `@alloy/sdk/money/currency` | `CurrencyManager`, `CurrencyDefinition`, `CurrencyMap`, `DefaultCurrencyProvider`, `ISOCodePattern` | currency registry and definitions |
| `@alloy/sdk/money/errors` | `MoneyError`, `MoneyErrors`, `ERR_*` constants | coded error values |
| `@alloy/sdk/money/exchange` | `ExchangeRates`, `ExchangeConverter` | exchange-rate tables and guarded access |
| `@alloy/sdk/money/format` | `MoneyFormatter` | display formatting and major-unit conversion |
| `@alloy/sdk/money/parser` | `MoneyParser` | parsing money strings into amount + currency |

Acceptance tests live in `sdk/money/tests` and run with
`pnpm --filter @alloy/sdk/money test`.

# Money TypeScript

The TypeScript package lives in `sdk/money` and exposes `@hara/sdk-money`.
It models monetary amounts as immutable values: a bigint amount in minor
units (cents) paired with an ISO currency definition. Arithmetic is
overflow-checked against the signed int64 range, and splitting/allocation
distributes remainders deterministically. Failures throw coded `MoneyError`
instances (e.g. `ERR_CURRENCY_MISMATCH`, `ERR_OVERFLOW`).

This is a private workspace package. Sibling packages consume it via
`workspace:*`; everyone else consumes the tarball attached to a `ts/v*` GitHub
release. That is the only external channel — it is never published to a public
registry. Pin it by URL:

```jsonc
"dependencies": {
  "@hara/sdk-money": "https://github.com/oullin/alloy/releases/download/ts/v0.2.0/hara-sdk-money-0.2.0.tgz"
}
```

Money values are created through a `MoneyManager`, which also owns all
arithmetic:

```ts
import { MoneyManager, MoneyValue } from '@hara/sdk-money';

const manager = MoneyManager.default();

const price = manager.create(1500n, 'SGD'); // S$15.00 from minor units
const fromString = manager.createFromString('12.34', 'USD');
const fromFactory = MoneyValue.fromUSD(500n); // static per-currency factory

const total = manager.add(fromString, fromFactory);

total.display(); // "$17.34"
total.amount(); // 1734n
total.asMajorUnits(); // 17.34
total.displayCompact(); // "$17.34" — abbreviates to $1.3M / 750K / 4B at scale

manager.split(price, 3).map((part) => part.amount()); // [500n, 500n, 500n]
manager.allocate(price, 70, 30).map((part) => part.amount()); // [1050n, 450n]
```

Currency conversion composes exchange rates with a converter:

```ts
import { ExchangeRates, MoneyConverter, MoneyManager } from '@hara/sdk-money';

const manager = MoneyManager.default();
const rates = ExchangeRates.create().addRate('USD', 'EUR', 0.9);
const converter = MoneyConverter.create(manager.getCurrencyManager(), rates);

converter.convert(manager.create(1000n, 'USD'), 'EUR').display(); // "€9.00"
```

Parsing human input detects the currency from symbols or ISO codes:

```ts
import { MoneyParser } from '@hara/sdk-money';

MoneyParser.create().parseAmount('$12.34'); // { amount: 12.34, currency: 'USD' }
```

## The default currency

A code that cannot be resolved falls back to the manager's default currency —
**SGD** as shipped. Change it per manager:

```ts
import { CurrencyManager } from '@hara/sdk-money/currency';

const currencies = CurrencyManager.default();

currencies.setDefault('AED'); // trimmed and upper-cased; throws ERR_CURRENCY_NOT_FOUND if unknown
currencies.resolve('NOPE').code; // "AED"
```

The default belongs to the manager, not the module, so this takes effect on that
instance immediately — including one built before the call — and leaves every
other manager alone. The Go twin's `(*currency.Manager).SetDefault` behaves
identically, and `conformance/money.json` pins that.

To change the currency `MoneyJson` assumes for a payload carrying no `currency`,
pass a handler to `MoneyJson.setCurrency` rather than changing a manager.

## API overview

| Entry point                  | Main exports                                                                                        | Purpose                                                      |
| ---------------------------- | --------------------------------------------------------------------------------------------------- | ------------------------------------------------------------ |
| `@hara/sdk-money`            | everything below                                                                                    | root export                                                  |
| `@hara/sdk-money`            | `MoneyValue`, `MoneyManager`, `MoneyAggregator`, `MoneyConverter`, `MoneyJson`                      | value type, arithmetic, aggregation, conversion, JSON codecs |
| `@hara/sdk-money/calculator` | `MoneyCalculator`, `MAX_INT64`, `MIN_INT64`                                                         | overflow-safe bigint arithmetic                              |
| `@hara/sdk-money/currency`   | `CurrencyManager`, `CurrencyDefinition`, `CurrencyMap`, `DefaultCurrencyProvider`, `ISOCodePattern` | currency registry and definitions                            |
| `@hara/sdk-money/errors`     | `MoneyError`, `MoneyErrors`, `ERR_*` constants                                                      | coded error values                                           |
| `@hara/sdk-money/exchange`   | `ExchangeRates`, `ExchangeConverter`                                                                | exchange-rate tables and guarded access                      |
| `@hara/sdk-money/format`     | `MoneyFormatter`                                                                                    | display formatting and major-unit conversion                 |
| `@hara/sdk-money/parser`     | `MoneyParser`                                                                                       | parsing money strings into amount + currency                 |

Acceptance tests live in `sdk/money/tests` and run with
`pnpm --filter @hara/sdk-money test`.

package money

import (
	"alloy.dev/go/money/calculator"
	"alloy.dev/go/money/currency"
)

// Amount is a data structure that stores the amount being used for calculations.
type Amount = calculator.Amount

// Value represents a monetary value with an amount and a currency.
type Value struct {
	amount   Amount               `db:"amount"`
	currency *currency.Definition `db:"currency"`
}

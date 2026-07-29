package money

import (
	"fmt"

	"hara.sh/alloy/money/currency"
	"hara.sh/alloy/money/exception"
	"hara.sh/alloy/money/exchange"
)

type Converter struct {
	currencies *currency.Manager
	exchange   *exchange.Rates
}

// NewConverter creates a new Converter with pre-loaded exchange rates
func NewConverter(currencies *currency.Manager, ex *exchange.Rates) (*Converter, error) {
	if currencies == nil {
		return nil, exception.ErrNoCurrencyManager
	}

	if ex == nil || ex.IsInvalid() {
		return nil, exception.ErrInvalidExchangeRate
	}

	return &Converter{
		currencies: currencies,
		exchange:   ex,
	}, nil
}

// Convert converts a Value object to another currency using the pre-loaded exchange rates
func (c *Converter) Convert(money *Value, toCurrency string) (*Value, error) {
	if c == nil || c.exchange == nil || c.exchange.IsInvalid() {
		return nil, exception.ErrInvalidExchangeRate
	}

	if err := ensureMoneyProvided(money); err != nil {
		return nil, err
	}

	amount, err := money.Amount()

	if err != nil {
		return nil, err
	}

	fromCurrency, err := money.Currency()

	if err != nil {
		return nil, err
	}

	target := c.currencies.FindByCode(toCurrency)

	if target == nil {
		return nil, fmt.Errorf("currency %s not found: %w", toCurrency, exception.ErrCurrencyNotFound)
	}

	convertedAmount, err := c.exchange.ConvertAmount(
		amount,
		fromCurrency.Code,
		fromCurrency.Fraction,
		target.Code,
		target.Fraction,
	)

	if err != nil {
		return nil, err
	}

	return &Value{
		amount:   convertedAmount,
		currency: target,
	}, nil
}

// ConvertWithRate converts a Value object to another currency using a specific rate
func (c *Converter) ConvertWithRate(money *Value, toCurrency string, rate float64) (*Value, error) {
	if c == nil || c.exchange == nil || c.exchange.IsInvalid() {
		return nil, exception.ErrInvalidExchangeRate
	}

	if err := ensureMoneyProvided(money); err != nil {
		return nil, err
	}

	amount, err := money.Amount()

	if err != nil {
		return nil, err
	}

	fromCurrency, err := money.Currency()

	if err != nil {
		return nil, err
	}

	target := c.currencies.FindByCode(toCurrency)

	if target == nil {
		return nil, fmt.Errorf("currency %s not found: %w", toCurrency, exception.ErrCurrencyNotFound)
	}

	convertedAmount, err := c.exchange.ConvertAmountWithRate(
		amount,
		fromCurrency.Fraction,
		target.Fraction,
		rate,
	)

	if err != nil {
		return nil, err
	}

	return &Value{
		amount:   convertedAmount,
		currency: target,
	}, nil
}

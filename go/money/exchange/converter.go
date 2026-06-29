package exchange

import (
	"alloy.dev/go/money/exception"
)

// Converter provides a simple interface for currency conversion with a fixed exchange
type Converter struct {
	exchange *Rates
}

// NewConverter creates a new Converter instance with the provided Rates.
func NewConverter(exchange *Rates) (*Converter, error) {
	if exchange == nil {
		return nil, exception.ErrNoConverterProvided
	}

	return &Converter{
		exchange: exchange,
	}, nil
}

// GetExchange returns the underlying Rates instance.
func (c *Converter) GetExchange() (*Rates, error) {
	if c == nil {
		return nil, exception.ErrNoConverterProvided
	}

	if c.exchange == nil {
		return nil, exception.ErrInvalidExchangeRate
	}

	return c.exchange, nil
}

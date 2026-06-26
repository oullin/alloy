package currency

import (
	"database/sql/driver"
	"fmt"

	"github.com/oullin/alloy/api/money/exception"
	"github.com/oullin/alloy/api/money/format"
)

// Definition represents the formatting rules for a specific currency.
type Definition struct {
	// Decimal is the decimal separator (e.g., ".").
	Decimal string
	// A Thousand is the thousand separators (e.g., ",").
	Thousand string
	// Code is the ISO 4217 currency code (e.g. "SGD").
	Code string
	// Fraction is the number of fraction digits (e.g. 2 for cents).
	Fraction int
	// NumericCode is the numeric code of the currency (e.g., "702" for SGD).
	NumericCode string
	// Grapheme is the currency symbol (e.g., "$").
	Grapheme string
	// Template is the formatting template (e.g., "$1").
	Template string
}

// Formatter returns currency formatter representing
func (c *Definition) Formatter() *format.Renderer {
	return format.NewFormatter(
		c.Fraction,
		c.Decimal,
		c.Thousand,
		c.Grapheme,
		c.Template,
	)
}

// Equals checks if two currencies are equal based on their code.
func (c *Definition) Equals(oc *Definition) bool {
	if c == nil || oc == nil {
		return c == oc
	}

	return c.Code == oc.Code
}

// Get returns the currency if it exists, otherwise returns an error.
func (c *Definition) Get() (*Definition, error) {
	if c == nil {
		return nil, exception.ErrCurrencyNotFound
	}

	return c, nil
}

// DbValue implements driver.Valuer to serialize a Definition code into a string for saving to a database
func (c *Definition) DbValue() (driver.Value, error) {
	if c == nil {
		return nil, exception.ErrCurrencyNotFound
	}

	return c.Code, nil
}

// DbScan implements sql.Scanner to deserialize a Definition from a string value read from a database
func (c *Definition) DbScan(src any) error {
	if c == nil {
		return exception.ErrCurrencyNotFound
	}

	var val *Definition

	var code string

	switch v := src.(type) {
	case string:
		code = v
	case []byte:
		code = string(v)
	default:
		return fmt.Errorf("%T is not a supported type for a Definition (store the Definition.Code value as a string only)", src)
	}

	currencies := NewCurrenciesMap()

	if val = currencies.FindByCode(code); val == nil {
		return fmt.Errorf("currency(%#v) returned nil", code)
	}

	// copy the value
	*c = *val

	return nil
}

package money

import (
	"alloy.dev/foundation/money/currency"
	"alloy.dev/foundation/money/exception"
)

// Currency returns the currency of the money object.
func (m *Value) Currency() (*currency.Definition, error) {
	if err := ensureMoneyProvided(m); err != nil {
		return nil, err
	}

	return m.currency, nil
}

// Amount returns the integer amount of the money object.
func (m *Value) Amount() (int64, error) {
	if err := ensureMoneyProvided(m); err != nil {
		return 0, err
	}

	return m.amount, nil
}

// AssertSameCurrency returns an error if the provided Value object does not have the same currency.
func (m *Value) AssertSameCurrency(om *Value) error {
	if err := ensureMoneyProvided(m); err != nil {
		return err
	}

	if om == nil {
		return exception.ErrNoMoneyProvided
	}

	ok, err := m.SameCurrency(om)

	if err != nil {
		return err
	}

	if !ok {
		return exception.ErrCurrencyMismatch
	}

	return nil
}

// SameCurrency checks if the provided Value object has the same currency.
func (m *Value) SameCurrency(om *Value) (bool, error) {
	if err := ensureMoneyProvided(m); err != nil {
		return false, err
	}

	if om == nil {
		return false, exception.ErrNoMoneyProvided
	}

	return m.currency.Equals(om.currency), nil
}

// CompareAmount compares the amount of two Value objects.
// Returns 1 if m > om, -1 if m < om, 0 if equal.
func (m *Value) CompareAmount(om *Value) (int, error) {
	if err := m.AssertSameCurrency(om); err != nil {
		return 0, err
	}

	switch {
	case m.amount > om.amount:
		return 1, nil
	case m.amount < om.amount:
		return -1, nil
	}

	return 0, nil
}

// Equals checks if the amount of the provided Value object is equal to this one.
func (m *Value) Equals(om *Value) (bool, error) {
	result, err := m.CompareAmount(om)

	if err != nil {
		return false, err
	}

	return result == 0, nil
}

// GreaterThan checks if the amount is greater than the provided Value object.
func (m *Value) GreaterThan(om *Value) (bool, error) {
	result, err := m.CompareAmount(om)

	if err != nil {
		return false, err
	}

	return result == 1, nil
}

// GreaterThanOrEqual checks if the amount is greater than or equal to the provided Value object.
func (m *Value) GreaterThanOrEqual(om *Value) (bool, error) {
	result, err := m.CompareAmount(om)

	if err != nil {
		return false, err
	}

	return result >= 0, nil
}

// LessThan checks if the amount is less than the provided Value object.
func (m *Value) LessThan(om *Value) (bool, error) {
	result, err := m.CompareAmount(om)

	if err != nil {
		return false, err
	}

	return result == -1, nil
}

// LessThanOrEqual checks if the amount is less than or equal to the provided Value object.
func (m *Value) LessThanOrEqual(om *Value) (bool, error) {
	result, err := m.CompareAmount(om)

	if err != nil {
		return false, err
	}

	return result <= 0, nil
}

// IsZero checks if the amount is zero.
func (m *Value) IsZero() (bool, error) {
	if err := ensureMoneyProvided(m); err != nil {
		return false, err
	}

	return m.amount == 0, nil
}

// IsPositive checks if the amount is positive.
func (m *Value) IsPositive() (bool, error) {
	if err := ensureMoneyProvided(m); err != nil {
		return false, err
	}

	return m.amount > 0, nil
}

// IsNegative checks if the amount is negative.
func (m *Value) IsNegative() (bool, error) {
	if err := ensureMoneyProvided(m); err != nil {
		return false, err
	}

	return m.amount < 0, nil
}

// Display returns the formatted string representation of the money.
func (m *Value) Display() (string, error) {
	if err := ensureMoneyProvided(m); err != nil {
		return "", err
	}

	c, err := m.currency.Get()

	if err != nil {
		return "", err
	}

	return c.Formatter().Format(m.amount), nil
}

// AsMajorUnits lets represent Value struct as subunits (float64) in a given Currency value
func (m *Value) AsMajorUnits() (float64, error) {
	if err := ensureMoneyProvided(m); err != nil {
		return 0, err
	}

	c, err := m.currency.Get()

	if err != nil {
		return 0, err
	}

	return c.Formatter().ToMajorUnits(m.amount), nil
}

// Compare function compares two money of the same type
//
//	if m.amount > om.amount returns (1, nil)
//	if m.amount == om.amount returns (0, nil
//	if m.amount == om.amount returns (0, nil)
//	if m.amount < om.amount returns (-1, nil)
//
// If CompareAmount moneys from distinct currency, return (m.amount, ErrCurrencyMismatch)
func (m *Value) Compare(om *Value) (int, error) {
	if err := ensureMoneyProvided(m); err != nil {
		return 0, err
	}

	result, err := m.CompareAmount(om)

	if err != nil {
		return int(m.amount), err
	}

	return result, nil
}

// UnmarshalJSON is an implementation of json.Unmarshaller
func (m *Value) UnmarshalJSON(b []byte) error {
	if m == nil {
		return exception.ErrNoMoneyProvided
	}

	return NewJson().Unmarshal(m, b)
}

// MarshalJSON is an implementation of json.Marshaller
func (m *Value) MarshalJSON() ([]byte, error) {
	if m == nil {
		return nil, exception.ErrNoMoneyProvided
	}

	return NewJson().Marshal(*m)
}

func ensureMoneyProvided(m *Value) error {
	if m == nil {
		return exception.ErrNoMoneyProvided
	}

	if m.currency == nil {
		return exception.ErrNoCurrencyInstance
	}

	return nil
}

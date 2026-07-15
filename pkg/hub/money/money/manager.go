package money

import (
	"math"
	"strings"

	"github.com/oullin/alloy/pkg/hub/money/calculator"
	"github.com/oullin/alloy/pkg/hub/money/currency"
	"github.com/oullin/alloy/pkg/hub/money/exception"
	"github.com/oullin/alloy/pkg/hub/money/parser"
)

// Manager handles Value creation and operations with a dependency-injected currency manager and calculator.
type Manager struct {
	parser          *parser.Reader
	currencyManager *currency.Manager
	calculator      *calculator.Engine
}

// NewManager creates a new Value Manager with the default currency manager.
func NewManager() *Manager {
	return &Manager{
		currencyManager: currency.NewManager(),
		calculator:      calculator.NewCalculator(),
		parser:          parser.NewParser(),
	}
}

// NewManagerWith creates a new Value Manager with the provided currency manager.
// Returns an error if the currency manager is nil.
func NewManagerWith(cm *currency.Manager) (*Manager, error) {
	if cm == nil {
		return nil, exception.ErrNoCurrencyManager
	}

	return &Manager{
		currencyManager: cm,
		calculator:      calculator.NewCalculator(),
		parser:          parser.NewParser(),
	}, nil
}

// Create creates a new Value instance with the given amount and currency code.
func (mm *Manager) Create(amount int64, code string) *Value {
	return &Value{
		amount:   amount,
		currency: mm.currencyManager.Resolve(code),
	}
}

// CreateFromFloat creates a Value instance from a float64 amount.
// Uses banker's rounding (round half to even) for precision (e.g. 12.345 SGD -> 1235 cents, 12.344 SGD -> 1234 cents).
//
// WARNING: Floats are imprecise and should NOT be used for exact financial calculations.
// Use this function only for user input conversion or external system integration.
// For precise amounts, use Create() with integer cents or CreateFromString() with exact decimal strings.
func (mm *Manager) CreateFromFloat(amount float64, code string) *Value {
	curr := mm.currencyManager.Resolve(code)
	currencyDecimals := math.Pow10(curr.Fraction)

	scaled := amount * currencyDecimals

	return mm.Create(int64(math.Round(scaled)), code)
}

// CreateFromString creates a Value instance from a string representation.
// Accepts decimal strings like "12.34", "-99.99", "100", etc.
// This function provides exact precision for financial calculations, unlike CreateFromFloat.
//
// Examples:
//   - CreateFromString("12.34", "SGD") -> 1234 cents
//   - CreateFromString("-99.99", "SGD") -> -9999 cents
//   - CreateFromString("100", "JPY") -> 100 yen (no decimal)
//
// Returns an error if:
//   - The string is not a valid decimal number
//   - The decimal places exceed the currency's fraction
func (mm *Manager) CreateFromString(amount string, code string) (*Value, error) {
	amount = strings.TrimSpace(amount)

	if amount == "" {
		return nil, exception.ErrEmptyAmountString
	}

	curr := mm.currencyManager.Resolve(code)
	fraction := curr.Fraction

	// Handle sign extraction
	amount, negative, err := mm.parser.ParseStringSign(amount)

	if err != nil {
		return nil, err
	}

	// Parse the amount string into int64
	value, err := mm.parser.ParseAmountString(amount, fraction, negative)

	if err != nil {
		return nil, err
	}

	return mm.Create(value, code), nil
}

// GetCurrencyManager returns the currency manager used by this MoneyManager.
func (mm *Manager) GetCurrencyManager() *currency.Manager {
	return mm.currencyManager
}

// Add returns a new Value struct representing the sum of the provided Value values.
func (mm *Manager) Add(m *Value, ms ...*Value) (*Value, error) {
	if err := ensureMoneyProvided(m); err != nil {
		return nil, err
	}

	if len(ms) == 0 {
		return mm.Create(m.amount, m.currency.Code), nil
	}

	result := mm.Create(m.amount, m.currency.Code)

	for _, m2 := range ms {
		if err := m.AssertSameCurrency(m2); err != nil {
			return nil, err
		}

		amount, err := mm.calculator.SafeAdd(result.amount, m2.amount)

		if err != nil {
			return nil, err
		}

		result.amount = amount
	}

	return result, nil
}

// Subtract returns a new Value struct representing the difference of the provided Value values.
func (mm *Manager) Subtract(m *Value, ms ...*Value) (*Value, error) {
	if err := ensureMoneyProvided(m); err != nil {
		return nil, err
	}

	if len(ms) == 0 {
		return mm.Create(m.amount, m.currency.Code), nil
	}

	result := mm.Create(m.amount, m.currency.Code)

	for _, m2 := range ms {
		if err := m.AssertSameCurrency(m2); err != nil {
			return nil, err
		}

		amount, err := mm.calculator.SafeSubtract(result.amount, m2.amount)

		if err != nil {
			return nil, err
		}

		result.amount = amount
	}

	return result, nil
}

// Multiply returns a new Value struct with the amount multiplied by the given values.
func (mm *Manager) Multiply(m *Value, values ...int64) (*Value, error) {
	if err := ensureMoneyProvided(m); err != nil {
		return nil, err
	}

	if len(values) == 0 {
		return nil, exception.ErrNoMultipliersProvided
	}

	result, err := mm.calculator.SafeMultiply(m.amount, values...)

	if err != nil {
		return nil, err
	}

	return mm.Create(result, m.currency.Code), nil
}

// Absolute returns a new Value struct with the absolute value of the amount.
func (mm *Manager) Absolute(m *Value) (*Value, error) {
	if err := ensureMoneyProvided(m); err != nil {
		return nil, err
	}

	amount := mm.calculator.Absolute(m.amount)

	return mm.Create(amount, m.currency.Code), nil
}

// Negative returns a new Value struct with the negated amount.
func (mm *Manager) Negative(m *Value) (*Value, error) {
	if err := ensureMoneyProvided(m); err != nil {
		return nil, err
	}

	return mm.Create(-m.amount, m.currency.Code), nil
}

// Round returns a new Value struct with the amount rounded according to the currency's fraction.
func (mm *Manager) Round(m *Value) (*Value, error) {
	if err := ensureMoneyProvided(m); err != nil {
		return nil, err
	}

	rounded := mm.calculator.Round(m.amount, m.currency.Fraction)

	return mm.Create(rounded, m.currency.Code), nil
}

// Split returns a slice of Value structs with the amount split into n equal parts.
// After division leftover pennies will be distributed round-robin amongst the parties.
// This means that parties listed first will likely receive more pennies than ones that are listed later.
func (mm *Manager) Split(m *Value, n int) ([]*Value, error) {
	if err := ensureMoneyProvided(m); err != nil {
		return nil, err
	}

	if n <= 0 {
		return nil, exception.ErrInvalidSplit
	}

	quotient := mm.calculator.Divide(m.amount, int64(n))
	remainder := mm.calculator.Modulus(m.amount, int64(n))

	ms := make([]*Value, n)

	for i := range n {
		ms[i] = mm.Create(quotient, m.currency.Code)
	}

	// Distribute the remainder
	absRemainder := mm.calculator.Absolute(remainder)

	increment := int64(1)

	if m.amount < 0 {
		increment = -1
	}

	for i := 0; i < int(absRemainder); i++ {
		ms[i].amount = mm.calculator.Add(ms[i].amount, increment)
	}

	return ms, nil
}

// Allocate returns a slice of Value structs with the amount split according to the given ratios.
// It lets split money by given ratios without losing pennies, and as Split operations distributes
// leftover pennies amongst the parties with a round-robin principle.
func (mm *Manager) Allocate(m *Value, rs ...int) ([]*Value, error) {
	if err := ensureMoneyProvided(m); err != nil {
		return nil, err
	}

	if len(rs) == 0 {
		return nil, exception.ErrNoRatiosProvided
	}

	// Calculate sum of ratios
	var sum int64

	for _, r := range rs {
		if r < 0 {
			return nil, exception.ErrNegativeRatios
		}

		if int64(r) > (math.MaxInt64 - sum) {
			return nil, exception.ErrRatiosExceedMaxInt
		}

		sum += int64(r)
	}

	var total int64
	ms := make([]*Value, 0, len(rs))

	for _, r := range rs {
		amount := mm.calculator.Allocate(m.amount, int64(r), sum)

		party := mm.Create(amount, m.currency.Code)
		ms = append(ms, party)
		total = mm.calculator.Add(total, amount)
	}

	// If the sum of all ratios is zero, return zeros
	if sum == 0 {
		return ms, nil
	}

	// Calculate leftover value and distribute to first parties
	leftover := mm.calculator.Subtract(m.amount, total)
	increment := int64(1)

	if leftover < 0 {
		increment = -1
	}

	for i := 0; leftover != 0 && i < len(ms); i++ {
		ms[i].amount = mm.calculator.Add(ms[i].amount, increment)
		leftover = mm.calculator.Subtract(leftover, increment)
	}

	return ms, nil
}

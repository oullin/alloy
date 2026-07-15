package exchange

import (
	"math"
	"math/big"
	"sync"

	"github.com/oullin/alloy/pkg/hub/money/exception"
)

// Rates provides currency conversion functionality.
// Safe for concurrent use by multiple goroutines.
type Rates struct {
	mu    sync.RWMutex
	rates map[string]map[string]float64
}

// rateScaleExponent stores exchange rates as scaled integers with 12 decimal
// places, so conversions retain exact amount precision across the full int64
// range, including amounts above 2^53 minor units.
const rateScaleExponent = 12

// NewExchange creates a new Rates instance.
func NewExchange() *Rates {
	return &Rates{
		rates: make(map[string]map[string]float64),
	}
}

// IsValid checks if the Rates is properly initialised with a valid rates map
func (e *Rates) IsValid() bool {
	if e == nil {
		return false
	}

	e.mu.RLock()

	defer e.mu.RUnlock()

	return e.rates != nil && len(e.rates) > 0
}

// IsInvalid checks if the Rates is not properly initialised or has no rates
func (e *Rates) IsInvalid() bool {
	return !e.IsValid()
}

// AddRate adds or updates a conversion rate between two tables.
func (e *Rates) AddRate(baseCurrency, counterCurrency string, rate float64) error {
	if e == nil || rate <= 0 {
		return exception.ErrInvalidExchangeRate
	}

	e.mu.Lock()

	defer e.mu.Unlock()

	if e.rates[baseCurrency] == nil {
		e.rates[baseCurrency] = make(map[string]float64)
	}

	e.rates[baseCurrency][counterCurrency] = rate

	return nil
}

// GetRate retrieves the conversion rate between two tables.
// If a direct rate is not found, it attempts to use the inverse rate.
func (e *Rates) GetRate(baseCurrency, counterCurrency string) (float64, error) {
	if e == nil {
		return -1, exception.ErrInvalidExchangeRate
	}

	if baseCurrency == counterCurrency {
		return 1.0, nil
	}

	e.mu.RLock()

	defer e.mu.RUnlock()

	if e.rates[baseCurrency] != nil {
		if rate, ok := e.rates[baseCurrency][counterCurrency]; ok {
			return rate, nil
		}
	}

	// Try inverse rate
	if e.rates[counterCurrency] != nil {
		if rate, ok := e.rates[counterCurrency][baseCurrency]; ok {
			return 1.0 / rate, nil
		}
	}

	return 0, exception.ErrCurrencyConversionNotFound
}

// ConvertAmount converts an amount from one currency to another using the stored exchange rates.
func (e *Rates) ConvertAmount(amount int64, fromCurrencyCode string, fromFraction int, toCurrencyCode string, toFraction int) (int64, error) {
	if e == nil {
		return -1, exception.ErrInvalidExchangeRate
	}

	if fromCurrencyCode == toCurrencyCode {
		return amount, nil
	}

	rate, err := e.GetRate(fromCurrencyCode, toCurrencyCode)

	if err != nil {
		return 0, err
	}

	return convertScaledAmount(amount, fromFraction, toFraction, rate)
}

// ConvertAmountWithRate converts an amount from one currency to another using a provided exchange rate.
func (e *Rates) ConvertAmountWithRate(amount int64, fromFraction int, toFraction int, rate float64) (int64, error) {
	if e == nil || rate <= 0 {
		return 0, exception.ErrInvalidExchangeRate
	}

	return convertScaledAmount(amount, fromFraction, toFraction, rate)
}

// convertScaledAmount converts a minor-unit amount using a rate represented as
// a scale-12 integer. The calculation uses exact integer arithmetic, preserving
// precision for the full int64 amount range, including values above 2^53. It
// returns exception.ErrOverflow when the scaled rate or converted result cannot
// be represented by int64.
func convertScaledAmount(amount int64, fromFraction, toFraction int, rate float64) (int64, error) {
	// Negative fractions would silently corrupt the math below: big.Int.Exp
	// returns 1 for a negative exponent instead of failing.
	if fromFraction < 0 || toFraction < 0 {
		return 0, exception.ErrInvalidAmountFraction
	}

	scaledRateFloat := rate * math.Pow10(rateScaleExponent)

	// float64(math.MaxInt64) rounds up to 2^63, which int64 cannot hold, so
	// the comparison must be >= — a scaled rate of exactly 2^63 would slip
	// past > and wrap in the int64 conversion below.
	if math.IsNaN(scaledRateFloat) || math.Round(scaledRateFloat) >= float64(math.MaxInt64) {
		return 0, exception.ErrOverflow
	}

	rateScaled := int64(math.Round(scaledRateFloat))

	ten := big.NewInt(10)
	numerator := new(big.Int).Mul(big.NewInt(amount), big.NewInt(rateScaled))
	numerator.Mul(numerator, new(big.Int).Exp(ten, big.NewInt(int64(toFraction)), nil))
	denominator := new(big.Int).Exp(ten, big.NewInt(int64(rateScaleExponent+fromFraction)), nil)

	sign := numerator.Sign()
	absNumerator := new(big.Int).Abs(numerator)
	quotient, remainder := new(big.Int), new(big.Int)
	quotient.QuoRem(absNumerator, denominator, remainder)

	twiceRemainder := new(big.Int).Lsh(remainder, 1)

	if twiceRemainder.Cmp(denominator) >= 0 {
		quotient.Add(quotient, big.NewInt(1))
	}

	if sign < 0 {
		quotient.Neg(quotient)
	}

	if quotient.Cmp(big.NewInt(math.MaxInt64)) > 0 || quotient.Cmp(big.NewInt(math.MinInt64)) < 0 {
		return 0, exception.ErrOverflow
	}

	return quotient.Int64(), nil
}

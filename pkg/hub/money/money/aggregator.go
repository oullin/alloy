package money

import (
	"hara.sh/alloy/money/exception"
)

// Aggregator performs aggregate operations on Value objects using a shared Manager.
type Aggregator struct {
	manager *Manager
}

// NewAggregator creates a new Aggregator with the provided Manager.
func NewAggregator(manager *Manager) *Aggregator {
	return &Aggregator{manager: manager}
}

// Sum adds multiple Value objects together and returns the result.
// All Value objects must have the same currency.
//
// Preconditions:
//   - All Value arguments must be non-nil (panics if nil)
func (a *Aggregator) Sum(moneys ...*Value) (*Value, error) {
	if a == nil || a.manager == nil {
		return nil, exception.ErrInvalidAggregatorProvider
	}

	if len(moneys) == 0 {
		return nil, exception.ErrNoMoneyProvided
	}

	if len(moneys) == 1 {
		return moneys[0], nil
	}

	// Use the first money as the base and add all others
	result := moneys[0]
	remaining := moneys[1:]

	return a.manager.Add(result, remaining...)
}

// Min returns the Value object with the minimum amount from the given slice.
// All Value objects must have the same currency.
//
// Preconditions:
//   - All Value arguments must be non-nil (panics if nil)
func (a *Aggregator) Min(moneys ...*Value) (*Value, error) {
	if a == nil || a.manager == nil {
		return nil, exception.ErrInvalidAggregatorProvider
	}

	if len(moneys) == 0 {
		return nil, exception.ErrNoMoneyProvided
	}

	if len(moneys) == 1 {
		return moneys[0], nil
	}

	money := moneys[0]

	for _, m := range moneys[1:] {
		if err := money.AssertSameCurrency(m); err != nil {
			return nil, err
		}

		if m.amount < money.amount {
			money = m
		}
	}

	return money, nil
}

// Max returns the Value object with the maximum amount from the given slice.
// All Value objects must have the same currency.
//
// Preconditions:
//   - All Value arguments must be non-nil (panics if nil)
func (a *Aggregator) Max(moneys ...*Value) (*Value, error) {
	if a == nil || a.manager == nil {
		return nil, exception.ErrInvalidAggregatorProvider
	}

	if len(moneys) == 0 {
		return nil, exception.ErrNoMoneyProvided
	}

	if len(moneys) == 1 {
		return moneys[0], nil
	}

	money := moneys[0]

	for _, m := range moneys[1:] {
		if err := money.AssertSameCurrency(m); err != nil {
			return nil, err
		}

		if m.amount > money.amount {
			money = m
		}
	}

	return money, nil
}

// Avg returns the average of multiple Value objects.
// All Value objects must have the same currency.
//
// The result is sum.amount / len(moneys) with any fractional minor unit
// rounded half away from zero — the same policy as Round and CreateFromFloat,
// and the same semantics as the TS twin (sdk/money/src/money/aggregator.ts).
// For example, the average of $1.00 and $1.01 is $1.01, the average of
// -$1.00 and -$1.01 is -$1.01, and the average of $0.02 and $0.01 is $0.02
// (rounded from $0.015).
//
// Preconditions:
//   - All Value arguments must be non-nil (panics if nil)
func (a *Aggregator) Avg(moneys ...*Value) (*Value, error) {
	if a == nil || a.manager == nil {
		return nil, exception.ErrInvalidAggregatorProvider
	}

	if len(moneys) == 0 {
		return nil, exception.ErrNoMoneyProvided
	}

	if len(moneys) == 1 {
		return moneys[0], nil
	}

	sum, err := a.Sum(moneys...)

	if err != nil {
		return nil, err
	}

	count := int64(len(moneys))
	avgAmount := sum.amount / count
	remainder := sum.amount % count

	if remainder < 0 {
		remainder = -remainder
	}

	// A leftover of half a minor unit or more rounds away from zero.
	if remainder*2 >= count {
		if sum.amount < 0 {
			avgAmount--
		} else {
			avgAmount++
		}
	}

	return &Value{
		amount:   avgAmount,
		currency: sum.currency,
	}, nil
}

//var defaultAggregator = NewAggregator(NewManager())
//
//// Sum adds multiple Value objects together and returns the result.
//// All Value objects must have the same currency.
////
//// Preconditions:
////   - All Value arguments must be non-nil (panics if nil)
//func Sum(moneys ...*Value) (*Value, error) {
//	return defaultAggregator.Sum(moneys...)
//}
//
//// Min returns the Value object with the minimum amount from the given slice.
//// All Value objects must have the same currency.
////
//// Preconditions:
////   - All Value arguments must be non-nil (panics if nil)
//func Min(moneys ...*Value) (*Value, error) {
//	return defaultAggregator.Min(moneys...)
//}
//
//// Max returns the Value object with the maximum amount from the given slice.
//// All Value objects must have the same currency.
////
//// Preconditions:
////   - All Value arguments must be non-nil (panics if nil)
//func Max(moneys ...*Value) (*Value, error) {
//	return defaultAggregator.Max(moneys...)
//}
//
//// Avg returns the average of multiple Value objects.
//// All Value objects must have the same currency.
////
//// The result is sum.amount / len(moneys) with any fractional minor unit
//// rounded half away from zero — the same policy as Round and CreateFromFloat.
////
//// Preconditions:
////   - All Value arguments must be non-nil (panics if nil)
//func Avg(moneys ...*Value) (*Value, error) {
//	return defaultAggregator.Avg(moneys...)
//}

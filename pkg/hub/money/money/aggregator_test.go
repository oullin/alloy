package money

import (
	"errors"
	"math"
	"testing"

	"hara.sh/alloy/money/currency"
	"hara.sh/alloy/money/exception"
	"hara.sh/alloy/money/exchange"
	testutil "hara.sh/alloy/money/tests"
)

func TestMoneyConverter_UnknownTargetCurrency(t *testing.T) {
	ex := exchange.NewExchange()
	testutil.TestRequireNoErr(t, ex.AddRate(currency.USD, currency.EUR, 1))

	converter := newTestConverter(t, currency.NewManager(), ex)
	usd := NewManager().Create(10, currency.USD)

	if _, err := converter.Convert(usd, "ZZZ"); err == nil || !errors.Is(err, exception.ErrCurrencyNotFound) {
		t.Fatalf("Convert() error = %v, want ErrCurrencyNotFound", err)
	}

	if _, err := converter.ConvertWithRate(usd, "ZZZ", 1); err == nil || !errors.Is(err, exception.ErrCurrencyNotFound) {
		t.Fatalf("ConvertWithRate() error = %v, want ErrCurrencyNotFound", err)
	}
}

func TestMoneyConverter_NilExchange(t *testing.T) {
	converter := &Converter{currencies: currency.NewManager(), exchange: nil}

	if _, err := converter.Convert(NewManager().Create(1, currency.USD), currency.EUR); !errors.Is(err, exception.ErrInvalidExchangeRate) {
		t.Fatalf("Convert() error = %v, want ErrInvalidExchangeRate", err)
	}
}

func TestAggregatorNilProviderErrors(t *testing.T) {
	var a *Aggregator

	if _, err := a.Sum(NewManager().Create(1, "SGD")); !errors.Is(err, exception.ErrInvalidAggregatorProvider) {
		t.Fatalf("Sum() error = %v, want ErrInvalidAggregatorProvider", err)
	}

	if _, err := a.Min(NewManager().Create(1, "SGD")); !errors.Is(err, exception.ErrInvalidAggregatorProvider) {
		t.Fatalf("Min() error = %v, want ErrInvalidAggregatorProvider", err)
	}

	if _, err := a.Max(NewManager().Create(1, "SGD")); !errors.Is(err, exception.ErrInvalidAggregatorProvider) {
		t.Fatalf("Max() error = %v, want ErrInvalidAggregatorProvider", err)
	}

	if _, err := a.Avg(NewManager().Create(1, "SGD")); !errors.Is(err, exception.ErrInvalidAggregatorProvider) {
		t.Fatalf("Avg() error = %v, want ErrInvalidAggregatorProvider", err)
	}

	a = &Aggregator{manager: nil}

	if _, err := a.Sum(NewManager().Create(1, "SGD")); !errors.Is(err, exception.ErrInvalidAggregatorProvider) {
		t.Fatalf("Sum() error = %v, want ErrInvalidAggregatorProvider", err)
	}
}

func TestAggregatorAvgRoundsHalfAwayFromZero(t *testing.T) {
	mm := NewManager()
	aggregator := NewAggregator(mm)

	cases := []struct {
		name    string
		amounts []int64
		want    int64
	}{
		{name: "positive half rounds up", amounts: []int64{100, 101}, want: 101},
		{name: "negative half rounds away", amounts: []int64{-100, -101}, want: -101},
		{name: "small positive half rounds up", amounts: []int64{2, 1}, want: 2},
		{name: "below half rounds toward zero", amounts: []int64{100, 200, 301}, want: 200},
		{name: "above half rounds away", amounts: []int64{100, 200, 302}, want: 201},
		{name: "negative below half rounds toward zero", amounts: []int64{-100, -200, -301}, want: -200},
		{name: "negative above half rounds away", amounts: []int64{-100, -200, -302}, want: -201},
		{name: "exact average unchanged", amounts: []int64{100, 200, 300}, want: 200},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			moneys := make([]*Value, 0, len(tc.amounts))

			for _, amount := range tc.amounts {
				moneys = append(moneys, mm.Create(amount, currency.SGD))
			}

			result, err := aggregator.Avg(moneys...)

			if err != nil {
				t.Fatalf("Avg() error = %v, want nil", err)
			}

			if result.amount != tc.want {
				t.Fatalf("Avg(%v) = %d, want %d", tc.amounts, result.amount, tc.want)
			}
		})
	}
}

func TestAggregatorSumOverflow(t *testing.T) {
	mm := NewManager()
	aggregator := NewAggregator(mm)

	result, err := aggregator.Sum(
		mm.Create(math.MaxInt64, currency.SGD),
		mm.Create(1, currency.SGD),
	)

	if !errors.Is(err, exception.ErrOverflow) {
		t.Fatalf("Sum() error = %v, want ErrOverflow", err)
	}

	if result != nil {
		t.Fatalf("Sum() result = %v, want nil", result)
	}
}

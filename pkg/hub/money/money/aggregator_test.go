package money

import (
	"errors"
	"math"
	"testing"

	"github.com/oullin/alloy/pkg/hub/money/currency"
	"github.com/oullin/alloy/pkg/hub/money/exception"
	"github.com/oullin/alloy/pkg/hub/money/exchange"
	testutil "github.com/oullin/alloy/pkg/hub/money/tests"
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

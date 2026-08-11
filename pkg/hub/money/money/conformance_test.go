package money

import (
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"

	"hara.sh/alloy/money/calculator"
	"hara.sh/alloy/money/currency"
	"hara.sh/alloy/money/exception"
	"hara.sh/alloy/money/exchange"
)

// moneyConformanceCase mirrors one entry in conformance/money.json. Numeric
// inputs and outputs are strings so JSON float precision never enters the
// comparison; error cases are matched by sentinel identity, not message text.
type moneyConformanceCase struct {
	Op       string   `json:"op"`
	Args     []string `json:"args"`
	Expected string   `json:"expected"`
	Error    string   `json:"error"`
	Note     string   `json:"note"`
}

type moneyConformanceFile struct {
	Cases []moneyConformanceCase `json:"cases"`
}

// moneyConformanceErrors maps the language-neutral error identity used in the
// fixtures to the Go sentinel it must resolve to.
var moneyConformanceErrors = map[string]error{
	"ERR_OVERFLOW":               exception.ErrOverflow,
	"ERR_INVALID_JSON_UNMARSHAL": exception.ErrInvalidJSONUnmarshal,
	"ERR_CURRENCY_NOT_FOUND":     exception.ErrCurrencyNotFound,
}

// TestMoneyConformance executes the shared Go<->TS money fixtures against the
// real Go API. It is the Go half of the cross-runtime drift guard (plan 008):
// a divergence in either runtime fails this suite. The TS half lives in
// sdk/money/tests/src/conformance.test.ts and reads the same JSON.
func TestMoneyConformance(t *testing.T) {
	cases := loadMoneyConformance(t)

	if len(cases) == 0 {
		t.Fatal("no money conformance cases loaded")
	}

	calc := calculator.NewCalculator()
	manager := NewManager()
	rates := exchange.NewExchange()

	for _, tc := range cases {
		tc := tc

		t.Run(tc.Op+"("+joinArgs(tc.Args)+")", func(t *testing.T) {
			got, err := runMoneyOp(t, calc, manager, rates, tc)

			if tc.Error != "" {
				want, ok := moneyConformanceErrors[tc.Error]

				if !ok {
					t.Fatalf("unknown error identity %q in fixture: %s", tc.Error, tc.Note)
				}

				if !errors.Is(err, want) {
					t.Fatalf("op %s%v: error = %v, want %v (%s)", tc.Op, tc.Args, err, want, tc.Note)
				}

				return
			}

			if err != nil {
				t.Fatalf("op %s%v: unexpected error %v (%s)", tc.Op, tc.Args, err, tc.Note)
			}

			if got != tc.Expected {
				t.Fatalf("op %s%v = %s, want %s (%s)", tc.Op, tc.Args, got, tc.Expected, tc.Note)
			}
		})
	}
}

func runMoneyOp(t *testing.T, calc *calculator.Engine, manager *Manager, rates *exchange.Rates, tc moneyConformanceCase) (string, error) {
	t.Helper()

	switch tc.Op {
	case "round":
		amount := mustInt64(t, tc.Args[0])
		exponent := mustInt(t, tc.Args[1])

		return strconv.FormatInt(calc.Round(amount, exponent), 10), nil
	case "absolute":
		return strconv.FormatInt(calc.Absolute(mustInt64(t, tc.Args[0])), 10), nil
	case "add":
		got, err := calc.SafeAdd(mustInt64(t, tc.Args[0]), mustInt64(t, tc.Args[1]))

		return strconv.FormatInt(got, 10), err
	case "subtract":
		got, err := calc.SafeSubtract(mustInt64(t, tc.Args[0]), mustInt64(t, tc.Args[1]))

		return strconv.FormatInt(got, 10), err
	case "multiply":
		got, err := calc.SafeMultiply(mustInt64(t, tc.Args[0]), mustInt64(t, tc.Args[1]))

		return strconv.FormatInt(got, 10), err
	case "createFromFloat":
		value := manager.CreateFromFloat(mustFloat(t, tc.Args[0]), tc.Args[1])
		amount, err := value.Amount()

		return strconv.FormatInt(amount, 10), err
	case "convertWithRate":
		got, err := rates.ConvertAmountWithRate(
			mustInt64(t, tc.Args[0]),
			mustInt(t, tc.Args[1]),
			mustInt(t, tc.Args[2]),
			mustFloat(t, tc.Args[3]),
		)

		return strconv.FormatInt(got, 10), err
	case "avg":
		values := make([]*Value, 0, len(tc.Args)-1)

		for _, raw := range tc.Args[1:] {
			values = append(values, manager.Create(mustInt64(t, raw), tc.Args[0]))
		}

		got, err := NewAggregator(manager).Avg(values...)

		if err != nil {
			return "", err
		}

		amount, err := got.Amount()

		return strconv.FormatInt(amount, 10), err
	case "unmarshalAmount":
		var value Value

		if err := json.Unmarshal([]byte(tc.Args[0]), &value); err != nil {
			return "", err
		}

		amount, err := value.Amount()

		return strconv.FormatInt(amount, 10), err
	case "unmarshalCurrency":
		var value Value

		if err := json.Unmarshal([]byte(tc.Args[0]), &value); err != nil {
			return "", err
		}

		curr, err := value.Currency()

		if err != nil {
			return "", err
		}

		return curr.Code, nil
	case "displayCompact":
		value := manager.Create(mustInt64(t, tc.Args[1]), tc.Args[0])

		return value.DisplayCompact()
	case "resolveWithDefault":
		// Per manager, so nothing leaks between cases.
		currencies := currency.NewManager()

		if err := currencies.SetDefault(tc.Args[0]); err != nil {
			return "", err
		}

		return currencies.Resolve(tc.Args[1]).Code, nil
	default:
		t.Fatalf("unknown money conformance op: %s", tc.Op)

		return "", nil
	}
}

func loadMoneyConformance(t *testing.T) []moneyConformanceCase {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)

	if !ok {
		t.Fatal("cannot resolve conformance test file path")
	}

	path := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "..", "conformance", "money.json")

	data, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var file moneyConformanceFile

	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	return file.Cases
}

func mustInt64(t *testing.T, value string) int64 {
	t.Helper()

	parsed, err := strconv.ParseInt(value, 10, 64)

	if err != nil {
		t.Fatalf("parse int64 %q: %v", value, err)
	}

	return parsed
}

func mustInt(t *testing.T, value string) int {
	t.Helper()

	parsed, err := strconv.Atoi(value)

	if err != nil {
		t.Fatalf("parse int %q: %v", value, err)
	}

	return parsed
}

func mustFloat(t *testing.T, value string) float64 {
	t.Helper()

	parsed, err := strconv.ParseFloat(value, 64)

	if err != nil {
		t.Fatalf("parse float %q: %v", value, err)
	}

	return parsed
}

func joinArgs(args []string) string {
	result := ""

	for index, arg := range args {
		if index > 0 {
			result += ","
		}

		result += arg
	}

	return result
}

package tempo

import (
	"encoding/json"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"testing"
)

// tempoConformanceComponents mirrors the calendar-component object used to
// build an instant in a named zone in conformance/tempo.json.
type tempoConformanceComponents struct {
	Year     int    `json:"year"`
	Month    int    `json:"month"`
	Day      int    `json:"day"`
	Hour     int    `json:"hour"`
	Minute   int    `json:"minute"`
	Second   int    `json:"second"`
	TimeZone string `json:"timeZone"`
}

// tempoConformanceCase mirrors one entry in conformance/tempo.json. Results are
// rendered as strings ('iso' -> UTC ISO-8601, 'date' -> YYYY-MM-DD, 'int' ->
// decimal) so the comparison is language-neutral.
type tempoConformanceCase struct {
	Op       string                      `json:"op"`
	Base     *tempoConformanceComponents `json:"base"`
	Other    *tempoConformanceComponents `json:"other"`
	Arg      int                         `json:"arg"`
	Input    string                      `json:"input"`
	Pattern  string                      `json:"pattern"`
	TimeZone string                      `json:"timeZone"`
	Render   string                      `json:"render"`
	Expected string                      `json:"expected"`
	Note     string                      `json:"note"`
}

type tempoConformanceFile struct {
	Cases []tempoConformanceCase `json:"cases"`
}

// TestTempoConformance executes the shared Go<->TS tempo fixtures against the
// real Go API. It is the Go half of the cross-runtime drift guard (plan 008);
// the TS half lives in sdk/tempo/tests/src/conformance.test.ts and reads the
// same JSON.
func TestTempoConformance(t *testing.T) {
	cases := loadTempoConformance(t)

	if len(cases) == 0 {
		t.Fatal("no tempo conformance cases loaded")
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.Op+"/"+tc.Note, func(t *testing.T) {
			got := runTempoOp(t, tc)

			if got != tc.Expected {
				t.Fatalf("op %s = %s, want %s (%s)", tc.Op, got, tc.Expected, tc.Note)
			}
		})
	}
}

func runTempoOp(t *testing.T, tc tempoConformanceCase) string {
	t.Helper()

	switch tc.Op {
	case "addDays":
		return renderTempo(t, buildTempo(t, tc.Base).AddDays(tc.Arg), tc.Render)
	case "addWeeks":
		return renderTempo(t, buildTempo(t, tc.Base).AddWeeks(tc.Arg), tc.Render)
	case "addHours":
		return renderTempo(t, buildTempo(t, tc.Base).AddHours(tc.Arg), tc.Render)
	case "addMonths":
		return renderTempo(t, buildTempo(t, tc.Base).AddMonths(tc.Arg), tc.Render)
	case "addMonthsNoOverflow":
		return renderTempo(t, buildTempo(t, tc.Base).AddMonthsNoOverflow(tc.Arg), tc.Render)
	case "diffInMonths":
		return strconv.Itoa(buildTempo(t, tc.Base).DiffInMonths(buildTempo(t, tc.Other)))
	case "diffInYears":
		return strconv.Itoa(buildTempo(t, tc.Base).DiffInYears(buildTempo(t, tc.Other)))
	case "parseFromPattern":
		factory, err := NewFactory(WithTimezone(tc.TimeZone))

		if err != nil {
			t.Fatalf("factory %q: %v", tc.TimeZone, err)
		}

		parsed, err := factory.FromFormat(tc.Input, tc.Pattern)

		if err != nil {
			t.Fatalf("parseFromPattern(%q, %q): %v", tc.Input, tc.Pattern, err)
		}

		return renderTempo(t, parsed, tc.Render)
	default:
		t.Fatalf("unknown tempo conformance op: %s", tc.Op)

		return ""
	}
}

func buildTempo(t *testing.T, components *tempoConformanceComponents) Time {
	t.Helper()

	if components == nil {
		t.Fatal("tempo conformance case missing components")
	}

	factory, err := NewFactory(WithTimezone(components.TimeZone))

	if err != nil {
		t.Fatalf("factory %q: %v", components.TimeZone, err)
	}

	value, err := factory.Create(Components{
		Year:   components.Year,
		Month:  components.Month,
		Day:    components.Day,
		Hour:   components.Hour,
		Minute: components.Minute,
		Second: components.Second,
	})

	if err != nil {
		t.Fatalf("create %+v: %v", components, err)
	}

	return value
}

func renderTempo(t *testing.T, value Time, render string) string {
	t.Helper()

	switch render {
	case "iso":
		return value.ISOString()
	case "date":
		return value.DateString()
	default:
		t.Fatalf("unknown tempo render mode: %s", render)

		return ""
	}
}

func loadTempoConformance(t *testing.T) []tempoConformanceCase {
	t.Helper()

	_, thisFile, _, ok := runtime.Caller(0)

	if !ok {
		t.Fatal("cannot resolve conformance test file path")
	}

	path := filepath.Join(filepath.Dir(thisFile), "..", "..", "..", "conformance", "tempo.json")

	data, err := os.ReadFile(path)

	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}

	var file tempoConformanceFile

	if err := json.Unmarshal(data, &file); err != nil {
		t.Fatalf("parse %s: %v", path, err)
	}

	return file.Cases
}

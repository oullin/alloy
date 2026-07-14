package tempo

import (
	"time"

	configpkg "github.com/oullin/alloy/pkg/hub/tempo/config"
	"github.com/oullin/alloy/pkg/hub/tempo/duration"
	"github.com/oullin/alloy/pkg/hub/tempo/factory"
	"github.com/oullin/alloy/pkg/hub/tempo/runtime"
)

type Unit = duration.Unit

type TimeStringPrecision string

type Components = factory.Components

type Object struct {
	Year          int
	Month         int
	Day           int
	Hour          int
	Minute        int
	Second        int
	Millisecond   int
	Timezone      string
	OffsetMinutes int
	Weekday       int
}

type Duration = duration.Span

type Span = duration.Span

type DiffOptions struct {
	Absolute bool
	Float    bool
}

// HumanDiffOptions aliases the equivalent type in the config package so
// env-loaded defaults and runtime overrides share a single shape. Settings
// cannot be aliased the same way — it carries TestNow *Time, which would
// create an import cycle if pushed down into config/.
type HumanDiffOptions = configpkg.HumanDiffOptions

type Settings struct {
	FallbackLocale string
	HumanDiff      HumanDiffOptions
	Locale         string
	MidDayAt       int
	MonthsOverflow bool
	StrictMode     bool
	TestNow        *Time
	Timezone       string
	WeekendDays    []time.Weekday
	YearsOverflow  bool
}

type Serializer func(Time) string

type ConfigOption func(*Config) error

type Translator = runtime.Translator

type StartOfWeekOptions struct {
	WeekStartsOn time.Weekday
}

type PeriodOptions struct {
	Step       Duration
	IncludeEnd bool
	ExcludeEnd bool
}

type Option func(*config) error

type config struct {
	location       *time.Location
	runtime        Context
	settings       Settings
	serializer     Serializer
	toStringFormat string
}

type Time struct {
	value          time.Time
	location       *time.Location
	runtime        Context
	settings       Settings
	serializer     Serializer
	toStringFormat string
}

type MutableTime struct {
	value          time.Time
	location       *time.Location
	runtime        Context
	settings       Settings
	serializer     Serializer
	toStringFormat string
}

type Interval struct {
	Start Time
	End   Time
}

type Period struct {
	Start      Time
	End        Time
	Step       Duration
	IncludeEnd bool
}

type Factory struct {
	clock          factory.Clock
	location       *time.Location
	runtime        Context
	settings       Settings
	serializer     Serializer
	toStringFormat string
}

const (
	Millisecond = duration.Millisecond
	Second      = duration.Second
	Minute      = duration.Minute
	Hour        = duration.Hour
	Day         = duration.Day
	Week        = duration.Week
	Month       = duration.Month
	Quarter     = duration.Quarter
	Year        = duration.Year
	Decade      = duration.Decade
	Century     = duration.Century
	Millennium  = duration.Millennium
)

const (
	SecondPrecision      TimeStringPrecision = "second"
	MillisecondPrecision TimeStringPrecision = "millisecond"
)

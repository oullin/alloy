package tempo

import (
	"time"

	"github.com/oullin/alloy/tempo/duration"
	"github.com/oullin/alloy/tempo/factory"
	"github.com/oullin/alloy/tempo/runtime"
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

type Duration = duration.Duration

type DiffOptions struct {
	Absolute bool
	Float    bool
}

type HumanDiffOptions struct {
	Absolute bool
	Locale   string
	Numeric  string
	Style    string
	Unit     Unit
}

type Settings struct {
	FallbackLocale string
	HumanDiff      HumanDiffOptions
	Locale         string
	MidDayAt       int
	MonthsOverflow bool
	StrictMode     bool
	TestNow        *Tempo
	Timezone       string
	WeekendDays    []time.Weekday
	YearsOverflow  bool
}

type Serializer func(Tempo) string

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
	location *time.Location
	runtime  Runtime
	app      *Config
}

type Tempo struct {
	value    time.Time
	location *time.Location
	runtime  Runtime
}

type MutableTempo struct {
	value    time.Time
	location *time.Location
	runtime  Runtime
}

type Interval struct {
	Start Tempo
	End   Tempo
}

type Period struct {
	Start      Tempo
	End        Tempo
	Step       Duration
	IncludeEnd bool
}

type Factory struct {
	clock    factory.Clock
	location *time.Location
	runtime  Runtime
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

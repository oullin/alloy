package tempo

import "time"

type Unit string

const (
	Millisecond Unit = "millisecond"
	Second      Unit = "second"
	Minute      Unit = "minute"
	Hour        Unit = "hour"
	Day         Unit = "day"
	Week        Unit = "week"
	Month       Unit = "month"
	Quarter     Unit = "quarter"
	Year        Unit = "year"
	Decade      Unit = "decade"
	Century     Unit = "century"
	Millennium  Unit = "millennium"
)

type TimeStringPrecision string

const (
	SecondPrecision      TimeStringPrecision = "second"
	MillisecondPrecision TimeStringPrecision = "millisecond"
)

type Components struct {
	Year        int
	Month       int
	Day         int
	Hour        int
	Minute      int
	Second      int
	Millisecond int
	Timezone    string
}

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

type Duration struct {
	Years        int
	Quarters     int
	Months       int
	Weeks        int
	Days         int
	Hours        int
	Minutes      int
	Seconds      int
	Milliseconds int
}

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

type Translator interface {
	Message(key string) (any, bool)
	Translate(key string, replacements map[string]string) (string, bool)
}

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
	now      *time.Time
	location *time.Location
	runtime  Runtime
}

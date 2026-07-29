package factory_test

import (
	"testing"
	"time"

	"hara.sh/alloy/tempo/duration"
	"hara.sh/alloy/tempo/factory"
)

func TestFromTimestampAndMs(t *testing.T) {
	if got := factory.FromTimestamp(0).UTC(); !got.Equal(time.Unix(0, 0).UTC()) {
		t.Fatalf("FromTimestamp(0) = %v, want epoch", got)
	}

	if got := factory.FromTimestampMs(1_700_000_000_000); got.UnixMilli() != 1_700_000_000_000 {
		t.Fatalf("FromTimestampMs round-trip lost precision: got %d", got.UnixMilli())
	}
}

func TestTimeFromComponents(t *testing.T) {
	components := factory.Components{Year: 2024, Month: 5, Day: 15, Hour: 9, Minute: 30}
	got := factory.TimeFromComponents(components, time.UTC)

	want := time.Date(2024, time.May, 15, 9, 30, 0, 0, time.UTC)

	if !got.Equal(want) {
		t.Fatalf("TimeFromComponents = %v, want %v", got, want)
	}

	zeroFill := factory.TimeFromComponents(factory.Components{Year: 2024}, time.UTC)

	if zeroFill.Month() != time.January || zeroFill.Day() != 1 {
		t.Fatalf("TimeFromComponents zero-fills month=1 day=1, got %v", zeroFill)
	}
}

func TestComponentsMatchTime(t *testing.T) {
	value := time.Date(2024, time.May, 15, 9, 30, 0, 0, time.UTC)
	match := factory.Components{Year: 2024, Month: 5, Day: 15, Hour: 9, Minute: 30}

	if !factory.ComponentsMatchTime(match, value, time.UTC) {
		t.Fatalf("ComponentsMatchTime: identical components flagged as different")
	}

	off := factory.Components{Year: 2024, Month: 5, Day: 15, Hour: 9, Minute: 31}

	if factory.ComponentsMatchTime(off, value, time.UTC) {
		t.Fatalf("ComponentsMatchTime: differing minute should not match")
	}
}

func TestClockNow(t *testing.T) {
	now := time.Date(2024, time.January, 1, 0, 0, 0, 0, time.UTC)
	clock := factory.NewClock(&now)

	if got := clock.Now(); !got.Equal(now) {
		t.Fatalf("Clock with fixed now returned %v, want %v", got, now)
	}

	live := factory.NewClock(nil)
	before := time.Now()
	got := live.Now()
	after := time.Now()

	if got.Before(before) || got.After(after) {
		t.Fatalf("live Clock.Now() outside [%v, %v]: %v", before, after, got)
	}
}

func TestMetadataAccessors(t *testing.T) {
	if cal := factory.CalendarFormats(); len(cal) == 0 {
		t.Fatalf("CalendarFormats() returned empty map")
	}

	if iso := factory.ISOFormats(); len(iso) == 0 {
		t.Fatalf("ISOFormats() returned empty map")
	}

	if units := factory.ISOUnits(); len(units) == 0 {
		t.Fatalf("ISOUnits() returned empty slice")
	}

	if got := factory.WeekStartsAt(); got != time.Monday {
		t.Fatalf("WeekStartsAt() = %v, want Monday", got)
	}

	if got := factory.WeekEndsAt(); got != time.Sunday {
		t.Fatalf("WeekEndsAt() = %v, want Sunday", got)
	}

	if !factory.IsModifiableUnit(duration.Day) {
		t.Fatalf("Day should be modifiable")
	}

	if got := factory.SingularUnit(duration.Unit("days")); got != duration.Day {
		t.Fatalf("SingularUnit(days) = %v, want Day", got)
	}

	if got := factory.PluralUnit(duration.Day); got == "" {
		t.Fatalf("PluralUnit(Day) returned empty string")
	}
}

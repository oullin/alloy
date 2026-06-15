package formatting_test

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/oullin/alloy/tempo/tempo"
)

func assertEqual(t *testing.T, label string, got string, want string) {
	t.Helper()

	if got != want {
		t.Fatalf("%s = %q, want %q", label, got, want)
	}
}

func TestISOStringFormatsUTC(t *testing.T) {
	parsed, err := tempo.Parse("2024-01-01T01:00:00+01:00")

	if err != nil {
		t.Fatalf("parse offset timestamp: %v", err)
	}

	if got := parsed.ISOString(); got != "2024-01-01T00:00:00.000Z" {
		t.Fatalf("ISOString() = %q, want UTC ISO string", got)
	}
}

func TestFromFormatPredicatesAndHumanDiffs(t *testing.T) {
	withOffset, err := tempo.FromFormat("2024/05/15 10:34:45.600 +0900", "YYYY/MM/DD HH:mm:ss.SSS ZZ")

	if err != nil {
		t.Fatalf("from format offset: %v", err)
	}

	if got := withOffset.ISOString(); got != "2024-05-15T01:34:45.600Z" {
		t.Fatalf("FromFormat offset ISOString() = %q, want parsed offset instant", got)
	}

	meridiem, err := tempo.FromFormat("05-15-24 10:34 PM", "MM-DD-YY hh:mm A")

	if err != nil {
		t.Fatalf("from format meridiem: %v", err)
	}

	if got := meridiem.ISOString(); got != "2024-05-15T22:34:00.000Z" {
		t.Fatalf("FromFormat meridiem ISOString() = %q, want parsed time", got)
	}

	named, err := tempo.FromFormat("Wednesday, May 15th 2024 10:34 PM", "dddd, MMMM Do YYYY hh:mm A")

	if err != nil {
		t.Fatalf("from format named month: %v", err)
	}

	if got := named.ISOString(); got != "2024-05-15T22:34:00.000Z" {
		t.Fatalf("FromFormat named month ISOString() = %q, want parsed time", got)
	}

	shortNamed, err := tempo.FromFormat("Wed, May 15 2024", "ddd, MMM D YYYY")

	if err != nil {
		t.Fatalf("from format short named month: %v", err)
	}

	if got := shortNamed.DateString(); got != "2024-05-15" {
		t.Fatalf("FromFormat short named DateString() = %q, want parsed date", got)
	}

	tokyo, err := tempo.FromFormat("2024-01-01 09:00", "YYYY-MM-DD HH:mm", tempo.WithTimezone("Asia/Tokyo"))

	if err != nil {
		t.Fatalf("from format tokyo: %v", err)
	}

	if got := tokyo.ISOString(); got != "2024-01-01T00:00:00.000Z" {
		t.Fatalf("FromFormat timezone ISOString() = %q, want local timezone instant", got)
	}

	if !tempo.CanParse("2024-05-15T10:34:45Z") {
		t.Fatalf("CanParse(valid) = false, want true")
	}

	if tempo.CanParse("not a date") {
		t.Fatalf("CanParse(invalid) = true, want false")
	}

	if parsed, ok := tempo.TryParse("2024-05-15"); !ok || parsed.DateString() != "2024-05-15" {
		t.Fatalf("TryParse(valid) = %q, %v, want date, true", parsed.DateString(), ok)
	}

	if _, ok := tempo.TryParse("not a date"); ok {
		t.Fatalf("TryParse(invalid) ok = true, want false")
	}

	if !tempo.HasFormat("2024/05/15 10:34", "YYYY/MM/DD HH:mm") {
		t.Fatalf("HasFormat(valid) = false, want true")
	}

	if tempo.HasFormat("2024-05-15", "YYYY/MM/DD") {
		t.Fatalf("HasFormat(invalid) = true, want false")
	}

	if parsed, ok := tempo.TryFromFormat("2024/05/15 10:34", "YYYY/MM/DD HH:mm"); !ok || parsed.ISOString() != "2024-05-15T10:34:00.000Z" {
		t.Fatalf("TryFromFormat(valid) = %q, %v, want parsed instant, true", parsed.ISOString(), ok)
	}

	if _, ok := tempo.TryFromFormat("2024-05-15", "YYYY/MM/DD"); ok {
		t.Fatalf("TryFromFormat(invalid) ok = true, want false")
	}

	base, err := tempo.Parse("2024-02-29T00:00:00Z")

	if err != nil {
		t.Fatalf("parse base: %v", err)
	}

	saturday, err := tempo.Parse("2024-03-02T00:00:00Z")

	if err != nil {
		t.Fatalf("parse saturday: %v", err)
	}

	referenceYesterday, err := tempo.Parse("2024-02-28T12:00:00Z")

	if err != nil {
		t.Fatalf("parse yesterday reference: %v", err)
	}

	referenceTomorrow, err := tempo.Parse("2024-03-01T12:00:00Z")

	if err != nil {
		t.Fatalf("parse tomorrow reference: %v", err)
	}

	if !base.IsLeapYear() {
		t.Fatalf("IsLeapYear() = false, want true")
	}

	if got := base.DaysInYear(); got != 366 {
		t.Fatalf("DaysInYear() = %d, want 366", got)
	}

	if base.IsLongYear() {
		t.Fatalf("IsLongYear() = true, want false")
	}

	longYear, err := tempo.Parse("2020-12-31T00:00:00Z")

	if err != nil {
		t.Fatalf("parse long year: %v", err)
	}

	if !longYear.IsLongYear() {
		t.Fatalf("IsLongYear() = false, want true")
	}

	if got := base.DaysInMonth(); got != 29 {
		t.Fatalf("DaysInMonth() = %d, want 29", got)
	}

	if !base.IsLastOfMonth() {
		t.Fatalf("IsLastOfMonth() = false, want true")
	}

	if base.SubDays(1).IsLastOfMonth() {
		t.Fatalf("IsLastOfMonth(previous day) = true, want false")
	}

	if !saturday.IsWeekend() {
		t.Fatalf("IsWeekend() = false, want true")
	}

	if !base.IsWeekday() {
		t.Fatalf("IsWeekday() = false, want true")
	}

	if !base.IsDayOfWeek(time.Thursday) {
		t.Fatalf("IsDayOfWeek(Thursday) = false, want true")
	}

	if base.IsDayOfWeek(time.Friday) {
		t.Fatalf("IsDayOfWeek(Friday) = true, want false")
	}

	if !base.IsTomorrow(referenceYesterday) {
		t.Fatalf("IsTomorrow() = false, want true")
	}

	if !base.IsYesterday(referenceTomorrow) {
		t.Fatalf("IsYesterday() = false, want true")
	}

	if !base.IsPast(referenceTomorrow) {
		t.Fatalf("IsPast() = false, want true")
	}

	if !base.IsFuture(referenceYesterday) {
		t.Fatalf("IsFuture() = false, want true")
	}

	farFuture, err := tempo.Parse("3000-01-01T00:00:00Z")

	if err != nil {
		t.Fatalf("parse far future: %v", err)
	}

	if !base.IsNowOrPast() || !farFuture.IsNowOrFuture() {
		t.Fatalf("now-or-past/future predicates did not match expected values")
	}

	if got := base.AddDays(2).DiffForHumans(base); got != "in 2 days" {
		t.Fatalf("DiffForHumans() = %q, want future diff", got)
	}

	if got := base.DiffForHumans(base.AddHours(3)); got != "3 hours ago" {
		t.Fatalf("DiffForHumans() = %q, want past diff", got)
	}
}

func TestNamedSerializationAndMapConversion(t *testing.T) {
	instance, err := tempo.Parse("2024-05-15T12:34:56.789Z", tempo.WithTimezone("Asia/Tokyo"))

	if err != nil {
		t.Fatalf("parse fixture input: %v", err)
	}

	assertEqual(t, "DateTimeLocalString()", instance.DateTimeLocalString(), "2024-05-15T21:34:56")
	assertEqual(t, "DateTimeLocalString(ms)", instance.DateTimeLocalString(tempo.MillisecondPrecision), "2024-05-15T21:34:56.789")
	assertEqual(t, "ISOFormat()", instance.ISOFormat("YYYY-MM-DD HH:mm:ss"), "2024-05-15 21:34:56")
	assertEqual(t, "TranslatedFormat()", instance.TranslatedFormat("YYYY-MM-DD HH:mm:ss"), "2024-05-15 21:34:56")
	assertEqual(t, "FormattedDateString()", instance.FormattedDateString(), "May 15, 2024")
	assertEqual(t, "FormattedDayDateString()", instance.FormattedDayDateString(), "Wed, May 15, 2024")
	assertEqual(t, "DayDateTimeString()", instance.DayDateTimeString(), "Wed, May 15, 2024 9:34 PM")
	assertEqual(t, "TimeString(ms)", instance.TimeString(tempo.MillisecondPrecision), "21:34:56.789")
	assertEqual(t, "ISO8601String()", instance.ISO8601String(), "2024-05-15T21:34:56+09:00")
	assertEqual(t, "ISO8601ZuluString()", instance.ISO8601ZuluString(), "2024-05-15T12:34:56Z")
	assertEqual(t, "ISO8601ZuluString(ms)", instance.ISO8601ZuluString(tempo.MillisecondPrecision), "2024-05-15T12:34:56.789Z")
	assertEqual(t, "RFC3339String(ms)", instance.RFC3339String(tempo.MillisecondPrecision), "2024-05-15T21:34:56.789+09:00")
	assertEqual(t, "RFC822String()", instance.RFC822String(), "Wed, 15 May 24 21:34:56 +0900")
	assertEqual(t, "RFC850String()", instance.RFC850String(), "Wednesday, 15-May-24 21:34:56 +0900")
	assertEqual(t, "RFC1036String()", instance.RFC1036String(), "Wed, 15 May 24 21:34:56 +0900")
	assertEqual(t, "RFC1123String()", instance.RFC1123String(), "Wed, 15 May 2024 21:34:56 +0900")
	assertEqual(t, "RFC2822String()", instance.RFC2822String(), "Wed, 15 May 2024 21:34:56 +0900")
	assertEqual(t, "W3CString()", instance.W3CString(), "2024-05-15T21:34:56+09:00")
	assertEqual(t, "AtomString()", instance.AtomString(), "2024-05-15T21:34:56+09:00")
	assertEqual(t, "RSSString()", instance.RSSString(), "Wed, 15 May 2024 21:34:56 +0900")
	assertEqual(t, "RFC7231String()", instance.RFC7231String(), "Wed, 15 May 2024 12:34:56 GMT")
	assertEqual(t, "CookieString()", instance.CookieString(), "Wed, 15-May-2024 12:34:56 GMT")
	assertEqual(t, "UnixString()", instance.UnixString(), "1715776496")

	if got := instance.Unix(); got != 1715776496 {
		t.Fatalf("Unix() = %d, want 1715776496", got)
	}

	if got := instance.GetTimestampMs(); got != 1715776496789 {
		t.Fatalf("GetTimestampMs() = %d, want 1715776496789", got)
	}

	assertEqual(t, "JSONSerialize()", instance.JSONSerialize(), "2024-05-15T12:34:56.789Z")
	assertEqual(t, "Serialize()", instance.Serialize(), "2024-05-15T12:34:56.789Z")
	tempoJSON, err := json.Marshal(instance)

	if err != nil {
		t.Fatalf("marshal tempo: %v", err)
	}

	assertEqual(t, "json.Marshal(Tempo)", string(tempoJSON), `"2024-05-15T12:34:56.789Z"`)
	mutableJSON, err := json.Marshal(tempo.NewMutable(instance))

	if err != nil {
		t.Fatalf("marshal mutable tempo: %v", err)
	}

	assertEqual(t, "json.Marshal(MutableTempo)", string(mutableJSON), `"2024-05-15T12:34:56.789Z"`)
	durationValue := tempo.Duration{Days: 1, Hours: 2}
	durationJSON, err := json.Marshal(&durationValue)

	if err != nil {
		t.Fatalf("marshal duration: %v", err)
	}

	assertEqual(t, "json.Marshal(Duration)", string(durationJSON), `"P1DT2H"`)

	fromSerialized, err := tempo.FromSerialized(`"2024-05-15T12:34:56.789Z"`)

	if err != nil {
		t.Fatalf("FromSerialized(): %v", err)
	}

	assertEqual(t, "FromSerialized().ISOString()", fromSerialized.ISOString(), "2024-05-15T12:34:56.789Z")

	var decodedMutable tempo.MutableTempo

	if err := json.Unmarshal(tempoJSON, &decodedMutable); err != nil {
		t.Fatalf("unmarshal mutable tempo: %v", err)
	}

	assertEqual(t, "json.Unmarshal(MutableTempo)", decodedMutable.ISOString(), "2024-05-15T12:34:56.789Z")

	existingMutable := tempo.NewMutable(instance).WithRuntime(tempo.NewRuntime(tempo.RuntimeLocale("es-ES")))

	if err := json.Unmarshal(tempoJSON, existingMutable); err != nil {
		t.Fatalf("unmarshal existing mutable tempo: %v", err)
	}

	assertEqual(t, "json.Unmarshal(existing MutableTempo)", existingMutable.ISOString(), "2024-05-15T12:34:56.789Z")

	if got := existingMutable.Runtime().Locale(); got != "es-ES" {
		t.Fatalf("json.Unmarshal(existing MutableTempo) runtime locale = %q, want es-ES", got)
	}

	var decodedDuration tempo.Duration

	if err := json.Unmarshal(durationJSON, &decodedDuration); err != nil {
		t.Fatalf("unmarshal duration: %v", err)
	}

	assertEqual(t, "json.Unmarshal(Duration)", decodedDuration.ISOString(), "P1DT2H")

	values := instance.ToMap()

	if values["timeZone"] != "Asia/Tokyo" {
		t.Fatalf("ToMap()[timeZone] = %v, want Asia/Tokyo", values["timeZone"])
	}

	if values["hour"] != 21 {
		t.Fatalf("ToMap()[hour] = %v, want 21", values["hour"])
	}

	if value, ok := instance.Get("year"); !ok || value != 2024 {
		t.Fatalf("Get(year) = %v, %v, want 2024, true", value, ok)
	}

	if value, ok := instance.GetPaddedUnit("month", 2); !ok || value != "05" {
		t.Fatalf("GetPaddedUnit(month) = %q, %v, want 05, true", value, ok)
	}

	assertEqual(t, "GetTranslatedDayName()", instance.GetTranslatedDayName(), "Wednesday")
	assertEqual(t, "GetTranslatedShortMonthName()", instance.GetTranslatedShortMonthName(), "May")
	assertEqual(t, "TranslateNumber()", instance.TranslateNumber(1234), "1234")
	assertEqual(t, "Translate()", instance.Translate("Hello :name", map[string]string{"name": "Tempo"}), "Hello Tempo")

	if _, ok := tempo.TryParse("not a date"); ok || tempo.GetLastErrors() == nil {
		t.Fatalf("TryParse(invalid), GetLastErrors() = %v, %v, want false and error", ok, tempo.GetLastErrors())
	}

	if got := tempo.ExecuteWithLocale("fr-FR", tempo.GetLocale); got != "fr-FR" {
		t.Fatalf("ExecuteWithLocale() = %q, want fr-FR", got)
	}

	if got := tempo.GetLocale(); got != "en-US" {
		t.Fatalf("GetLocale() after ExecuteWithLocale = %q, want restored default", got)
	}

	tempo.SetToStringFormat("YYYY-MM-DD")
	assertEqual(t, "String() formatted", instance.String(), "2024-05-15")
	tempo.ResetToStringFormat()
	tempo.SerializeUsing(func(value tempo.Tempo) string { return value.DateString() })
	hookedJSON, err := json.Marshal(instance)

	if err != nil {
		t.Fatalf("marshal hooked tempo: %v", err)
	}

	assertEqual(t, "json.Marshal(hooked Tempo)", string(hookedJSON), `"2024-05-15"`)
	tempo.SerializeUsing(nil)
	dateOnly := func(value tempo.Tempo) string { return value.DateString() }
	assertEqual(t, "composable dateOnly", dateOnly(instance), "2024-05-15")
}

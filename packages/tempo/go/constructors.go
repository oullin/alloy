package tempo

import (
	"errors"
	"strconv"
	"strings"
	"time"
)

func Now(options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)
	if err != nil {
		return Tempo{}, err
	}

	if tempoSettings.TestNow != nil {
		location := cfg.location
		if len(options) == 0 && tempoSettings.Timezone != "" {
			configured, err := loadLocation(tempoSettings.Timezone)
			if err != nil {
				return Tempo{}, err
			}
			location = configured
		}

		return newTempo(tempoSettings.TestNow.value, location, cfg.runtime), nil
	}

	return newTempo(time.Now(), cfg.location, cfg.runtime), nil
}

func Today(options ...Option) (Tempo, error) {
	now, err := Now(options...)
	if err != nil {
		return Tempo{}, err
	}

	return now.StartOfDay(), nil
}

func Tomorrow(options ...Option) (Tempo, error) {
	today, err := Today(options...)
	if err != nil {
		return Tempo{}, err
	}

	return today.AddDays(1), nil
}

func Yesterday(options ...Option) (Tempo, error) {
	today, err := Today(options...)
	if err != nil {
		return Tempo{}, err
	}

	return today.SubDays(1), nil
}

func FromTime(value time.Time, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)
	if err != nil {
		return Tempo{}, err
	}

	return newTempo(value, cfg.location, cfg.runtime), nil
}

func Parse(input string, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)
	if err != nil {
		return Tempo{}, err
	}

	return newParser(cfg.location, cfg.runtime).Parse(input)
}

func RawParse(input string, options ...Option) (Tempo, error) {
	return Parse(input, options...)
}

func TryParse(input string, options ...Option) (Tempo, bool) {
	tempo, err := Parse(input, options...)
	lastTempoError = err
	return tempo, err == nil
}

func CanParse(input string, options ...Option) bool {
	_, ok := TryParse(input, options...)
	return ok
}

func FromSerialized(input string, options ...Option) (Tempo, error) {
	value, err := strconv.Unquote(input)
	if err != nil {
		return Tempo{}, err
	}

	return Parse(value, options...)
}

func GetLastErrors() error { return lastTempoError }

func ExecuteWithLocale[T any](locale string, callback func() T) T {
	previous := tempoSettings.Locale
	tempoSettings.Locale = locale
	defer func() { tempoSettings.Locale = previous }()

	return callback()
}

func SerializeUsing(next Serializer) { serializer = next }

func SetToStringFormat(pattern string) { toStringFormat = pattern }

func ResetToStringFormat() { toStringFormat = "" }

func GetClock() *Tempo { return GetTestNow() }

func Make(input string, options ...Option) (Tempo, error) {
	return Parse(input, options...)
}

func ParseFromLocale(input string, _ string, options ...Option) (Tempo, error) {
	return Parse(input, options...)
}

func GetDays() []string {
	return append([]string(nil), dayNames[:]...)
}

func GetCalendarFormats() map[string]string {
	return map[string]string{
		"lastDay":  "[Yesterday at] HH:mm",
		"lastWeek": "[Last] dddd [at] HH:mm",
		"nextDay":  "[Tomorrow at] HH:mm",
		"nextWeek": "dddd [at] HH:mm",
		"sameDay":  "[Today at] HH:mm",
		"sameElse": "YYYY-MM-DD",
	}
}

func GetIsoFormats() map[string]string {
	return map[string]string{
		"atom":     "YYYY-MM-DDTHH:mm:ssZZ",
		"cookie":   "ddd, DD-MMM-YYYY HH:mm:ss [GMT]",
		"date":     "YYYY-MM-DD",
		"dateTime": "YYYY-MM-DD HH:mm:ss",
		"iso8601":  "YYYY-MM-DDTHH:mm:ssZZ",
		"time":     "HH:mm:ss",
	}
}

func GetIsoUnits() []Unit {
	return []Unit{Millisecond, Second, Minute, Hour, Day, Week, Month, Quarter, Year}
}

func GetTimeFormatByPrecision(precision TimeStringPrecision) string {
	if precision == MillisecondPrecision {
		return "HH:mm:ss.SSS"
	}

	return "HH:mm:ss"
}

func GetWeekStartsAt() time.Weekday { return time.Monday }

func GetWeekEndsAt() time.Weekday { return time.Sunday }

func LocaleHasDiffSyntax(locale string) bool { return strings.TrimSpace(locale) != "" }

func LocaleHasDiffOneDayWords(locale string) bool { return LocaleHasDiffSyntax(locale) }

func LocaleHasDiffTwoDayWords(locale string) bool { return LocaleHasDiffSyntax(locale) }

func LocaleHasPeriodSyntax(locale string) bool { return LocaleHasDiffSyntax(locale) }

func LocaleHasShortUnits(locale string) bool { return LocaleHasDiffSyntax(locale) }

func Sleep(seconds float64) {
	if seconds <= 0 {
		return
	}

	time.Sleep(time.Duration(seconds * float64(time.Second)))
}

func HasRelativeKeywords(input string) bool {
	value := strings.TrimSpace(strings.ToLower(input))
	return strings.HasPrefix(value, "+") ||
		strings.HasPrefix(value, "-") ||
		strings.Contains(value, "now") ||
		strings.Contains(value, "today") ||
		strings.Contains(value, "tomorrow") ||
		strings.Contains(value, "yesterday") ||
		strings.Contains(value, "next") ||
		strings.Contains(value, "last") ||
		strings.Contains(value, "ago")
}

func IsModifiableUnit(unit Unit) bool {
	switch normalizeUnit(unit) {
	case Millisecond, Second, Minute, Hour, Day, Week, Month, Quarter, Year, Decade, Century, Millennium:
		return true
	default:
		return false
	}
}

func SingularUnit(unit Unit) Unit {
	return normalizeUnit(unit)
}

func PluralUnit(unit Unit) string {
	return string(normalizeUnit(unit)) + "s"
}

func FromFormat(input string, pattern string, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)
	if err != nil {
		return Tempo{}, err
	}

	return newParser(cfg.location, cfg.runtime).FromFormat(input, pattern)
}

func CreateFromFormat(input string, pattern string, options ...Option) (Tempo, error) {
	return FromFormat(input, pattern, options...)
}

func CreateFromIsoFormat(input string, pattern string, options ...Option) (Tempo, error) {
	return FromFormat(input, pattern, options...)
}

func CreateFromLocaleFormat(input string, pattern string, _ string, options ...Option) (Tempo, error) {
	return FromFormat(input, pattern, options...)
}

func CreateFromLocaleIsoFormat(input string, pattern string, _ string, options ...Option) (Tempo, error) {
	return FromFormat(input, pattern, options...)
}

func RawCreateFromFormat(input string, pattern string, options ...Option) (Tempo, error) {
	return FromFormat(input, pattern, options...)
}

func TryFromFormat(input string, pattern string, options ...Option) (Tempo, bool) {
	tempo, err := FromFormat(input, pattern, options...)
	return tempo, err == nil
}

func HasFormat(input string, pattern string, options ...Option) bool {
	_, ok := TryFromFormat(input, pattern, options...)
	return ok
}

func CanBeCreatedFromFormat(input string, pattern string, options ...Option) bool {
	return HasFormat(input, pattern, options...)
}

func HasFormatWithModifiers(input string, pattern string, options ...Option) bool {
	return HasFormat(input, pattern, options...)
}

func Create(components Components) (Tempo, error) {
	location, err := loadLocation(components.Timezone)
	if err != nil {
		return Tempo{}, err
	}

	return newTempo(
		timeFromComponents(components, location),
		location,
		NewRuntime(
			RuntimeLocale(tempoSettings.Locale),
			RuntimeFallbackLocale(tempoSettings.FallbackLocale),
		),
	), nil
}

func CreateSafe(components Components) (Tempo, error) {
	location, err := loadLocation(components.Timezone)
	if err != nil {
		return Tempo{}, err
	}

	value := timeFromComponents(components, location)
	if !componentsMatchTime(components, value, location) {
		return Tempo{}, errors.New("invalid Tempo local date/time components")
	}

	return newTempo(
		value,
		location,
		NewRuntime(
			RuntimeLocale(tempoSettings.Locale),
			RuntimeFallbackLocale(tempoSettings.FallbackLocale),
		),
	), nil
}

func CreateStrict(components Components) (Tempo, error) {
	return CreateSafe(components)
}

func Instance(input Tempo) Tempo {
	return input.Clone()
}

func CreateFromDate(year int, month int, day int, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)
	if err != nil {
		return Tempo{}, err
	}

	return newTempo(
		timeFromComponents(Components{Year: year, Month: month, Day: day}, cfg.location),
		cfg.location,
		cfg.runtime,
	), nil
}

func CreateMidnightDate(year int, month int, day int, options ...Option) (Tempo, error) {
	return CreateFromDate(year, month, day, options...)
}

func CreateFromTime(hour int, minute int, second int, millisecond int, options ...Option) (Tempo, error) {
	today, err := Today(options...)
	if err != nil {
		return Tempo{}, err
	}

	return today.SetTime(hour, minute, second, millisecond), nil
}

func CreateFromTimeString(input string, options ...Option) (Tempo, error) {
	today, err := Today(options...)
	if err != nil {
		return Tempo{}, err
	}

	return today.SetTimeFromTimeString(input)
}

func FromObject(components Components) (Tempo, error) {
	return Create(components)
}

func FromTimestamp(timestamp int64, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)
	if err != nil {
		return Tempo{}, err
	}

	return newTempo(time.Unix(timestamp, 0), cfg.location, cfg.runtime), nil
}

func CreateFromTimestamp(timestamp int64, options ...Option) (Tempo, error) {
	return FromTimestamp(timestamp, options...)
}

func FromTimestampMs(timestamp int64, options ...Option) (Tempo, error) {
	cfg, err := applyOptions(options...)
	if err != nil {
		return Tempo{}, err
	}

	return newTempo(time.UnixMilli(timestamp), cfg.location, cfg.runtime), nil
}

func CreateFromTimestampMs(timestamp int64, options ...Option) (Tempo, error) {
	return FromTimestampMs(timestamp, options...)
}

func FromTimestampUTC(timestamp int64) (Tempo, error) {
	return FromTimestamp(timestamp, WithTimezone(defaultLocation.String()))
}

func FromTimestampMsUTC(timestamp int64) (Tempo, error) {
	return FromTimestampMs(timestamp, WithTimezone(defaultLocation.String()))
}

func CreateFromTimestampUTC(timestamp int64) (Tempo, error) {
	return FromTimestampUTC(timestamp)
}

func CreateFromTimestampMsUTC(timestamp int64) (Tempo, error) {
	return FromTimestampMsUTC(timestamp)
}

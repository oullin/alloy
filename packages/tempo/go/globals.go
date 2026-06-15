package tempo

import (
	"regexp"
	"time"
)

var (
	defaultLocation = time.UTC
	lastTempoError  error
	serializer      Serializer
	tempoSettings   = Settings{
		FallbackLocale: "en-US",
		HumanDiff:      HumanDiffOptions{Locale: "en-US", Numeric: "always", Style: "long"},
		Locale:         "en-US",
		MidDayAt:       12,
		MonthsOverflow: true,
		StrictMode:     true,
		Timezone:       defaultLocation.String(),
		WeekendDays:    []time.Weekday{time.Sunday, time.Saturday},
		YearsOverflow:  true,
	}
	toStringFormat  string
	dateOnlyPattern = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})$`)
	durationPattern = regexp.MustCompile(`^(-)?P(?:(\d+)Y)?(?:(\d+)M)?(?:(\d+)W)?(?:(\d+)D)?(?:T(?:(\d+)H)?(?:(\d+)M)?(?:(\d+(?:\.\d+)?)S)?)?$`)
	localPattern    = regexp.MustCompile(`^(\d{4})-(\d{2})-(\d{2})(?:[T\s](\d{2})(?::?(\d{2}))?(?::?(\d{2})(?:\.(\d{1,9}))?)?)?$`)
	modifierPattern = regexp.MustCompile(`^([+-]?\d+(?:\.\d+)?)\s*(milliseconds?|seconds?|minutes?|hours?|days?|weeks?|months?|quarters?|years?|decades?|centuries?|millenniums?)$`)
	movePattern     = regexp.MustCompile(`^(next|last|previous)\s+(milliseconds?|seconds?|minutes?|hours?|days?|weeks?|months?|quarters?|years?|decades?|centuries?|millenniums?)$`)
	zonePattern     = regexp.MustCompile(`(?:Z|[+-]\d{2}:?\d{2})$`)
	monthNames      = [...]string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
	shortMonthNames = [...]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	dayNames        = [...]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	shortDayNames   = [...]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
)

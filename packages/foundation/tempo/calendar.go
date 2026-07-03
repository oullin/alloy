package tempo

import "github.com/oullin/alloy/packages/foundation/tempo/calendar"

func calendarDays() []string {
	return calendar.Days()
}

func calendarMonthName(month int) string {
	return calendar.MonthName(month)
}

func calendarShortMonthName(month int) string {
	return calendar.ShortMonthName(month)
}

func calendarDayName(weekday int) string {
	return calendar.DayName(weekday)
}

func calendarShortDayName(weekday int) string {
	return calendar.ShortDayName(weekday)
}

func monthDiff(left Time, right Time, unit Unit) float64 {
	sign := 1.0
	start := right
	end := left

	if left.Before(right) {
		sign = -1
		start = left
		end = right
	}

	startObject := start.ToObject()
	endObject := end.ToObject()
	value := float64(calendar.MonthDiff(
		startObject.Year,
		startObject.Month,
		startObject.Day,
		endObject.Year,
		endObject.Month,
		endObject.Day,
	))

	switch normalizeUnit(unit) {
	case Quarter:
		value /= 3
	case Year:
		value /= 12
	}

	return value * sign
}

func daysInMonth(year int, month int) int {
	return calendar.DaysInMonth(year, month)
}

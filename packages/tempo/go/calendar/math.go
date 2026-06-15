package calendar

import "time"

func MonthDiff(startYear int, startMonth int, startDay int, endYear int, endMonth int, endDay int) int {
	months := (endYear-startYear)*12 + (endMonth - startMonth)

	if endDay < startDay {
		months--
	}

	return months
}

func DaysInMonth(year int, month int) int {
	return time.Date(year, time.Month(month)+1, 0, 0, 0, 0, 0, time.UTC).Day()
}

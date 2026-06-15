package tempo

import "github.com/oullin/alloy/tempo/temporal"

func monthDiff(left Tempo, right Tempo, unit Unit) float64 {
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
	months := (endObject.Year-startObject.Year)*12 + (endObject.Month - startObject.Month)
	if endObject.Day < startObject.Day {
		months--
	}

	value := float64(months)
	switch normalizeUnit(unit) {
	case Quarter:
		value /= 3
	case Year:
		value /= 12
	}

	return value * sign
}

func daysInMonth(year int, month int) int {
	return temporal.DaysInMonth(year, month)
}

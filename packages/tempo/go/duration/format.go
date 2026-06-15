package duration

import (
	"strconv"
	"strings"
)

func (duration Duration) ISOString() string {
	if duration.IsZero() {
		return "PT0S"
	}

	normalized := duration.Normalize()
	sign := ""

	if normalized.direction() < 0 {
		sign = "-"
	}

	value := normalized.Abs()
	dateParts := strings.Builder{}

	if value.Years != 0 {
		dateParts.WriteString(strconv.Itoa(value.Years) + "Y")
	}

	if value.Months != 0 {
		dateParts.WriteString(strconv.Itoa(value.Months) + "M")
	}

	if value.Days != 0 {
		dateParts.WriteString(strconv.Itoa(value.Days) + "D")
	}

	timeParts := strings.Builder{}

	if value.Hours != 0 {
		timeParts.WriteString(strconv.Itoa(value.Hours) + "H")
	}

	if value.Minutes != 0 {
		timeParts.WriteString(strconv.Itoa(value.Minutes) + "M")
	}

	if value.Seconds != 0 || value.Milliseconds != 0 {
		seconds := strconv.Itoa(value.Seconds)

		if value.Milliseconds != 0 {
			seconds += "." + pad(value.Milliseconds, 3)
		}

		timeParts.WriteString(seconds + "S")
	}

	result := sign + "P" + dateParts.String()

	if timeParts.Len() > 0 {
		result += "T" + timeParts.String()
	}

	return result
}

func (duration Duration) String() string {
	return duration.ISOString()
}

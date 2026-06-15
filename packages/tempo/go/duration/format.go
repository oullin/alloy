package duration

import (
	"strconv"
	"strings"
)

func (value *Duration) ISOString() string {
	if value.IsZero() {
		return "PT0S"
	}

	normalized := value.Normalize()
	sign := ""

	if normalized.direction() < 0 {
		sign = "-"
	}

	absolute := normalized.Abs()
	dateParts := strings.Builder{}

	if absolute.Years != 0 {
		dateParts.WriteString(strconv.Itoa(absolute.Years) + "Y")
	}

	if absolute.Months != 0 {
		dateParts.WriteString(strconv.Itoa(absolute.Months) + "M")
	}

	if absolute.Days != 0 {
		dateParts.WriteString(strconv.Itoa(absolute.Days) + "D")
	}

	timeParts := strings.Builder{}

	if absolute.Hours != 0 {
		timeParts.WriteString(strconv.Itoa(absolute.Hours) + "H")
	}

	if absolute.Minutes != 0 {
		timeParts.WriteString(strconv.Itoa(absolute.Minutes) + "M")
	}

	if absolute.Seconds != 0 || absolute.Milliseconds != 0 {
		seconds := strconv.Itoa(absolute.Seconds)

		if absolute.Milliseconds != 0 {
			seconds += "." + pad(absolute.Milliseconds, 3)
		}

		timeParts.WriteString(seconds + "S")
	}

	result := sign + "P" + dateParts.String()

	if timeParts.Len() > 0 {
		result += "T" + timeParts.String()
	}

	return result
}

func (value *Duration) String() string {
	return value.ISOString()
}

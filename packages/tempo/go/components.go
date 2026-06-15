package tempo

import "time"

func timeFromComponents(components Components, location *time.Location) time.Time {
	month := components.Month
	if month == 0 {
		month = 1
	}

	day := components.Day
	if day == 0 {
		day = 1
	}

	return time.Date(
		components.Year,
		time.Month(month),
		day,
		components.Hour,
		components.Minute,
		components.Second,
		components.Millisecond*int(time.Millisecond),
		location,
	)
}

func componentsMatchTime(components Components, value time.Time, location *time.Location) bool {
	month := components.Month
	if month == 0 {
		month = 1
	}

	day := components.Day
	if day == 0 {
		day = 1
	}

	local := value.In(location)
	return local.Year() == components.Year &&
		int(local.Month()) == month &&
		local.Day() == day &&
		local.Hour() == components.Hour &&
		local.Minute() == components.Minute &&
		local.Second() == components.Second &&
		local.Nanosecond()/int(time.Millisecond) == components.Millisecond
}

package parser

import "time"

type Components struct {
	Year        int
	Month       int
	Day         int
	Hour        int
	Minute      int
	Second      int
	Millisecond int
}

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
	).UTC()
}

func componentsMatchTime(components Components, value time.Time, location *time.Location) bool {
	local := value.In(location)

	if components.Month == 0 {
		components.Month = 1
	}

	if components.Day == 0 {
		components.Day = 1
	}

	return local.Year() == components.Year &&
		int(local.Month()) == components.Month &&
		local.Day() == components.Day &&
		local.Hour() == components.Hour &&
		local.Minute() == components.Minute &&
		local.Second() == components.Second &&
		local.Nanosecond()/int(time.Millisecond) == components.Millisecond
}

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

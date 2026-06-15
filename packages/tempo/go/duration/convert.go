package duration

func (duration Duration) ToMap() map[string]int {
	return map[string]int{
		"years":        duration.Years,
		"quarters":     duration.Quarters,
		"months":       duration.Months,
		"weeks":        duration.Weeks,
		"days":         duration.Days,
		"hours":        duration.Hours,
		"minutes":      duration.Minutes,
		"seconds":      duration.Seconds,
		"milliseconds": duration.Milliseconds,
	}
}

func (duration Duration) ToArray() [9]int {
	return [9]int{
		duration.Years,
		duration.Quarters,
		duration.Months,
		duration.Weeks,
		duration.Days,
		duration.Hours,
		duration.Minutes,
		duration.Seconds,
		duration.Milliseconds,
	}
}

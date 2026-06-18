package duration

func (value *Duration) ToMap() map[string]int {
	return map[string]int{
		"years":        value.Years,
		"quarters":     value.Quarters,
		"months":       value.Months,
		"weeks":        value.Weeks,
		"days":         value.Days,
		"hours":        value.Hours,
		"minutes":      value.Minutes,
		"seconds":      value.Seconds,
		"milliseconds": value.Milliseconds,
	}
}

func (value *Duration) ToSlice() []int {
	return []int{
		value.Years,
		value.Quarters,
		value.Months,
		value.Weeks,
		value.Days,
		value.Hours,
		value.Minutes,
		value.Seconds,
		value.Milliseconds,
	}
}

package calendar

var (
	monthNames      = [...]string{"January", "February", "March", "April", "May", "June", "July", "August", "September", "October", "November", "December"}
	shortMonthNames = [...]string{"Jan", "Feb", "Mar", "Apr", "May", "Jun", "Jul", "Aug", "Sep", "Oct", "Nov", "Dec"}
	dayNames        = [...]string{"Sunday", "Monday", "Tuesday", "Wednesday", "Thursday", "Friday", "Saturday"}
	shortDayNames   = [...]string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
)

func MonthName(month int) string {
	return monthNames[month-1]
}

func ShortMonthName(month int) string {
	return shortMonthNames[month-1]
}

func DayName(weekday int) string {
	return dayNames[weekday]
}

func ShortDayName(weekday int) string {
	return shortDayNames[weekday]
}

func Days() []string {
	return append([]string(nil), dayNames[:]...)
}

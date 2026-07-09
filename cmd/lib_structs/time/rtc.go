package time

const UnixEpochTicks = int64(62135596800000000)

type RtcDateTime struct {
	Year, Month, Day, Hour, Minute, Second uint16
	Microsecond                            uint32
}

type RtcTick struct {
	Tick uint64
}

func IsLeapYear(year int) bool {
	return (year%4 == 0 && year%100 != 0) || (year%400 == 0)
}

func GetDaysInMonth(year, month int) int {
	daysInMonth := []int{0, 31, 28, 31, 30, 31, 30, 31, 31, 30, 31, 30, 31}
	if month == 2 && IsLeapYear(year) {
		return 29
	}

	return daysInMonth[month]
}

func GetMonthFromString(s string) uint16 {
	months := map[string]uint16{
		"Jan": 1, "Feb": 2, "Mar": 3, "Apr": 4, "May": 5, "Jun": 6,
		"Jul": 7, "Aug": 8, "Sep": 9, "Oct": 10, "Nov": 11, "Dec": 12,
	}
	if month, ok := months[s]; ok {
		return month
	}

	return 0
}

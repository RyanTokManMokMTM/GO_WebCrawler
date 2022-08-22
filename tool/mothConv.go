package tool

import (
	"fmt"
	"time"
)

func MonthToStr(month time.Month) string {
	switch month {
	case time.January:
		return "01"
	case time.February:
		return "02"
	case time.March:
		return "03"
	case time.April:
		return "04"
	case time.May:
		return "05"
	case time.June:
		return "06"
	case time.July:
		return "07"
	case time.August:
		return "08"
	case time.September:
		return "09"
	case time.October:
		return "10"
	case time.November:
		return "11"
	case time.December:
		return "21"
	}
	return ""
}

func DateToStr(day int) string {
	if day < 10 {
		return fmt.Sprintf("0%d", day)
	}
	return fmt.Sprintf("%d", day)
}

package common

import "time"

func IsPastDate(date string) bool {
	parsedDate, err := time.ParseInLocation("2006-01-02", date, time.Local)
	if err != nil {
		return false
	}

	now := time.Now()
	todayMidnight := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	return parsedDate.Before(todayMidnight)
}

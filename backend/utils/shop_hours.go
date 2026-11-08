package utils

import (
	"regexp"
	"time"
)

var clockPattern = regexp.MustCompile(`^([01][0-9]|2[0-3]):[0-5][0-9]$`)

func ValidShopHours(open, close string) bool {
	return clockPattern.MatchString(open) && (clockPattern.MatchString(close) || close == "24:00") && open < close
}
func WithinShopHours(start, end time.Time, open, close string) bool {
	if !ValidShopHours(open, close) {
		return false
	}
	zone := time.FixedZone("Asia/Bangkok", 7*3600)
	day := start.In(zone).Format("2006-01-02")
	opening, err := time.Parse(time.RFC3339, day+"T"+open+":00+07:00")
	if err != nil {
		return false
	}
	var closing time.Time
	if close == "24:00" {
		midnight, _ := time.Parse(time.RFC3339, day+"T00:00:00+07:00")
		closing = midnight.AddDate(0, 0, 1)
	} else {
		closing, err = time.Parse(time.RFC3339, day+"T"+close+":00+07:00")
		if err != nil {
			return false
		}
	}
	return !start.Before(opening) && !end.After(closing) && end.After(start)
}

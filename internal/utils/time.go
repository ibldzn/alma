package utils

import (
	"time"

	"github.com/ibldzn/alma/internal/constants"
)

func IsDateEqual(d1, d2 time.Time) bool {
	y1, m1, day1 := d1.Date()
	y2, m2, day2 := d2.Date()

	return y1 == y2 && m1 == m2 && day1 == day2
}

func GetTodayInJakarta() time.Time {
	return time.Now().In(time.FixedZone(constants.AsiaJakarta, 7*60*60))
}

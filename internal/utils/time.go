package utils

import (
	"fmt"
	"time"

	"github.com/ibldzn/alma/internal/constants"
	"github.com/ibldzn/alma/internal/types"
)

func ValidateDateRange(startDate, endDate string) (time.Time, time.Time, error) {
	start, err := ParseDateInJakarta(startDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: start_date=%q: %v", types.ErrInvalidDateFormat, startDate, err)
	}

	end, err := ParseDateInJakarta(endDate)
	if err != nil {
		return time.Time{}, time.Time{}, fmt.Errorf("%w: end_date=%q: %v", types.ErrInvalidDateFormat, endDate, err)
	}

	if end.Before(start) {
		return time.Time{}, time.Time{}, types.ErrInvalidDateRange
	}

	return start, end, nil
}

func JakartaLocation() *time.Location {
	return time.FixedZone(constants.AsiaJakarta, 7*60*60)
}

func ParseDateInJakarta(date string) (time.Time, error) {
	return time.ParseInLocation(constants.DateFormat, date, JakartaLocation())
}

func IsDateEqual(d1, d2 time.Time) bool {
	y1, m1, day1 := d1.Date()
	y2, m2, day2 := d2.Date()

	return y1 == y2 && m1 == m2 && day1 == day2
}

func GetTodayInJakarta() time.Time {
	return time.Now().In(JakartaLocation())
}

func GetTodayDateInJakarta() time.Time {
	now := GetTodayInJakarta()
	year, month, day := now.Date()
	return time.Date(year, month, day, 0, 0, 0, 0, now.Location())
}

package types

import "errors"

var (
	ErrInvalidDateRange           = errors.New("invalid date range: end date cannot be before start date")
	ErrInvalidDateFormat          = errors.New("invalid date format: expected YYYY-MM-DD")
	ErrExpectedStructOrMap        = errors.New("invalid type: expected a struct or map")
	ErrInvalidData                = errors.New("invalid data")
	ErrUnableToMapDwhToAppModel   = errors.New("unable to map DWH model to application model")
	ErrEndDateCannotBeInTheFuture = errors.New("end date cannot be in the future")
)

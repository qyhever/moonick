package timeutil

import (
	"fmt"
	"time"
)

func ParseDepartureDate(value string, location *time.Location) (time.Time, error) {
	if location == nil {
		location = time.Local
	}

	departureDate, err := time.ParseInLocation(time.DateOnly, value, location)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse departure date: %w", err)
	}
	return departureDate, nil
}

func CombineDeparture(date time.Time, clock string, location *time.Location) (time.Time, error) {
	if location == nil {
		location = time.Local
	}

	clockTime, err := time.ParseInLocation("15:04", clock, location)
	if err != nil {
		return time.Time{}, fmt.Errorf("parse departure clock: %w", err)
	}

	return time.Date(
		date.Year(), date.Month(), date.Day(),
		clockTime.Hour(), clockTime.Minute(), 0, 0, location,
	), nil
}

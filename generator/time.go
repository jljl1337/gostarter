package generator

import (
	"time"

	"github.com/jljl1337/gostarter/format"
)

/*
NowISO8601 returns the current time in ISO 8601 format.
*/
func NowISO8601() string {
	return format.TimeToISO8601(time.Now())
}

/*
NowISO8601Number returns the current time in ISO 8601 format without
separators.
*/
func NowISO8601Number() string {
	return format.TimeToISO8601Number(time.Now())
}

/*
MinutesFromNowISO8601 returns the time in ISO 8601 format for the specified
number of minutes from now.
*/
func MinutesFromNowISO8601(minutes int) string {
	return DurationFromNowISO8601(time.Duration(minutes) * time.Minute)
}

/*
DurationFromNowISO8601 returns the time in ISO 8601 format for the specified
duration from now.
*/
func DurationFromNowISO8601(duration time.Duration) string {
	return format.TimeToISO8601(time.Now().Add(duration))
}

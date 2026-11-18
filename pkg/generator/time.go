package generator

import (
	"math"
	"math/rand"
	"time"

	"github.com/jljl1337/gostarter/pkg/format"
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

func MinutesBeforeNowISO8601(minutes int) string {
	return DurationBeforeNowISO8601(time.Duration(minutes) * time.Minute)
}

func DurationBeforeNowISO8601(duration time.Duration) string {
	return format.TimeToISO8601(time.Now().Add(-duration))
}

// Backoff calculates the backoff duration based on the base duration and the
// number of failures, using exponential backoff with jitter.
func Backoff(base time.Duration, failures int) time.Duration {
	if base <= 0 || failures <= 0 {
		return 0
	}

	delay := base

	for i := 1; i < failures; i++ {
		if delay > time.Duration(math.MaxInt64/2) {
			break
		}
		delay *= 2
	}

	half := delay / 2

	return half + time.Duration(rand.Int63n(int64(half)+1))
}

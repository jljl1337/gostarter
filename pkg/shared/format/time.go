package format

import "time"

/*
TimeToISO8601 converts a time.Time object to an ISO 8601 formatted string in
UTC.
*/
func TimeToISO8601(timestamp time.Time) string {
	return timestamp.UTC().Format("2006-01-02T15:04:05.000Z")
}

/*
ISO8601ToTime converts an ISO 8601 formatted string in UTC to a time.Time
object.
*/
func ISO8601ToTime(timestamp string) (time.Time, error) {
	return time.Parse("2006-01-02T15:04:05.000Z", timestamp)
}

/*
TimeToISO8601Number converts a time.Time object to an ISO 8601 formatted string
without separators in UTC.
*/
func TimeToISO8601Number(timestamp time.Time) string {
	return timestamp.UTC().Format("20060102150405")
}

/*
ISO8601NumberToTime converts an ISO 8601 formatted string without separators in
UTC to a time.Time object.
*/
func ISO8601NumberToTime(timestamp string) (time.Time, error) {
	return time.Parse("20060102150405", timestamp)
}

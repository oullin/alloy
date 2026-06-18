package factory

import "time"

func FromTimestamp(timestamp int64) time.Time {
	return time.Unix(timestamp, 0)
}

func FromTimestampMs(timestamp int64) time.Time {
	return time.UnixMilli(timestamp)
}

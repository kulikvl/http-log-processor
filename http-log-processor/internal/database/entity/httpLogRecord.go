package entity

import (
	"time"
)

type HttpLogRecord struct {
	Timestamp        time.Time
	ResourceID       uint64
	BytesSent        uint64
	RequestTimeMilli uint64
	ResponseStatus   uint16
	CacheStatus      string
	Method           string
	RemoteAddr       string
	URL              string
}

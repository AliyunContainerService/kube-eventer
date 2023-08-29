package util

import (
	"k8s.io/api/core/v1"
	"time"
)

func GetLastEventTimestamp(event *v1.Event) time.Time {

	if !event.LastTimestamp.IsZero() {
		return event.LastTimestamp.Time
	}

	if !event.EventTime.IsZero() {
		return event.EventTime.Time
	}

	return time.Now()
}

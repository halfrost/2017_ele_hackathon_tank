package ttracker

import (
	tracker "github.com/eleme/thrift-tracker"
)

type nexTracker struct {
	*tracker.SimpleTracker
}

// NewTracker creates a SimpleTracker, code at here act as a placeholder.
func NewTracker(name string) tracker.Tracker {
	return &nexTracker{
		SimpleTracker: tracker.NewSimpleTracker(name).(*tracker.SimpleTracker),
	}
}

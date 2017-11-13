package etrace

import "github.com/eleme/etrace-go/metric"

// NewCounter returns a new counter.
func (t *Trace) NewCounter(name string) *metric.Counter {
	return t.mc.NewCounter(name)
}

// NewTimer returns a new timer.
func (t *Trace) NewTimer(name string) *metric.Timer {
	return t.mc.NewTimer(name)
}

// NewGauge returns a new gauge.
func (t *Trace) NewGauge(name string) *metric.Gauge {
	return t.mc.NewGauge(name)
}

// NewPayload returns a new payload.
func (t *Trace) NewPayload(name string) *metric.Payload {
	return t.mc.NewPayload(name)
}

package metric

import (
	"strings"
	"time"

	"github.com/eleme/statslib"
)

// GoUnusedProtection is used to ensure import safety.
const GoUnusedProtection = ""
const defaultReportInterval = 2 * time.Second

// Start starts the statsd service with statsd addr. Enable CPU and runtime profiling.
func Start(statsdAddr string) error {
	return StartWithOptions(statsdAddr, true, true)
}

// StartWithOptions starts the statsd service with statsd addr, cpu and runtime options.
func StartWithOptions(statsdAddr string, enableCPU bool, enableRuntime bool) error {
	statsdAddr = strings.TrimPrefix(statsdAddr, "statsd://")
	cfg := statsd.DefaultConfig(statsdAddr)
	cfg.EnableCPUMetrics = enableCPU
	cfg.EnableRuntimeMetrics = enableRuntime
	cfg.Interval = defaultReportInterval
	if err := statsd.StartStatsdService(cfg); err != nil {
		return err
	}
	return nil
}

// Stop stop the statsd service.
func Stop() {
	statsd.Close()
}

// MeasureSince mesures one stat timer with begin time.
func MeasureSince(metric string, begin time.Time) (time.Duration, error) {
	return statsd.MeasureSince(metric, begin)
}

// CounterInc increments one stat counter.
func CounterInc(metric string, value int) {
	statsd.Increment(metric, value)
}

// SetGauge sets one stat gauges.
func SetGauge(metric string, value int) {
	statsd.SetGaugeInt(metric, value)
}

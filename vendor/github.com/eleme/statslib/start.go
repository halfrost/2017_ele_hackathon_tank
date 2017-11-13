package statsd

import (
	"errors"
	"time"
)

var (
	service = &StatsdService{}

	ErrEmptyConfig        = errors.New("Please indicate config")
	ErrStatsdAddr         = errors.New("StatsdAddr could not be empty")
	ErrNotRunning         = errors.New("StatsdService was not running")
	ErrUnimplementsLogger = errors.New("logger not implements Logger interface")
)

// StartStatsdService start a statsd collect service, please regist a custom logger by call RegisterLogger
func StartStatsdService(conf *Config) error {
	if conf == nil {
		return ErrEmptyConfig
	}
	if conf.StatsdAddr == "" {
		return ErrStatsdAddr
	}

	service = NewStatsdService(conf)
	service.Start()
	return nil
}

type Logger interface {
	Printf(format string, v ...interface{})
}

// RegisterLogger register a logger in this package, logger must implements Logger
func RegisterLogger(log Logger) {
	logger = log
}

// Avaliable runtime metrics: "num_goroutines", "gc_pause_ms", "gc_pause_total_ms", "alloc_bytes",
// "total_alloc_bytes", "sys_bytes", "heap_alloc_bytes", "heap_sys_bytes", "heap_idle_bytes",
// "heap_inuse_bytes", "heap_released_bytes", "heap_objects", "stack_inuse_bytes", "stack_sys_bytes",
// "num_gc", "lookups", "mallocs", "frees"
func DisalbeRuntimeMetrics(runtimeMetrics ...string) {
	for _, disalbedMetric := range runtimeMetrics {
		if _, ok := service.runtimeMetrics[disalbedMetric]; ok {
			delete(service.runtimeMetrics, disalbedMetric)
		}
	}
}

func EnableHostname() {
	service.EnableHostname()
}

func DisableHostname() {
	service.DisableHostname()
}

func Close() {
	if service.IsRunning() {
		service.Stop()
	}
}

/// Gauges ///

func SetGaugeFloat64(metric string, val float64) error {
	return SetGaugesFloat64([]string{metric}, val)
}

func SetGaugesFloat64(keys []string, val float64) error {
	if !service.IsRunning() {
		return ErrNotRunning
	}
	service.SetGaugeFloat64(keys, val, true)
	return nil
}

func SetGaugeFloat64NoHost(metric string, val float64) error {
	return SetGaugesFloat64NoHost([]string{metric}, val)
}

func SetGaugesFloat64NoHost(keys []string, val float64) error {
	if !service.IsRunning() {
		return ErrNotRunning
	}
	service.SetGaugeFloat64(keys, val, false)
	return nil
}

func SetGaugeUInt64(metric string, val uint64) error {
	return SetGaugesUInt64([]string{metric}, val)
}

func SetGaugesUInt64(keys []string, val uint64) error {
	if !service.IsRunning() {
		return ErrNotRunning
	}
	service.SetGaugeUInt64(keys, val, true)
	return nil
}

func SetGaugeUInt64NoHost(metric string, val uint64) error {
	return SetGaugesUInt64NoHost([]string{metric}, val)
}

func SetGaugesUInt64NoHost(keys []string, val uint64) error {
	if !service.IsRunning() {
		return ErrNotRunning
	}
	service.SetGaugeUInt64(keys, val, false)
	return nil
}

func SetGaugeInt(metric string, val int) error {
	return SetGaugesInt([]string{metric}, val)
}

func SetGaugesInt(keys []string, val int) error {
	if !service.IsRunning() {
		return ErrNotRunning
	}
	service.SetGaugeInt(keys, val, true)
	return nil
}

func SetGaugeIntNoHost(metric string, val int) error {
	return SetGaugesIntNoHost([]string{metric}, val)
}

func SetGaugesIntNoHost(keys []string, val int) error {
	if !service.IsRunning() {
		return ErrNotRunning
	}
	service.SetGaugeInt(keys, val, false)
	return nil
}

/// Counters ///

func Increment(metric string, val int) error {
	return IncrCounters([]string{metric}, val)
}

func IncrCounters(keys []string, val int) error {
	if !service.IsRunning() {
		return ErrNotRunning
	}
	service.IncrCounter(keys, val, true)
	return nil
}

func IncrementNoHost(metric string, val int) error {
	return IncrCountersNoHost([]string{metric}, val)
}

func IncrCountersNoHost(keys []string, val int) error {
	if !service.IsRunning() {
		return ErrNotRunning
	}
	service.IncrCounter(keys, val, false)
	return nil
}

/// Timers ///

func AddSample(metric string, val float64) error {
	return AddSamples([]string{metric}, val)
}

func AddSamples(keys []string, val float64) error {
	if !service.IsRunning() {
		return ErrNotRunning
	}
	service.AddSample(keys, val, true)
	return nil
}

func AddSampleNoHost(metric string, val float64) error {
	return AddSamplesNoHost([]string{metric}, val)
}

func AddSamplesNoHost(keys []string, val float64) error {
	if !service.IsRunning() {
		return ErrNotRunning
	}
	service.AddSample(keys, val, false)
	return nil
}

func MeasureSince(metric string, start time.Time) (time.Duration, error) {
	return MeasureSinceKeys([]string{metric}, start)
}

func MeasureSinceKeys(keys []string, start time.Time) (time.Duration, error) {
	if !service.IsRunning() {
		return 0, ErrNotRunning
	}
	return service.MeasureSince(keys, start, true), nil
}

func MeasureSinceNoHost(metric string, start time.Time) error {
	return MeasureSinceKeysNoHost([]string{metric}, start)
}

func MeasureSinceKeysNoHost(keys []string, start time.Time) error {
	if !service.IsRunning() {
		return ErrNotRunning
	}
	service.MeasureSince(keys, start, false)
	return nil
}

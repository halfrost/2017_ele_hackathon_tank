// Copyright 2016 Eleme Inc. All rights reserved.

package circuitbreaker

import (
	"errors"
	"time"
)

// Errors
var (
	ErrOptionsServiceName                                 = errors.New("circuitbreaker: options service name")
	ErrOptionsRecoveryTime                                = errors.New("circuitbreaker: options recovery time")
	ErrOptionsNumCallsTriggerPerInterval                  = errors.New("circuitbreaker: options num calls trigger")
	ErrOptionsPercentageTimeoutErrorsThresholdPerInterval = errors.New("circuitbreaker: options percentage timeout errors threshold")
	ErrOptionsPercentageSystemErrorsThresholdPerInterval  = errors.New("circuitbreaker: options percentage system errors threshold")
	ErrOptionsPercentageUnknownErrorsThresholdPerInterval = errors.New("circuitbreaker: options percentage unknown errors threshold")
)

// Defaults
const (
	DefaultMinRecoveryTime                             = 20 * time.Second
	DefaultMaxRecoveryTime                             = 2 * time.Minute
	DefaultNumCallsTriggerPerInterval                  = 10
	DefaultMetricRollingSize                           = 10
	DefaultMetricRollingGranularity                    = 3
	DefaultPercentageTimeoutErrorsThresholdPerInterval = 0.5
	DefaultPercentageSystemErrorsThresholdPerInterval  = 0.5
	DefaultPercentageUnknownErrorsThresholdPerInterval = 0.5
)

// Options is the CircuitBreaker options.
type Options struct {
	// ServiceName is the service name for circuitbreaker to work with.
	ServiceName string
	// MinRecoveryTime is the shortest time to release the unhealthy api, a
	// locked api won't be released if the locked time is too short.
	MinRecoveryTime time.Duration
	// MaxRecoveryTime is the longest time to release the unhealthy api, a
	// locked api must be released if the locked time is too long.
	MaxRecoveryTime time.Duration
	// MetricRollingSize is the metric rolling number window size.
	MetricRollingSize int
	// MetricRollingGranularity is the metric rolling number granularity in
	// seconds.
	MetricRollingGranularity int
	// NumCallsTriggerPerInterval is the minuim number of api calls in an
	// interval to trigger the circuit breaker to work.
	NumCallsTriggerPerInterval int
	// PercentageTimeoutErrorsThresholdPerInterval is the minuim number of api
	// timeout errors percentage in an interval to trigger the circuit breaker
	// to lock the api.
	PercentageTimeoutErrorsThresholdPerInterval float64
	// PercentageSystemErrorsThresholdPerInterval is the minuim number of api
	// system errors percentage in an interval to trigger the circuit breaker
	// to lock the api.
	PercentageSystemErrorsThresholdPerInterval float64
	// PercentageUnknownErrorsThresholdPerInterval is the minuim number of api
	// unknown errors percentage in an interval to trigger the circuit breaker
	// to lock the api.
	PercentageUnknownErrorsThresholdPerInterval float64
}

// NewOptionsWithDefaults creates options with default values.
func NewOptionsWithDefaults() *Options {
	return &Options{
		MinRecoveryTime:                             DefaultMinRecoveryTime,
		MaxRecoveryTime:                             DefaultMaxRecoveryTime,
		MetricRollingSize:                           DefaultMetricRollingSize,
		MetricRollingGranularity:                    DefaultMetricRollingGranularity,
		NumCallsTriggerPerInterval:                  DefaultNumCallsTriggerPerInterval,
		PercentageTimeoutErrorsThresholdPerInterval: DefaultPercentageTimeoutErrorsThresholdPerInterval,
		PercentageSystemErrorsThresholdPerInterval:  DefaultPercentageSystemErrorsThresholdPerInterval,
		PercentageUnknownErrorsThresholdPerInterval: DefaultPercentageUnknownErrorsThresholdPerInterval,
	}
}

// Update the options with another.
func (options *Options) Update(opts *Options) {
	if opts != nil {
		options.ServiceName = opts.ServiceName
		options.MinRecoveryTime = opts.MinRecoveryTime
		options.MaxRecoveryTime = opts.MaxRecoveryTime
		options.MetricRollingSize = opts.MetricRollingSize
		options.MetricRollingGranularity = opts.MetricRollingGranularity
		options.NumCallsTriggerPerInterval = opts.NumCallsTriggerPerInterval
		options.PercentageTimeoutErrorsThresholdPerInterval = opts.PercentageTimeoutErrorsThresholdPerInterval
		options.PercentageSystemErrorsThresholdPerInterval = opts.PercentageSystemErrorsThresholdPerInterval
		options.PercentageUnknownErrorsThresholdPerInterval = opts.PercentageUnknownErrorsThresholdPerInterval
	}
}

// Validate the options.
func (options *Options) Validate() error {
	if options.ServiceName == "" {
		return ErrOptionsServiceName
	}
	if options.MaxRecoveryTime <= options.MinRecoveryTime {
		return ErrOptionsRecoveryTime
	}
	if options.NumCallsTriggerPerInterval < 0 {
		return ErrOptionsNumCallsTriggerPerInterval
	}
	if options.PercentageTimeoutErrorsThresholdPerInterval <= 0 || options.PercentageTimeoutErrorsThresholdPerInterval >= 1 {
		return ErrOptionsPercentageTimeoutErrorsThresholdPerInterval
	}
	if options.PercentageSystemErrorsThresholdPerInterval <= 0 || options.PercentageSystemErrorsThresholdPerInterval >= 1 {
		return ErrOptionsPercentageSystemErrorsThresholdPerInterval
	}
	if options.PercentageUnknownErrorsThresholdPerInterval <= 0 || options.PercentageUnknownErrorsThresholdPerInterval >= 1 {
		return ErrOptionsPercentageUnknownErrorsThresholdPerInterval
	}
	return nil
}

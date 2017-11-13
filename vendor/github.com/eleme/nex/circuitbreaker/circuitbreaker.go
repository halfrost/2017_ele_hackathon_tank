package circuitbreaker

import (
	"reflect"
	"time"

	"github.com/eleme/circuitbreaker"
	"github.com/eleme/nex/app"
	"github.com/eleme/nex/endpoint"
	"github.com/eleme/nex/timeout"
)

var (
	defaultOptions = &circuitbreaker.Options{
		MinRecoveryTime:                             10 * time.Second,
		MaxRecoveryTime:                             60 * time.Second,
		MetricRollingSize:                           15,
		MetricRollingGranularity:                    4,
		NumCallsTriggerPerInterval:                  10,
		PercentageTimeoutErrorsThresholdPerInterval: 0.5,
		PercentageSystemErrorsThresholdPerInterval:  0.5,
		PercentageUnknownErrorsThresholdPerInterval: 0.5,
	}
	// ErrAPINotHealthy returns from circuitbreaker when api is not healthy.
	ErrAPINotHealthy = app.NewUnknownException("NotHealthy", "api is not healthy")
)

// CircuitBreaker is a wrapper of circuitbreaker.CircuitBreaker, for easy use
type CircuitBreaker struct {
	*circuitbreaker.CircuitBreaker
	appErrTypes *endpoint.ErrTypes
}

// New creates a new CircuitBreaker.
func New(serviceName string, appErrTypes *endpoint.ErrTypes) *CircuitBreaker {
	options := &circuitbreaker.Options{}
	options.Update(defaultOptions)
	options.ServiceName = serviceName
	return &CircuitBreaker{
		CircuitBreaker: circuitbreaker.New(options),
		appErrTypes:    appErrTypes,
	}
}

// Test tests the api whether it is healthy.
func (cb *CircuitBreaker) Test(api string) error {
	if !cb.CircuitBreaker.Test(api) {
		return ErrAPINotHealthy
	}
	return nil
}

// RecordMetrics records the metrics.
func (cb *CircuitBreaker) RecordMetrics(api string, errPointer *error) {
	cb.CircuitBreaker.AfterAPICalled(api)

	err := *errPointer
	if err == nil {
		cb.CircuitBreaker.AfterAPICalledOk(api)
	} else if err == ErrAPINotHealthy {
		return
	} else if err == timeout.ErrTimeout {
		cb.CircuitBreaker.AfterAPICalledTimeoutError(api)
	} else if cb.appErrTypes != nil {
		switch reflect.TypeOf(err) {
		case cb.appErrTypes.UserErr:
			cb.CircuitBreaker.AfterAPICalledUserError(api)
		case cb.appErrTypes.SysErr:
			cb.CircuitBreaker.AfterAPICalledSystemError(api)
		default:
			cb.CircuitBreaker.AfterAPICalledUnknownError(api)
		}
	}
}

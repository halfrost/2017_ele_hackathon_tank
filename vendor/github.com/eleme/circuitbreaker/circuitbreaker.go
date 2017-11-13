// Copyright 2016 Eleme Inc. All rights reserved.

/*

Package circuitbreaker ports the health checking to golang from eleme/zeus_core.

Example

	c := circuitbreaker.New(nil)
	if !c.Test("foo") {
		return errors.New("api not health!")
	}
	err := apiFoo()
	c.AfterAPICalled("foo")
	switch err {
	case nil:
		c.AfterAPICalledOk("foo")
	case ErrTimeout:
		c.AfterAPICalledTimeoutError("foo")
	case ErrUserError:
		c.AfterAPICalledUserError("foo")
	case ErrSystemError:
		c.AfterAPICalledSystemError("foo")
	case ErrUnknownError:
		c.AfterAPICalledUnknownError("foo")
	}

Goroutine Safety

Yes.

*/
package circuitbreaker

import (
	"sync"
	"time"
)

// TestContext is the context to hold runtime data on method Test.
type TestContext struct {
	startAt     time.Time
	endAt       time.Time
	serviceName string
	apiName     string
	lockedAt    time.Time // 0 for unlocked.
	result      bool      // true for ok.
}

// StartAt returns the time the Test starts.
func (ctx *TestContext) StartAt() time.Time {
	return ctx.startAt
}

// EndAt returns the time the Test ends.
func (ctx *TestContext) EndAt() time.Time {
	return ctx.endAt
}

// ServiceName returns the service name.
func (ctx *TestContext) ServiceName() string {
	return ctx.serviceName
}

// APIName returns the api name for the Test.
func (ctx *TestContext) APIName() string {
	return ctx.apiName
}

// LockedAt returns the time the api is locked.
func (ctx *TestContext) LockedAt() time.Time {
	return ctx.lockedAt
}

// Result returns the test result.
func (ctx *TestContext) Result() bool {
	return ctx.result
}

// TestCallback is the type of callback functions to be called with Test
// context.
type TestCallback func(ctx *TestContext)

type callbacks struct {
	beforeTest       TestCallback
	afterTest        TestCallback
	afterAPILocked   TestCallback
	afterAPIUnlocked TestCallback
}

// CircuitBreaker is to the circuit breaker.
type CircuitBreaker struct {
	sync.RWMutex
	options  *Options
	apiStats map[string]*apiStats
	// Test callbacks
	cbs *callbacks
}

// New creates a new CircuitBreaker.
func New(options *Options) *CircuitBreaker {
	opts := NewOptionsWithDefaults()
	opts.Update(options)
	return &CircuitBreaker{
		options:  opts,
		apiStats: make(map[string]*apiStats, 10),
		cbs:      &callbacks{},
	}
}

func (c *CircuitBreaker) apiStatsBy(apiName string) *apiStats {
	c.RLock()
	if a, in := c.apiStats[apiName]; in {
		c.RUnlock()
		return a
	}
	c.RUnlock()

	c.Lock()
	defer c.Unlock()
	if a, in := c.apiStats[apiName]; in {
		return a
	}
	a := newAPIStats(apiName, *c.options)
	c.apiStats[apiName] = a
	return a
}

// AfterAPICalled tells the circuitbreaker to increment the api number of calls
// by one.
func (c *CircuitBreaker) AfterAPICalled(apiName string) {
	a := c.apiStatsBy(apiName)
	a.calledCounter.Increment(1)
}

// AfterAPICalledOk tells the circuitbreaker to set the api latest state to
// ok.
func (c *CircuitBreaker) AfterAPICalledOk(apiName string) {
	a := c.apiStatsBy(apiName)
	a.setLatestCallStatus(true)
}

// AfterAPICalledUserError tells the circuitbreaker to set the api latest state to
// ok.
func (c *CircuitBreaker) AfterAPICalledUserError(apiName string) {
	a := c.apiStatsBy(apiName)
	a.setLatestCallStatus(true)
}

// AfterAPICalledTimeoutError tells the circuitbreaker to increment the api
// number of timeout errors by one, and set the api latest state to bad.
func (c *CircuitBreaker) AfterAPICalledTimeoutError(apiName string) {
	a := c.apiStatsBy(apiName)
	a.timeoutCounter.Increment(1)
	a.setLatestCallStatus(false)
}

// AfterAPICalledSystemError tells the circuitbreaker to increment the api
// number of system errors by one, and set the api latest state to bad.
func (c *CircuitBreaker) AfterAPICalledSystemError(apiName string) {
	a := c.apiStatsBy(apiName)
	a.syserrCounter.Increment(1)
	a.setLatestCallStatus(false)
}

// AfterAPICalledUnknownError tells the circuitbreaker to increment the api
// number of unknown errors by one, and set the api latest state to bad.
func (c *CircuitBreaker) AfterAPICalledUnknownError(apiName string) {
	a := c.apiStatsBy(apiName)
	a.unkwnerrCounter.Increment(1)
	a.setLatestCallStatus(false)
}

// BeforeTest adds a callback that would be called before the Test is
// performed.
func (c *CircuitBreaker) BeforeTest(cb TestCallback) {
	c.cbs.beforeTest = cb
}

// AfterTest adds a callback that would be called right after the Test is
// completed.
func (c *CircuitBreaker) AfterTest(cb TestCallback) {
	c.cbs.afterTest = cb
}

// AfterAPILocked adds a callback that would be called after an api is
// locked
func (c *CircuitBreaker) AfterAPILocked(cb TestCallback) {
	c.cbs.afterAPILocked = cb
}

// AfterAPIUnlocked adds a callback that would be called after an api is
// unlocked.
func (c *CircuitBreaker) AfterAPIUnlocked(cb TestCallback) {
	c.cbs.afterAPIUnlocked = cb
}

// IsHealthy checks current api health status by metrics, returns true for
// status ok.
// Returns false only if the timeout, system and unknown errors are all below
// the percentage threshold.
func (c *CircuitBreaker) IsHealthy(apiName string) bool {
	a := c.apiStatsBy(apiName)
	return a.IsHealthy(*c.options)
}

// Test current api health before the request is processed, returns true for
// ok.
//
// Detail logic notes:
//
//	1. If current api is unlocked, lock it until IsHealthy returns false.
//	2. If current api is locked, recover it until IsHealthy returns true and
//	   the locked time span is not too short. One request will be released for
//	   health checking once this aoi enters recover mode.
//	3. If current api is in recover mode, try to unlock it if the latest
//	   request (the request just released) executed without errors.
//	4. Requests on an apiare unlocked gradually, but not immediately. It allows
//	   more requests to pass as the time becomes longer from the time turns to
//	   health ok, but it will be unlocked anyway when the time span is over the
//	   max recover time.
//
func (c *CircuitBreaker) Test(apiName string) bool {
	a := c.apiStatsBy(apiName)
	return a.Test(*c.options, c.cbs)
}

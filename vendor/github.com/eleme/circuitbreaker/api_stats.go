package circuitbreaker

import (
	"math/rand"
	"sync"
	"time"

	"github.com/eleme/circuitbreaker/metric"
)

const (
	// statusUnlocked is to tell that the target api is currently unlocked.
	statusUnlocked = iota
	// statusLocked is to tell that the target api is currently locked.
	statusLocked
	// statusRecovering is to tell that the target api is currently recovering
	// from locked to unlocked.
	statusRecovering
)

type apiStats struct {
	sync.Mutex
	name       string
	lockedAt   time.Time
	lockStatus int
	latestOk   bool

	calledCounter   *metric.RollingNumber
	timeoutCounter  *metric.RollingNumber
	syserrCounter   *metric.RollingNumber
	unkwnerrCounter *metric.RollingNumber
}

func newAPIStats(apiName string, opts Options) *apiStats {
	return &apiStats{
		name:            apiName,
		lockedAt:        time.Time{},
		lockStatus:      statusUnlocked,
		latestOk:        true,
		calledCounter:   metric.NewRollingNumber(opts.MetricRollingSize, opts.MetricRollingGranularity),
		timeoutCounter:  metric.NewRollingNumber(opts.MetricRollingSize, opts.MetricRollingGranularity),
		syserrCounter:   metric.NewRollingNumber(opts.MetricRollingSize, opts.MetricRollingGranularity),
		unkwnerrCounter: metric.NewRollingNumber(opts.MetricRollingSize, opts.MetricRollingGranularity),
	}
}

func (a *apiStats) setLatestCallStatus(ok bool) {
	a.Lock()
	a.latestOk = ok
	a.Unlock()
}

func (a *apiStats) Test(opts Options, cbs *callbacks) bool {
	a.Lock()
	defer a.Unlock()
	timeNow := time.Now()
	ctx := &TestContext{}
	ctx.startAt = timeNow
	ctx.serviceName = opts.ServiceName
	ctx.apiName = a.name
	ctx.lockedAt = a.lockedAt
	if cbs.beforeTest != nil {
		cbs.beforeTest(ctx)
	}

	if a.lockStatus == statusLocked && a.IsHealthy(opts) && timeNow.Sub(a.lockedAt) >= opts.MinRecoveryTime { // Turns to OK.
		// Enter into recover mode.
		a.lockStatus = statusRecovering
		// Release this request for health check.
		ctx.result = true
	} else if a.lockStatus == statusRecovering {
		if a.latestOk {
			lockedTimeSpan := timeNow.Sub(a.lockedAt)
			if lockedTimeSpan >= opts.MaxRecoveryTime {
				a.lockedAt = time.Time{}
				a.lockStatus = statusUnlocked
				ctx.result = true
				if cbs.afterAPILocked != nil {
					cbs.afterAPIUnlocked(ctx)
				}
			} else {
				if rand.Float64() < float64(lockedTimeSpan)/float64(opts.MaxRecoveryTime) {
					// Allow pass gradually.
					ctx.result = true
				} else {
					// Not lucky.
					ctx.result = false
				}
			}
		} else { // Still suffering, lock it again.
			a.lockedAt = timeNow
			a.lockStatus = statusLocked
			ctx.lockedAt = timeNow
			ctx.result = false
			if cbs.afterAPILocked != nil {
				cbs.afterAPILocked(ctx)
			}
		}
	} else if a.lockStatus == statusUnlocked {
		if !a.IsHealthy(opts) { // Turns BAD.
			a.lockedAt = timeNow
			a.lockStatus = statusLocked
			ctx.lockedAt = timeNow
			ctx.result = false
			if cbs.afterAPILocked != nil {
				cbs.afterAPILocked(ctx)
			}
		} else {
			ctx.result = true
		}
	}

	ctx.endAt = time.Now()
	if cbs.afterTest != nil {
		cbs.afterTest(ctx)
	}
	return ctx.result
}

func (a *apiStats) IsHealthy(opts Options) bool {
	nCalls := a.calledCounter.Value()
	nTimeoutErrs := a.timeoutCounter.Value()
	nSysErrs := a.syserrCounter.Value()
	nUnknnErrs := a.unkwnerrCounter.Value()
	if nCalls >= opts.NumCallsTriggerPerInterval {
		return (float64(nTimeoutErrs)/float64(nCalls) < opts.PercentageTimeoutErrorsThresholdPerInterval &&
			float64(nSysErrs)/float64(nCalls) < opts.PercentageSystemErrorsThresholdPerInterval &&
			float64(nUnknnErrs)/float64(nCalls) < opts.PercentageUnknownErrorsThresholdPerInterval)
	}
	return true
}

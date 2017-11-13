package timeout

import (
	"context"
	"fmt"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/eleme/huskar/config"
	"github.com/eleme/nex/app"
	"github.com/eleme/nex/consts/ctxkeys"
	"github.com/eleme/nex/consts/huskarkeys"
	"github.com/eleme/nex/endpoint"
)

var (
	// ErrTimeout is used for context time out.
	ErrTimeout = app.NewUnknownException("Timedout", "context timed out or deadline exceeded")
)

const (
	defaultHardTimeout int64 = 20 * 1e3 // s -> ms
)

type result struct {
	response interface{}
	err      error
}

func endpointTimeoutMiddleware(huskarConfiger config.Configer, key ctxkeys.CtxKey) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			api := ctx.Value(key).(string)
			key := fmt.Sprintf(huskarkeys.HardTimeout, api)

			timeout := defaultHardTimeout
			if raw, err := huskarConfiger.Get(key); err == nil {
				if v, err := strconv.ParseInt(raw, 10, 64); err == nil {
					timeout = v
				}
			}

			resultCh := make(chan result, 1)
			ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Millisecond)
			defer cancel()

			go func() {
				defer func() {
					if e := recover(); e != nil {
						err := fmt.Errorf("panic: %v: %v", e, string(debug.Stack()))
						resultCh <- result{response: nil, err: err}
					}
				}()
				response, err := next(ctx, request)
				resultCh <- result{response: response, err: err}
			}()

			var res result
			select {
			case <-ctx.Done():
				return nil, ErrTimeout
			case res = <-resultCh:
			}

			if res.err != nil && (res.err == context.Canceled || res.err == context.DeadlineExceeded) {
				res.err = ErrTimeout
			}
			return res.response, res.err
		}
	}
}

// EndpointTimeoutSOAServerMiddleware is used for setting the global timeout for RPC handler.
func EndpointTimeoutSOAServerMiddleware(huskarConfiger config.Configer) endpoint.Middleware {
	return endpointTimeoutMiddleware(huskarConfiger, ctxkeys.APIName)
}

// EndpointTimeoutSOAClientMiddleware is used for setting the global timeout for RPC call.
func EndpointTimeoutSOAClientMiddleware(huskarConfiger config.Configer) endpoint.Middleware {
	return endpointTimeoutMiddleware(huskarConfiger, ctxkeys.CliAPIName)
}

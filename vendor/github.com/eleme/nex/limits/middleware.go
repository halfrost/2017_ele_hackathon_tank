package limits

import (
	"context"

	"github.com/eleme/nex/app"
	"github.com/eleme/nex/endpoint"
)

var (
	// ErrLimitExceeded is returned when max requests in progress exceeded.
	ErrLimitExceeded = app.NewUnknownException("LimitExceeded", "max requests in progress exceeded")
)

// EndpointMaxRequestsInProgressMiddleware limits the max in progress requests.
func EndpointMaxRequestsInProgressMiddleware(limitsCh chan struct{}) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (interface{}, error) {
			// If client timed out very quickly and reestablish a new connection,
			// timing will be a problem.
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case limitsCh <- struct{}{}:
				defer func() { <-limitsCh }()
			default:
				return nil, ErrLimitExceeded
			}
			return next(ctx, request)
		}
	}
}

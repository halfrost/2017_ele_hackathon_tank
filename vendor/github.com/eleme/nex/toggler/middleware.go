package toggler

import (
	"context"

	"github.com/eleme/huskar/toggle"
	"github.com/eleme/nex/app"
	"github.com/eleme/nex/consts/ctxkeys"
	"github.com/eleme/nex/endpoint"
)

// ErrAPINotSwitchedOn is the exception to return if api is switched off
var ErrAPINotSwitchedOn = app.NewUnknownException("APIDown", "api is shutdown")

// EndpointHuskarTogglerSOAServerMiddleware is used for soa server side.
func EndpointHuskarTogglerSOAServerMiddleware(huskarToggler toggle.Toggler) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			api := ctx.Value(ctxkeys.APIName).(string)

			if !huskarToggler.IsOnOr(api, true) {
				return nil, ErrAPINotSwitchedOn
			}

			return next(ctx, request)
		}
	}
}

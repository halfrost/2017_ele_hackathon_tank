package metric

import (
	"context"
	"reflect"
	"time"

	"github.com/eleme/nex/circuitbreaker"
	"github.com/eleme/nex/consts/ctxkeys"
	"github.com/eleme/nex/consts/names"
	"github.com/eleme/nex/endpoint"
	"github.com/eleme/nex/timeout"
	"github.com/eleme/statslib"
)

// All metrics must be compatible with the design:
//  https://t.elenet.me/zeus_core_doc/design/metrics.html

// endpointStatsdSOAMiddleware is a middleware which send soa-related metrics to statsd.
// prefix must be one of "soa", "soa_client".
func endpointStatsdSOAMiddleware(args *endpoint.SOAMiddlewareArgs, prefix string) endpoint.Middleware {
	prefixes := []string{prefix, args.AppID}

	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				var api string
				var keys []string
				if prefix == "soa" {
					api = ctx.Value(ctxkeys.APIName).(string)
					keys = append(keys, api)
				} else if prefix == "soa_client" {
					api = ctx.Value(ctxkeys.CliAPIName).(string)
					keys = append(prefixes, args.ThriftServiceName, api)
				}

				statsd.MeasureSinceKeysNoHost(keys, begin)
				statsd.IncrCountersNoHost(keys, 1)
				if err == nil {
				} else if err == timeout.ErrTimeout {
					statsd.IncrCountersNoHost(append(keys, names.TimeoutErr), 1)
				} else if err == circuitbreaker.ErrAPINotHealthy {
					statsd.IncrCountersNoHost(append(keys, names.NotHealthyErr), 1)
				} else {
					switch reflect.TypeOf(err) {
					case args.ErrTypes.UserErr:
						statsd.IncrCountersNoHost(append(keys, names.UserErr), 1)
					case args.ErrTypes.SysErr:
						statsd.IncrCountersNoHost(append(keys, names.SysErr), 1)
					default:
						statsd.IncrCountersNoHost(append(keys, names.UnkwnErr), 1)
					}
				}
			}(time.Now())
			return next(ctx, request)
		}
	}
}

// EndpointStatsdSOAServerMiddleware is used for soa server side endpoint,
// which has statsd schema like this: soa.<team>.<service>.api.xxx
func EndpointStatsdSOAServerMiddleware(args *endpoint.SOAMiddlewareArgs) endpoint.Middleware {
	return endpointStatsdSOAMiddleware(args, "soa")
}

// EndpointStatsdSOAClientMiddleware is used for soa client side endpoint,
// which has statsd schema like this: soa_client.<team>.<service>.api.xxx
func EndpointStatsdSOAClientMiddleware(args *endpoint.SOAMiddlewareArgs) endpoint.Middleware {
	return endpointStatsdSOAMiddleware(args, "soa_client")
}

// EndpointStatsdCommonMiddleware is used for thrid part libs,
// which has statsd schema like this: <lib-name>.<team>.<service>.api.xxx
func EndpointStatsdCommonMiddleware(appName, prefix string) endpoint.Middleware {
	prefixes := []string{prefix, appName}

	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				api := ctx.Value(ctxkeys.OthAPIName).(string)
				keys := append(prefixes, api)
				statsd.MeasureSinceKeysNoHost(keys, begin)
				statsd.IncrCountersNoHost(keys, 1)
				if err != nil {
					statsd.IncrCountersNoHost(append(keys, names.OthErr), 1)
				}
			}(time.Now())
			return next(ctx, request)
		}
	}
}

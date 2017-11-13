package log

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/eleme/nex/app"
	"github.com/eleme/nex/circuitbreaker"
	"github.com/eleme/nex/consts/ctxkeys"
	"github.com/eleme/nex/consts/names"
	"github.com/eleme/nex/endpoint"
	"github.com/eleme/nex/timeout"
	"github.com/eleme/nex/utils"
)

// ALogStringer is a abstraction to format Thrift arguments.
type ALogStringer interface {
	ALogString() string
}

// EndpointLoggingSOAServerMiddleware is a middleware which do logging works in server side.
func EndpointLoggingSOAServerMiddleware(logger RPCContextLogger, args *endpoint.SOAMiddlewareArgs) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				api := ctx.Value(ctxkeys.APIName).(string)

				if err == nil {
					logger.ContextInfof(ctx, "%v(%v) response time: %v", api, request, time.Since(begin))
				} else if err == timeout.ErrTimeout {
					logger.ContextErrorf(ctx, "%v(%v) -> %v: %v", api, request, names.TimeoutErr, utils.MarshalThriftError(err))
				} else if err == circuitbreaker.ErrAPINotHealthy {
					logger.ContextErrorf(ctx, "%v(%v) -> %v: %v", api, request, names.NotHealthyErr, utils.MarshalThriftError(err))
				} else {
					switch reflect.TypeOf(err) {
					case args.ErrTypes.UserErr:
						logger.ContextWarnf(ctx, "%v(%v) -> %v: %v", api, request, names.UserErr, utils.MarshalThriftError(err))
					case args.ErrTypes.SysErr:
						logger.ContextErrorf(ctx, "%v(%v) -> %v: %v", api, request, names.SysErr, utils.MarshalThriftError(err))
					case args.ErrTypes.UnkwnErr:
						logger.ContextErrorf(ctx, "%v(%v) -> %v: %v", api, request, names.UnkwnErr, utils.MarshalThriftError(err))
					default:
						logger.ContextErrorf(ctx, "%v(%v) -> %v: %v", api, request, names.UnkwnErr, utils.MarshalThriftError(err))
					}
				}

				// convert UnknownException into thrift exception
				if rerr, ok := err.(*app.UnknownException); ok {
					if e := app.ThriftUnknownExceptionFrom(rerr, args.ErrTypes.UnkwnErr); e != nil {
						err = e
					}
				}
			}(time.Now())
			return next(ctx, request)
		}
	}
}

// EndpointLoggingSOAClientMiddleware is a middleware which do logging works in client side.
func EndpointLoggingSOAClientMiddleware(logger RPCContextLogger, args *endpoint.SOAMiddlewareArgs) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func() {
				if err == nil { // Ignore unnecessary log
					return
				}

				api := ctx.Value(ctxkeys.CliAPIName).(string)
				api = fmt.Sprintf("%s.%s", args.ThriftServiceName, api)
				if err == timeout.ErrTimeout {
					logger.ContextErrorf(ctx, "remote call %v -> %v: %v", api, names.TimeoutErr, utils.MarshalThriftError(err))
				} else if err == circuitbreaker.ErrAPINotHealthy {
					logger.ContextErrorf(ctx, "remote call %v -> %v: %v", api, names.NotHealthyErr, utils.MarshalThriftError(err))
				} else {
					switch reflect.TypeOf(err) {
					case args.ErrTypes.UserErr:
						logger.ContextWarnf(ctx, "remote call %v -> %v: %v", api, names.UserErr, utils.MarshalThriftError(err))
					case args.ErrTypes.SysErr:
						logger.ContextErrorf(ctx, "remote call %v -> %v: %v", api, names.SysErr, utils.MarshalThriftError(err))
					case args.ErrTypes.UnkwnErr:
						logger.ContextErrorf(ctx, "remote call %v -> %v: %v", api, names.UnkwnErr, utils.MarshalThriftError(err))
					default:
						logger.ContextErrorf(ctx, "remote call %v -> %v: %v", api, names.UnkwnErr, err)
					}
				}
			}()
			return next(ctx, request)
		}
	}
}

// EndpointLoggingCommonMiddleware is a middleware which do logging works for other libs.
func EndpointLoggingCommonMiddleware(logger RPCContextLogger) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			defer func(begin time.Time) {
				api := ctx.Value(ctxkeys.OthAPIName).(string)

				if err == nil {
					logger.ContextInfof(ctx, "call %v took: %v", api, time.Since(begin))
				} else {
					logger.ContextErrorf(ctx, "call %v -> %v: %v", api, names.OthErr, err)
				}
			}(time.Now())
			return next(ctx, request)
		}
	}
}

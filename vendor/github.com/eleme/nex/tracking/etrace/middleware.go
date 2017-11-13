package etrace

import (
	"context"
	"fmt"
	"reflect"
	"strings"
	"time"

	etrace "github.com/eleme/etrace-go"
	"github.com/eleme/nex/circuitbreaker"
	"github.com/eleme/nex/consts/ctxkeys"
	"github.com/eleme/nex/endpoint"
	"github.com/eleme/nex/timeout"
	"github.com/eleme/nex/utils"
	tracker "github.com/eleme/thrift-tracker"
)

const (
	typeSuccess       = "success"
	typeUsrErr        = "UserException"
	typeSysErr        = "SystemException"
	typeUnkwnErr      = "UnknownException"
	typeTimeoutErr    = "TimeoutError"
	typeNotHealthyErr = "NotHealthyError"
)

// EndpointEtraceSOAServerMiddleware is a middleware which tracing soa service.
func EndpointEtraceSOAServerMiddleware(trace *Trace, args *endpoint.SOAMiddlewareArgs) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			apiName := ctx.Value(ctxkeys.APIName).(string)

			var root etrace.Transactioner
			service := fmt.Sprintf("%s.%s", args.ThriftServiceName, apiName)
			root, ctx = trace.NewTransaction(ctx, etrace.TypeService, service)
			root.AddTag(etrace.TagServiceClientApp, root.GetClientAppID())
			root.AddTag(etrace.TagServiceClientIP, args.RemoteAddr)
			ctx = context.WithValue(ctx, ctxkeys.EtraceTransactioner, root)

			defer func() {
				tagShadingKey(ctx, root)
				tagRPCResult(root, err, args.ErrTypes)
				root.Commit()
			}()
			return next(ctx, request)
		}
	}
}

// EndpointEtraceSOAClientMiddleware is a middleware which tracing soa client calls.
func EndpointEtraceSOAClientMiddleware(trace *Trace, args *endpoint.SOAMiddlewareArgs) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			apiName := ctx.Value(ctxkeys.CliAPIName).(string)

			var trans etrace.Transactioner
			service := fmt.Sprintf("%s.%v", args.ThriftServiceName, apiName)
			if root, ok := ctx.Value(ctxkeys.EtraceTransactioner).(etrace.Transactioner); ok && root != nil {
				trans = root.Fork(etrace.TypeCall, service)
			} else {
				trans, ctx = trace.NewRPCTransaction(ctx, etrace.TypeCall, service)
			}
			rpcID := trans.GetCurrentRPCID()
			ctx = context.WithValue(ctx, tracker.CtxKeySequenceID, rpcID)

			defer func() {
				remoteAddr := ctx.Value(ctxkeys.RemoteAddr).(string)
				trans.AddTag(etrace.TagCallServiceApp, args.AppID)
				trans.AddTag(etrace.TagCallServiceIP, remoteAddr)
				tagShadingKey(ctx, trans)
				tagRPCResult(trans, err, args.ErrTypes)

				evt := trans.CreateEvent("ETraceLink", "RemoteCall")
				evt.SetData(rpcID)
				evt.Commit()
				trans.Commit()
			}()
			return next(ctx, request)
		}
	}
}

func tagRPCResult(trans etrace.Transactioner, err error, errTypes *endpoint.ErrTypes) {
	if err == nil {
		trans.AddTag(etrace.TagServiceResult, typeSuccess)
		trans.SetStatus(etrace.StatusSuccess)
		return
	}

	var typ, status, msg string
	if err == timeout.ErrTimeout {
		typ = typeTimeoutErr
		status = typeTimeoutErr
		msg = utils.MarshalThriftError(err)
	} else if err == circuitbreaker.ErrAPINotHealthy {
		typ = typeNotHealthyErr
		status = typeNotHealthyErr
		msg = utils.MarshalThriftError(err)
	} else {
		switch reflect.TypeOf(err) {
		case errTypes.UserErr:
			typ = typeSuccess
			status = etrace.StatusSuccess
		case errTypes.SysErr:
			typ = typeSysErr
			status = typeSysErr
			msg = utils.MarshalThriftError(err)
		case errTypes.UnkwnErr:
			typ = typeUnkwnErr
			status = typeUnkwnErr
			msg = utils.MarshalThriftError(err)
		default:
			typ = typeUnkwnErr
			status = typeUnkwnErr
			msg = err.Error()
		}
	}

	if status != etrace.StatusSuccess {
		trans.LogError(typ, msg)
	}
	trans.AddTag(etrace.TagServiceResult, typ)
	trans.SetStatus(status)
}

func tagShadingKey(ctx context.Context, trans etrace.Transactioner) {
	if meta, ok := ctx.Value(tracker.CtxKeyRequestMeta).(map[string]string); ok {
		if rk, in := meta["routing-key"]; in {
			trans.AddTag(etrace.TagShadingKey, rk)
		}
	}
}

// EndpointEtraceDBMiddleware is a middleware which tracing SQL events.
func EndpointEtraceDBMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		root, ok := ctx.Value(ctxkeys.EtraceTransactioner).(etrace.Transactioner)
		if !ok || root == nil {
			return next(ctx, request)
		}

		apiName := ctx.Value(ctxkeys.OthAPIName).(string)
		if apiName == "PING" || apiName == "BEGIN" || apiName == "ROLLBACK" {
			return next(ctx, request)
		} else if apiName == "COMMIT" {
			apiName = "unknown.commit"
		} else {
			apiName = parsesSimpleSQL(apiName)
		}
		trans := root.Fork(etrace.TypeSQL, apiName)
		ctx = context.WithValue(ctx, tracker.CtxKeySequenceID, trans.GetCurrentRPCID())
		tagShadingKey(ctx, trans)

		response, err = next(ctx, request)
		if err == nil {
			trans.SetStatus(etrace.StatusSuccess)
		} else {
			trans.SetStatus(err.Error())
		}
		trans.Commit()
		return response, err
	}
}

// EndpointEtraceRedisMiddleware is a middleware which tracing Redis operations.
func EndpointEtraceRedisMiddleware(next endpoint.Endpoint) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		root, ok := ctx.Value(ctxkeys.EtraceTransactioner).(etrace.Transactioner)
		if !ok || root == nil {
			return next(ctx, request)
		}

		apiName := ctx.Value(ctxkeys.OthAPIName).(string)
		trans := root.Fork(etrace.TypeRedis, "Stats")

		if rtrans, ok := trans.(*redisTransaction); ok {
			ctx = context.WithValue(ctx, tracker.CtxKeySequenceID, rtrans.GetCurrentRPCID())
			defer func(begin time.Time) {
				parts := strings.Split(apiName, "@")
				rtrans.Merge(parts[0], parts[1], (err == nil), time.Since(begin))
			}(time.Now())
		}
		return next(ctx, request)
	}
}

// EndpointEtraceMQPublisherMiddleware is a middleware which tracing MQ producer operations.
func EndpointEtraceMQPublisherMiddleware(trace *Trace) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			root, ok := ctx.Value(ctxkeys.EtraceTransactioner).(etrace.Transactioner)
			if !ok || root == nil {
				return next(ctx, request)
			}
			trans := root.Fork(etrace.TypeRMQProduce, "Q.Publish")

			apiName := ctx.Value(ctxkeys.OthAPIName).(string)
			parts := strings.SplitN(apiName, " ", 3)
			server, exchange, routingKey := parts[0], parts[1], parts[2]
			sendtime := fmt.Sprintf("%v", time.Now().UnixNano()/1e6)
			rpcpid := fmt.Sprintf("%s^%s", trans.GetUpstreamRPCID(), sendtime)
			trans.AddTag("server", server)
			trans.AddTag("exchange", exchange)
			trans.AddTag("routing", routingKey)
			trans.AddTag("rpcid", rpcpid)
			tagShadingKey(ctx, trans)
			ctx = context.WithValue(ctx, ctxkeys.EtraceInfo, map[string]interface{}{
				"server":   server,
				"exchange": exchange,
				"routing":  routingKey,
				"rpcid":    rpcpid,
				"sendtime": sendtime,
				"rid":      trans.GetRequestID(),
				"clientid": trace.AppName(),
			})

			defer func() {
				if err == nil {
					trans.SetStatus(etrace.StatusSuccess)
				} else {
					trans.SetStatus(err.Error())
				}
				trans.Commit()
			}()
			return next(ctx, request)
		}
	}
}

// EndpointEtraceMQConsumerMiddleware is a middleware which tracing MQ consumer operations.
func EndpointEtraceMQConsumerMiddleware(trace *Trace, op string) endpoint.Middleware {
	return func(next endpoint.Endpoint) endpoint.Endpoint {
		return func(ctx context.Context, request interface{}) (response interface{}, err error) {
			traceInfo := ctx.Value(ctxkeys.EtraceInfo).(map[string]interface{})

			var trans etrace.Transactioner
			name := fmt.Sprintf("Q.%s", op)
			if root, ok := ctx.Value(ctxkeys.EtraceTransactioner).(etrace.Transactioner); ok && root != nil {
				trans = root.Fork(etrace.TypeRMQConsume, name)
			} else {
				trans, ctx = trace.NewRPCTransaction(ctx, etrace.TypeRMQConsume, name)
				ctx = context.WithValue(ctx, ctxkeys.EtraceTransactioner, trans)
			}
			for k, v := range traceInfo {
				trans.AddTag(k, v.(string))
			}
			tagShadingKey(ctx, trans)

			defer func() {
				if err == nil {
					trans.SetStatus(etrace.StatusSuccess)
				} else {
					trans.SetStatus(err.Error())
				}
				trans.Commit()
			}()
			return next(ctx, request)
		}
	}
}

package etrace

import (
	"context"
	"sync"

	"github.com/damnever/cc"
	"github.com/eleme/etrace-go"
	"github.com/eleme/nex/consts/ctxkeys"
	tracker "github.com/eleme/thrift-tracker"
)

// Trace is wrapper for etrace.Trace.
type Trace struct {
	appName string
	*etrace.Trace
}

// New creates a new Trace.
func New(nexCfg cc.Configer) (*Trace, error) {
	appName := nexCfg.String("app_name")
	trace, err := etrace.New(etrace.Config{
		AppID:           appName,
		Cluster:         nexCfg.String("cluster"),
		EZone:           nexCfg.String("ezone"),
		IDC:             nexCfg.String("idc"),
		EtraceConfigURL: nexCfg.String("etrace_uri"),
	})
	if err != nil {
		return nil, err
	}
	return &Trace{
		appName: appName,
		Trace:   trace,
	}, nil
}

// NewTransaction creates a new etrace.Transactioner with context, it extract request id and rpc id from context if posssible.
func (t *Trace) NewTransaction(ctx context.Context, typ, name string) (etrace.Transactioner, context.Context) {
	return t.newTransaction(ctx, typ, name, false)
}

// NewRPCTransaction creates a new etrace.Transactioner with context, it is used for RPC,
// and it extract request id and rpc id from context if posssible.
func (t *Trace) NewRPCTransaction(ctx context.Context, typ, name string) (etrace.Transactioner, context.Context) {
	return t.newTransaction(ctx, typ, name, true)
}

func (t *Trace) newTransaction(ctx context.Context, typ, name string, isRPC bool) (etrace.Transactioner, context.Context) {
	var reqID, rpcID string
	reqID, ctx = t.requestID(ctx)
	rpcID, ctx = t.rpcID(ctx, isRPC)
	cs := t.Trace.NewTransaction(reqID, rpcID, typ, name)
	return newCallStack(cs.(*etrace.CallStack)), ctx
}

// AppName returns the current app id.
func (t *Trace) AppName() string {
	return t.appName
}

func (t *Trace) requestID(ctx context.Context) (string, context.Context) {
	if reqID, ok := ctx.Value(tracker.CtxKeyRequestID).(string); ok {
		return reqID, ctx
	}
	reqID := etrace.CreateRequestID(t.appName)
	ctx = context.WithValue(ctx, tracker.CtxKeyRequestID, reqID)
	return reqID, ctx
}

func (t *Trace) rpcID(ctx context.Context, isRPC bool) (string, context.Context) {
	rpcID, ok := ctx.Value(tracker.CtxKeySequenceID).(string)
	if ok && !isRPC {
		return rpcID, ctx
	}
	rpcID, _ = etrace.CreateRPCID(t.appName, rpcID)
	ctx = context.WithValue(ctx, tracker.CtxKeySequenceID, rpcID)
	return rpcID, ctx
}

type callStack struct {
	sync.Mutex
	*etrace.CallStack
	redisTrans *redisTransaction
}

func newCallStack(cs *etrace.CallStack) *callStack {
	return &callStack{
		CallStack: cs,
	}
}

func (cs *callStack) Fork(typ, name string) etrace.Transactioner {
	trans := cs.CallStack.Fork(typ, name)
	if typ == etrace.TypeRedis {
		cs.Lock()
		defer cs.Unlock()
		if cs.redisTrans == nil {
			cs.redisTrans = newRedisTransaction(trans)
		}
		return cs.redisTrans
	}
	return trans
}

func (cs *callStack) Commit() {
	cs.Lock()
	if cs.redisTrans != nil {
		// do not care about the `child` GoRoutines which using redis commands,
		// they better have done merging before handler return..
		duration := cs.redisTrans.TagStatsAndSum()
		cs.redisTrans.CommitWithDuration(duration)
		cs.redisTrans = nil
	}
	cs.Unlock()
	cs.CallStack.Commit()
}

// LogErrorContext log error event.
func LogErrorContext(ctx context.Context, typ, msg string) {
	trans, ok := ctx.Value(ctxkeys.EtraceTransactioner).(etrace.Transactioner)
	if !ok {
		return
	}
	trans.LogError(typ, msg)
}

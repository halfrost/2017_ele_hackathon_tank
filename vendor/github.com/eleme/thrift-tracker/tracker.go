package tracker

import (
	"context"
	"sync"

	"github.com/apache/thrift/lib/go/thrift"
	"github.com/eleme/thrift-tracker/tracking"
	"github.com/google/uuid"
)

type ctxKey string

const (
	CtxKeySequenceID  ctxKey = "__thrift_tracking_sequence_id"
	CtxKeyRequestID   ctxKey = "__thrift_tracking_request_id"
	CtxKeyRequestMeta ctxKey = "__thrift_tracking_request_meta"
	// CtxKeyResponseMeta           ctxKey = "__thrift_tracking_response_meta"
	TrackingAPIName string = "__thriftpy_tracing_method_name__v2"
)

type HandShaker interface {
	Negotiation(curSeqID int32, iprot, oprot thrift.TProtocol) error
	TryUpgrade(seqID int32, iprot, oprot thrift.TProtocol) (bool, thrift.TException)
	RequestHeaderSupported() bool
	// ResponseHeaderSupported() bool
}

type Tracker interface {
	HandShaker

	RequestSeqIDFromCtx(ctx context.Context) (string, string)
	TryReadRequestHeader(iprot thrift.TProtocol) (context.Context, error) // context will pass into service handler
	TryWriteRequestHeader(ctx context.Context, oprot thrift.TProtocol) error
	// TryReadResponseHeader(iprot thrift.TProtocol) error
	// TryWriteResponseHeader(ctx context.Context, oprot thrift.TProtocol) error
}

type NewTrackerFactoryFunc func(name string) func() Tracker

type SimpleTracker struct {
	mu       *sync.RWMutex
	upgraded bool
	name     string
}

func NewSimpleTrackerFactory(name string) func() Tracker {
	return func() Tracker {
		return NewSimpleTracker(name)
	}
}

func NewSimpleTracker(name string) Tracker {
	return &SimpleTracker{
		mu:       &sync.RWMutex{},
		upgraded: false,
		name:     name,
	}
}

func (t *SimpleTracker) Negotiation(curSeqID int32, iprot, oprot thrift.TProtocol) error {
	// send
	if err := oprot.WriteMessageBegin(TrackingAPIName, thrift.CALL, curSeqID); err != nil {
		return err
	}
	args := tracking.NewUpgradeArgs_()
	args.AppID = t.name
	if err := args.Write(oprot); err != nil {
		return err
	}
	if err := oprot.WriteMessageEnd(); err != nil {
		return err
	}
	if err := oprot.Flush(); err != nil {
		return err
	}

	// recv
	method, mTypeID, seqID, err := iprot.ReadMessageBegin()
	if err != nil {
		return err
	}
	if method != TrackingAPIName {
		return thrift.NewTApplicationException(thrift.WRONG_METHOD_NAME,
			"tracker negotiation failed: wrong method name")
	}
	if curSeqID != seqID {
		return thrift.NewTApplicationException(thrift.BAD_SEQUENCE_ID,
			"tracker negotiation failed: out of sequence response")
	}
	if mTypeID == thrift.EXCEPTION {
		err0 := thrift.NewTApplicationException(thrift.UNKNOWN_APPLICATION_EXCEPTION,
			"Unknown Exception")
		var err1 thrift.TApplicationException
		if err1, err = err0.Read(iprot); err != nil {
			return err
		}
		if err = iprot.ReadMessageEnd(); err != nil {
			return err
		}
		if err1.TypeId() == thrift.UNKNOWN_METHOD { // server does not support tracker, ignore
			return nil
		}
		return err1
	}
	if mTypeID != thrift.REPLY {
		return thrift.NewTApplicationException(thrift.INVALID_MESSAGE_TYPE_EXCEPTION,
			"tracker negotiation failed: invalid message type")
	}
	reply := tracking.NewUpgradeReply()
	if err := reply.Read(iprot); err != nil {
		return err
	}
	if err := iprot.ReadMessageEnd(); err != nil {
		return err
	}
	t.upgradeProtocol()
	return nil
}

func (t *SimpleTracker) TryUpgrade(seqID int32, iprot, oprot thrift.TProtocol) (bool, thrift.TException) {
	args := tracking.NewUpgradeArgs_()
	if err := args.Read(iprot); err != nil {
		iprot.ReadMessageEnd()
		x := thrift.NewTApplicationException(thrift.PROTOCOL_ERROR, err.Error())
		x.Write(oprot)
		oprot.WriteMessageEnd()
		oprot.Flush()
		return false, err
	}
	iprot.ReadMessageEnd()

	result := tracking.NewUpgradeReply()
	if err := oprot.WriteMessageBegin(TrackingAPIName, thrift.REPLY, seqID); err != nil {
		return false, err
	}
	if err := result.Write(oprot); err != nil {
		return false, err
	}
	if err := oprot.WriteMessageEnd(); err != nil {
		return false, err
	}
	if err := oprot.Flush(); err != nil {
		return false, err
	}
	t.upgradeProtocol()
	return true, nil
}

func (t *SimpleTracker) upgradeProtocol() {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.upgraded = true
}

func (t *SimpleTracker) RequestHeaderSupported() bool {
	t.mu.RLock()
	defer t.mu.RUnlock()
	return t.upgraded
}

func (t *SimpleTracker) RequestSeqIDFromCtx(ctx context.Context) (string, string) {
	var reqID, seqID string

	if v, ok := ctx.Value(CtxKeyRequestID).(string); ok {
		reqID = v
	} else {
		reqID = uuid.New().String()
	}

	if v, ok := ctx.Value(CtxKeySequenceID).(string); ok {
		seqID = v
	} else {
		seqID = "1"
	}

	return reqID, seqID
}

func (t *SimpleTracker) TryReadRequestHeader(iprot thrift.TProtocol) (context.Context, error) {
	if !t.RequestHeaderSupported() {
		return context.TODO(), nil
	}
	header := tracking.NewRequestHeader()
	if err := header.Read(iprot); err != nil {
		return context.TODO(), err
	}
	ctx := context.Background()
	ctx = context.WithValue(ctx, CtxKeyRequestID, header.GetRequestID())
	ctx = context.WithValue(ctx, CtxKeySequenceID, header.GetSeq())
	ctx = context.WithValue(ctx, CtxKeyRequestMeta, header.GetMeta())
	return ctx, nil
}

func (t *SimpleTracker) TryWriteRequestHeader(ctx context.Context, oprot thrift.TProtocol) error {
	if !t.RequestHeaderSupported() {
		return nil
	}
	header := tracking.NewRequestHeader()
	if meta, ok := ctx.Value(CtxKeyRequestMeta).(map[string]string); ok {
		header.Meta = make(map[string]string, len(meta))
		for k, v := range meta {
			header.Meta[k] = v
		}
	}
	header.RequestID, header.Seq = t.RequestSeqIDFromCtx(ctx)
	return header.Write(oprot)
}

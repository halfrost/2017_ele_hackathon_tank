package etrace

import (
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/mailru/easyjson/jwriter"
)

// Transaction is a remote call transaction.
type Transaction struct {
	sync.Mutex
	callStack   *CallStack
	parent      *Transaction
	id          int64
	traceID     *int64
	typ         string
	name        string
	status      string
	timestamp   time.Time
	isCompleted bool
	duration    time.Duration
	tags        map[string]string
	children    []interface{}
}

func newTransaction(stack *CallStack, parent *Transaction, traceID *int64, typ, name string) *Transaction {
	if traceID == nil {
		var id int64
		traceID = &id
	}
	id := atomic.AddInt64(traceID, 1)
	return &Transaction{
		callStack: stack,
		parent:    parent,
		id:        id,
		traceID:   traceID,
		typ:       typ,
		name:      name,
		timestamp: time.Now(),
		status:    "0",
	}
}

// GetClientAppID returns current client app id .
func (t *Transaction) GetClientAppID() string {
	return t.callStack.clientAppID
}

// GetRequestID returns current request id.
func (t *Transaction) GetRequestID() string {
	return t.callStack.requestID
}

// GetUpstreamRPCID returns the upstream rpc id.
func (t *Transaction) GetUpstreamRPCID() string {
	parts := strings.Split(t.callStack.rpcID, "|")
	return parts[len(parts)-1]
}

// GetCurrentRPCID returns current rpc id.
func (t *Transaction) GetCurrentRPCID() string {
	return fmt.Sprintf("%s.%d", t.GetUpstreamRPCID(), t.id)
}

// GetCurrentRPCIDWithAppID returns current rpc id.
func (t *Transaction) GetCurrentRPCIDWithAppID() string {
	lastRPCID := t.callStack.rpcID
	return fmt.Sprintf("%s.%d", lastRPCID, t.id)
}

// CreateEvent add an event in current transaction.
func (t *Transaction) CreateEvent(typ, name string) *Event {
	return newEvent(t, t.traceID, typ, name)
}

// Fork create a child transaction in current transaction.
func (t *Transaction) Fork(typ, name string) Transactioner {
	return newTransaction(t.callStack, t, t.traceID, typ, name)
}

// AddTag add tags for current transaction.
func (t *Transaction) AddTag(key, val string) {
	t.Lock()
	defer t.Unlock()
	if t.isCompleted {
		return
	}
	if t.tags == nil {
		t.tags = make(map[string]string)
	}
	t.tags[key] = val

}

// LogError tries to catch exception.
func (t *Transaction) LogError(typ string, msg string) {
	evt := t.CreateEvent(TypeException, typ)
	evt.SetData(msg)
	evt.SetStatus("ERROR")
	evt.Commit()
}

// SetStatus set current transaction status,default status is "0".
func (t *Transaction) SetStatus(status string) {
	t.Lock()
	defer t.Unlock()
	if t.isCompleted {
		return
	}
	t.status = status
}

// Commit commit current transaction.
func (t *Transaction) Commit() {
	t.CommitWithDuration(time.Now().Sub(t.timestamp))
}

// CommitWithDuration commit current transaction with given duration.
func (t *Transaction) CommitWithDuration(duration time.Duration) {
	t.Lock()
	if t.isCompleted {
		t.Unlock()
		return
	}
	t.duration = duration
	t.isCompleted = true
	t.Unlock()
	if t.parent != nil {
		t.parent.commitChild(t)
	}
}

func (t *Transaction) commitChild(val interface{}) {
	t.Lock()
	defer t.Unlock()
	if t.isCompleted {
		return
	}
	t.children = append(t.children, val)
}

// MarshalJSON returns bytes based on etrace protocol.
func (t *Transaction) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	t.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

// MarshalEasyJSON appends bytes to the writer.
func (t *Transaction) MarshalEasyJSON(w *jwriter.Writer) {
	t.Lock()
	defer t.Unlock()
	w.RawString(`["transaction",`)
	w.String(t.typ)
	w.RawByte(',')
	w.String(t.name)
	w.RawByte(',')
	w.String(t.status)
	w.RawByte(',')
	w.Int64(t.id)
	w.RawByte(',')
	w.Int64(t.timestamp.UnixNano() / 1e6)
	w.RawByte(',')
	w.Bool(t.isCompleted)
	w.RawByte(',')
	if len(t.tags) == 0 {
		w.RawString("null")
	} else {
		w.RawString("{")
		first := true
		for k, v := range t.tags {
			if !first {
				w.RawByte(',')
			}
			first = false
			w.String(k)
			w.RawByte(':')
			w.String(v)
		}
		w.RawString("}")
	}
	w.RawByte(',')
	w.Int(int(t.duration.Nanoseconds() / 1e6))
	w.RawByte(',')
	if len(t.children) == 0 {
		w.RawString("null")
	} else {
		w.RawString("[")
		first := true
		for _, child := range t.children {
			if !first {
				w.RawByte(',')
			}
			first = false
			if trans, ok := child.(*Transaction); ok {
				trans.MarshalEasyJSON(w)
			}
			if evt, ok := child.(*Event); ok {
				evt.MarshalEasyJSON(w)
			}
		}
		w.RawByte(']')
	}
	w.RawByte(']')
}

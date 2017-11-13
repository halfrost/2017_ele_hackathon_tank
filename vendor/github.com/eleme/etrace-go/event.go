package etrace

import (
	"sync"
	"sync/atomic"
	"time"

	"github.com/mailru/easyjson/jwriter"
)

// Event is an etrace event.
type Event struct {
	sync.Mutex
	parent      *Transaction
	id          int64
	typ         string
	name        string
	status      string
	timeStamp   time.Time
	isCompleted bool
	tags        map[string]string
	data        string
}

func newEvent(parent *Transaction, traceID *int64, typ, name string) *Event {
	id := atomic.LoadInt64(traceID)
	return &Event{
		parent:      parent,
		id:          id,
		typ:         typ,
		name:        name,
		timeStamp:   time.Now(),
		status:      "0",
		isCompleted: true,
	}
}

// AddTag add tag for event.
func (e *Event) AddTag(key, val string) {
	e.Lock()
	if e.tags == nil {
		e.tags = make(map[string]string)
	}
	e.tags[key] = val
	e.Unlock()
}

// SetData set data for event.
func (e *Event) SetData(data string) {
	e.Lock()
	e.data = data
	e.Unlock()
}

// SetStatus set status for event.
func (e *Event) SetStatus(status string) {
	e.Lock()
	e.status = status
	e.Unlock()
}

// Commit commit current event.
func (e *Event) Commit() {
	if e.parent != nil {
		e.parent.commitChild(e)
	}
}

// MarshalJSON returns bytes based on etrace protocol.
func (e *Event) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	e.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

// MarshalEasyJSON appends bytes to the writer.
func (e *Event) MarshalEasyJSON(w *jwriter.Writer) {
	e.Lock()
	defer e.Unlock()
	w.RawString(`["event"`)
	w.RawByte(',')
	w.String(e.typ)
	w.RawByte(',')
	w.String(e.name)
	w.RawByte(',')
	w.String(e.status)
	w.RawByte(',')
	w.Int64(e.id)
	w.RawByte(',')
	w.Int64(e.timeStamp.UnixNano() / 1e6)
	w.RawByte(',')
	w.Bool(e.isCompleted)
	w.RawByte(',')
	if len(e.tags) == 0 {
		w.RawString("null")
	} else {
		w.RawString("{")
		childFirst := true
		for k, v := range e.tags {
			if !childFirst {
				w.RawByte(',')
			}
			childFirst = false
			w.String(k)
			w.RawByte(':')
			w.String(v)
		}
		w.RawString("}")
	}
	w.RawByte(',')
	w.String(e.data)
	w.RawByte(']')
}

package metric

import (
	"sync/atomic"

	"github.com/eleme/etrace-go/config"
	"github.com/mailru/easyjson/jwriter"
)

// Counter log basic integer.
type Counter struct {
	baseMetric
	value int
}

// NewCounter needs a commiter and a name.
func newCounter(cfg *config.Config, cm Commiter, name string) *Counter {
	return &Counter{
		baseMetric: newBaseMetric(cfg, cm, TypeCounter, name),
	}
}

// Merge try to merge two counter with same name and same tags.
func (c *Counter) Merge(other Metric) {
	if cc, ok := other.(*Counter); ok {
		c.value = c.value + cc.value
	}
}

// AddTag add a tag for counter.
func (c *Counter) AddTag(key, val string) *Counter {
	c.baseMetric.AddTag(key, val)
	return c
}

// SetValue change counter value.
func (c *Counter) SetValue(val int) {
	if atomic.CompareAndSwapInt32(&c.completed, 0, 1) {
		c.value = val
		c.cm.Commit(c)
	}
}

// MarshalJSON encode counter into json format.
func (c *Counter) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	c.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

// MarshalEasyJSON encode counter using easy json writer.
func (c *Counter) MarshalEasyJSON(w *jwriter.Writer) {
	if atomic.LoadInt32(&c.completed) == 0 {
		return
	}
	w.RawString(`["counter",`)
	w.String(c.name)
	w.RawByte(',')
	w.Int64(c.timestamp.UnixNano() / 1e6)
	w.RawByte(',')
	c.tag.MarshalEasyJSON(w)
	w.RawByte(',')
	w.Int(c.value)
	w.RawByte(']')
}

package metric

import (
	"math"
	"sync/atomic"

	"github.com/eleme/etrace-go/config"
	"github.com/mailru/easyjson/jwriter"
)

// Payload for data payload stats.
type Payload struct {
	baseMetric
	min   int64
	max   int64
	count int64
	sum   int64
}

// NewPayload create a payload with a name.
func newPayload(cfg *config.Config, cm Commiter, name string) *Payload {
	return &Payload{
		baseMetric: newBaseMetric(cfg, cm, TypePayload, name),
		min:        math.MaxInt64,
		max:        math.MinInt64,
		count:      1,
		sum:        0,
	}
}

// Merge merge tow payload with same name and same tags.
func (g *Payload) Merge(other Metric) {
	if gg, ok := other.(*Payload); ok {
		g.sum = g.sum + gg.sum
		g.count = g.count + gg.count
		if g.max < gg.max {
			g.max = gg.max
		}
		if g.min > gg.min {
			g.min = gg.min
		}
	}
}

// AddTag add a tag for payload.
func (g *Payload) AddTag(key, val string) *Payload {
	g.baseMetric.AddTag(key, val)
	return g
}

// SetValue change payload value.
func (g *Payload) SetValue(val int) {
	if atomic.CompareAndSwapInt32(&g.completed, 0, 1) {
		g.sum = int64(val)
		g.max = int64(val)
		g.min = int64(val)
		g.cm.Commit(g)
	}
}

// MarshalJSON encode payload into json format.
func (g *Payload) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	g.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

// MarshalEasyJSON encode into json format with easy json writer.
func (g *Payload) MarshalEasyJSON(w *jwriter.Writer) {
	w.RawString(`["payload",`)
	w.String(g.name)
	w.RawByte(',')
	w.Int64(g.timestamp.UnixNano() / 1e6)
	w.RawByte(',')
	g.tag.MarshalEasyJSON(w)
	w.RawByte(',')
	w.Int64(g.sum)
	w.RawByte(',')
	w.Int64(g.count)
	w.RawByte(',')
	w.Int64(g.min)
	w.RawByte(',')
	w.Int64(g.max)
	w.RawByte(']')
}

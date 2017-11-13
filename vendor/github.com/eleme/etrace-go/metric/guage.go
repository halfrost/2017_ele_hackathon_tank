package metric

import (
	"sync/atomic"

	"github.com/eleme/etrace-go/config"
	"github.com/mailru/easyjson/jwriter"
)

// Gauge if for gauge data.
type Gauge struct {
	baseMetric
	value int
}

// NewGauge returns a gauge with name and commiter.
func newGauge(cfg *config.Config, cm Commiter, name string) *Gauge {
	return &Gauge{
		baseMetric: newBaseMetric(cfg, cm, TypeGauge, name),
	}
}

// Merge will merge two gauge with same name and same tags.
func (g *Gauge) Merge(other Metric) {
	if gg, ok := other.(*Gauge); ok {
		g.value = g.value + gg.value
	}
}

// AddTag adds a new tag for gauge.
func (g *Gauge) AddTag(key, val string) *Gauge {
	g.baseMetric.AddTag(key, val)
	return g
}

// SetValue sets value for gauge.
func (g *Gauge) SetValue(val int) {
	if atomic.CompareAndSwapInt32(&g.completed, 0, 1) {
		g.value = val
		g.cm.Commit(g)
	}
}

// MarshalJSON encode gauge into json.
func (g *Gauge) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	g.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

// MarshalEasyJSON encode gauge with easy json writer.
func (g *Gauge) MarshalEasyJSON(w *jwriter.Writer) {
	w.RawString(`["gauge",`)
	w.String(g.name)
	w.RawByte(',')
	w.Int64(g.timestamp.UnixNano() / 1e6)
	w.RawByte(',')
	g.tag.MarshalEasyJSON(w)
	w.RawByte(',')
	w.Int(g.value)
	w.RawByte(']')
}

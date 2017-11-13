package metric

import (
	"math"
	"sync/atomic"
	"time"

	"github.com/eleme/etrace-go/config"
	"github.com/mailru/easyjson/jwriter"
)

// Timer for counting lasted time.
type Timer struct {
	baseMetric
	min         int64
	max         int64
	sum         int64
	count       int64
	upperEnable bool
}

// NewTimer creates a timer.
func newTimer(cfg *config.Config, cm Commiter, name string) *Timer {
	return &Timer{
		baseMetric:  newBaseMetric(cfg, cm, TypeTimer, name),
		count:       1,
		min:         math.MaxInt64,
		max:         math.MinInt64,
		upperEnable: true,
	}
}

// Merge merge two timer with same name and same tags.
func (t *Timer) Merge(other Metric) {
	if tt, ok := other.(*Timer); ok {
		t.sum = t.sum + tt.sum
		t.count = t.count + tt.count
		if t.max < tt.max {
			t.max = tt.max
		}
		if t.min > tt.min {
			t.min = tt.min
		}
	}
}

// EnableUpper will enable or disable upper calcation.
func (t *Timer) EnableUpper(flag bool) {
	t.upperEnable = flag
}

// UpperEnabled check if timer is upper enabled.
func (t *Timer) UpperEnabled() bool {
	return t.upperEnable
}

// AddTag adds a tag for timer.
func (t *Timer) AddTag(key, val string) *Timer {
	t.baseMetric.AddTag(key, val)
	return t
}

// SetValue change timer value.
func (t *Timer) SetValue(val int) {
	if atomic.CompareAndSwapInt32(&t.completed, 0, 1) {
		t.sum = int64(val)
		t.max = int64(val)
		t.min = int64(val)
		t.cm.Commit(t)
	}
}

// End ends a timer.
func (t *Timer) End() {
	duration := time.Since(t.timestamp)
	t.SetValue(int(duration.Nanoseconds() / 1e6))
}

// MarshalJSON encode timer into json format.
func (t *Timer) MarshalJSON() ([]byte, error) {
	w := jwriter.Writer{}
	t.MarshalEasyJSON(&w)
	return w.BuildBytes()
}

// MarshalEasyJSON encode timer with easyjson writer.
func (t *Timer) MarshalEasyJSON(w *jwriter.Writer) {
	w.RawString(`["timer",`)
	w.String(t.name)
	w.RawByte(',')
	w.Int64(t.timestamp.UnixNano() / 1e6)
	w.RawByte(',')
	t.tag.MarshalEasyJSON(w)
	w.RawByte(',')
	w.Int64(t.sum)
	w.RawByte(',')
	w.Int64(t.count)
	w.RawByte(',')
	w.Int64(t.min)
	w.RawByte(',')
	w.Int64(t.max)
	if t.upperEnable {
		w.RawString(",1]")
	} else {
		w.RawString(",0]")
	}
}
